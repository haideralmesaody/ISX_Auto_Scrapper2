function applyDefaultSort(arr) {
    return arr.sort((a, b) => {
        // 1) date descending
        const dA = new Date(a.date);
        const dB = new Date(b.date);
        if (dA.getTime() !== dB.getTime()) return dB - dA;

        // 2) symbol ascending
        if (a.symbol !== b.symbol) return a.symbol.localeCompare(b.symbol);

        // 3) change% descending
        return b.change - a.change;
    });
}

async function loadReport() {
    const container = document.getElementById('reportContent');
    if (container) container.innerHTML = '<p style="padding:1rem;">Loading daily report…</p>'; // placeholder

    try {
        const res = await fetch('/api/daily_report');
        if (!res.ok) {
            throw new Error(`Server responded ${res.status}`);
        }
        const data = await res.json();
        renderReport(data);
    } catch (err) {
        console.error('Failed to load daily report', err);
        if (container) container.innerHTML = `<p style="color:red; padding:1rem;">Failed to load daily report: ${err.message}</p>`;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    loadReport();

    const dl = document.getElementById('downloadReport');
    if (dl) dl.addEventListener('click', () => { window.location = '/api/daily_report_excel'; });

    setupRowInteractions();
});

function buildTopCard(title, rows, chartId, metric = 'volume') {
    return `<div class="report-card">
        <h4 class="card-title">${title}</h4>
        <div id="${chartId}" class="mini-chart"></div>
        ${buildTopTable(rows, metric)}
    </div>`;
}

function buildTopTable(rows, metric = 'volume') {
    // Determine the 5th column label and value extractor based on metric
    let colLabel = '';
    let cellRenderer = () => '';
    if (metric === 'volume') {
        colLabel = 'Volume';
        cellRenderer = r => fmtInt(r.volume);
    } else if (metric === 'value') {
        colLabel = 'Value';
        cellRenderer = r => fmtFloat(r.value);
    }

    // Build table header
    let headerHtml = '<tr><th>Ticker</th><th>Name</th><th>Close</th><th>Change%</th>';
    if (metric !== 'none') {
        headerHtml += `<th>${colLabel}</th>`;
    }
    headerHtml += '</tr>';

    let html = `<table class="simple-table interactive"><thead>${headerHtml}</thead><tbody>`;
    rows.forEach(r => {
        const changeClass = r.change_pct >= 0 ? 'positive' : 'negative';
        const changeSign = r.change_pct >= 0 ? '+' : '';
        html += `<tr data-ticker="${r.ticker}"><td>${r.ticker}</td><td>${r.name}</td><td>${fmtFloat(r.close)}</td><td class="${changeClass}">${changeSign}${fmtFloat(r.change_pct)}</td>`;
        if (metric !== 'none') {
            html += `<td>${cellRenderer(r)}</td>`;
        }
        html += '</tr>';
    });
    html += '</tbody></table>';
    return html;
}

function renderReport(data) {
    const container = document.getElementById('reportContent');
    if (!container) return;

    const topCardsHtml = `
        <section class="report-cards">
            ${buildTopCard('Top 5 by Volume', data.top_volume, 'chart-volume', 'volume')}
            ${buildTopCard('Top 5 by Value', data.top_value, 'chart-value', 'value')}
            ${buildTopCard('Top 5 Gainers', data.top_gain, 'chart-gainers', 'none')}
            ${buildTopCard('Top 5 Losers', data.top_loss, 'chart-losers', 'none')}
        </section>`;

    const tradedSorted = sortTraded([...data.traded]);
    const nonSorted = sortCompanies([...data.non_traded]);

    container.innerHTML = `
        <div class="report-header-section">
            <h2>Daily Report – ${data.date}</h2>
        </div>
        ${topCardsHtml}
        <section class="report-section">
            <h3>Traded Companies</h3>
            ${buildCompanyTable(tradedSorted)}
        </section>
        <section class="report-section">
            <h3>Non-Traded Companies</h3>
            ${buildNonTradedTable(nonSorted)}
        </section>`;

    createSparklines(tradedSorted);
    createSparklines(nonSorted);

    drawTopPie('chart-volume', 'Volume Share', data.top_volume, 'volume');
    drawTopPie('chart-value', 'Value Share', data.top_value, 'value');
    drawTopPie('chart-gainers', 'Gain %', data.top_gain, 'change_pct');
    drawTopPie('chart-losers', 'Loss %', data.top_loss, 'change_pct');
}

function buildCompanyTable(rows) {
    if (!Array.isArray(rows) || rows.length === 0) {
        return '<p>No data available.</p>';
    }
    let html = '<table class="simple-table interactive"><thead><tr><th>Code</th><th>Name</th><th>Open</th><th>High</th><th>Low</th><th>Avg</th><th>Prev Avg</th><th>Close</th><th>Prev Close</th><th>Change%</th><th>Trades</th><th>Volume</th><th>Value</th><th></th></tr></thead><tbody>';
    rows.forEach(r => {
        const cp = (r.change_pct ?? 0);
        const changeCls = cp >= 0 ? 'positive' : 'negative';
        const sign = cp >= 0 ? '+' : '';
        html += `<tr data-ticker="${r.code}"><td>${r.code}</td><td>${r.name}</td><td>${fmtFloat(r.open)}</td><td>${fmtFloat(r.high)}</td><td>${fmtFloat(r.low)}</td><td>${fmtFloat(r.avg_price)}</td><td>${fmtFloat(r.prev_avg_price)}</td><td>${fmtFloat(r.close)}</td><td>${fmtFloat(r.prev_close)}</td><td class="${changeCls}">${sign}${fmtFloat(cp)}</td><td>${fmtInt(r.trades)}</td><td>${fmtInt(r.volume)}</td><td>${fmtFloat(r.value)}</td><td><div class="sparkline" id="spark-${r.code}"></div></td></tr>`;
    });
    html += '</tbody></table>';
    return html;
}

function buildNonTradedTable(rows) {
    if (!Array.isArray(rows) || rows.length === 0) {
        return '<p>No data available.</p>';
    }
    let html = '<table class="simple-table interactive"><thead><tr><th>Code</th><th>Name</th><th>Last Traded</th><th>Open</th><th>High</th><th>Low</th><th>Avg</th><th>Close</th><th>Trades</th><th>Volume</th><th>Value</th><th></th></tr></thead><tbody>';
    rows.forEach(r => {
        html += `<tr data-ticker="${r.code}"><td>${r.code}</td><td>${r.name}</td><td>${r.last_traded || '-'}</td><td>${fmtFloat(r.open)}</td><td>${fmtFloat(r.high)}</td><td>${fmtFloat(r.low)}</td><td>${fmtFloat(r.avg_price)}</td><td>${fmtFloat(r.close)}</td><td>${fmtInt(r.trades)}</td><td>${fmtInt(r.volume)}</td><td>${fmtFloat(r.value)}</td><td><div class="sparkline" id="spark-${r.code}"></div></td></tr>`;
    });
    html += '</tbody></table>';
    return html;
}

// -------- Number formatting helpers --------
function fmtFloat(val) {
    if (val === null || val === undefined) return '-';
    return Number(val).toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}
function fmtInt(val) {
    if (val === null || val === undefined) return '-';
    return Number(val).toLocaleString('en-US');
}

// ---------------- Sparkline util (inline SVG) ----------------
function renderSparkline(containerId, data, isPositive) {
    const width = 70;
    const height = 30;
    if (!data || data.length === 0) {
        const el = document.getElementById(containerId);
        if (el) el.innerHTML = '';
        return;
    }
    const min = Math.min(...data);
    const max = Math.max(...data);
    const range = max - min || 1;
    const stepX = width / (data.length - 1);
    let d = '';
    data.forEach((v, i) => {
        const x = (i * stepX).toFixed(1);
        const y = (height - ((v - min) / range) * height).toFixed(1);
        d += (i === 0 ? 'M' : 'L') + x + ' ' + y + ' ';
    });

    let dotsSvg = '';
    data.forEach((v, i) => {
        const x = (i * stepX).toFixed(1);
        const y = (height - ((v - min) / range) * height).toFixed(1);
        const dotId = `${containerId}-dot-${i}`;
        dotsSvg += `<circle id="${dotId}" class="spark-dot" cx="${x}" cy="${y}" r="2" fill="${isPositive ? '#4ade80' : '#f87171'}" />`;
        dotsSvg += `<circle class="spark-hit" data-dot="${dotId}" cx="${x}" cy="${y}" r="8" fill="transparent"><title>${fmtFloat(v)}</title></circle>`;
    });

    const stroke = isPositive ? '#4ade80' : '#f87171';
    const svg = `<svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">
        <path d="${d.trim()}" stroke="${stroke}" stroke-width="1" fill="none" stroke-linejoin="round" stroke-linecap="round"/>
        ${dotsSvg}
    </svg>`;
    const target = document.getElementById(containerId);
    if (target) target.innerHTML = svg;
}

// ------------- Default company sort -------------
function sortCompanies(list) {
    return list.sort((a, b) => {
        const dA = new Date(a.last_traded || '1970-01-01');
        const dB = new Date(b.last_traded || '1970-01-01');
        if (dA.getTime() !== dB.getTime()) return dB - dA;
        if (a.code !== b.code) return a.code.localeCompare(b.code);
        return b.change_pct - a.change_pct;
    });
}

function createSparklines(list){
    list.forEach(r=>{
        if(r.sparkline && r.sparkline.length>1){
            const isPos = r.sparkline[r.sparkline.length-1] >= r.sparkline[0];
            renderSparkline(`spark-${r.code}`, r.sparkline, isPos);
        }
    });
}

// Sort for the traded companies: change % descending, then code A-Z
function sortTraded(list) {
    return list.sort((a, b) => {
        if (a.change_pct !== b.change_pct) return b.change_pct - a.change_pct;
        return a.code.localeCompare(b.code);
    });
}

// ----------- Pie chart rendering -----------
function drawTopPie(container, title, rows, metric){
    if (!window.Highcharts || !Array.isArray(rows) || rows.length===0) return;
    const seriesData = rows.map(r=>{
        let val = 0;
        if(metric==='change_pct'){
            val = Math.abs(Number(r.change_pct || 0));
        }else{
            val = Number(r[metric] ?? r.volume ?? 0);
        }
        return { name: r.ticker || r.code, y: val };
    }).filter(p=>p.y>0);
    const el = document.getElementById(container);
    if(!el) return;
    if(seriesData.length===0){
        el.innerHTML='';
        return;
    }
    Highcharts.chart(container, {
        chart: { type: 'pie', height: 220, backgroundColor: 'transparent' },
        title: { text: null },
        tooltip: { pointFormat: '<b>{point.y:.2f}</b>' },
        plotOptions: {
            pie: {
                allowPointSelect: false,
                cursor: 'pointer',
                dataLabels: { enabled: false }
            }
        },
        credits: { enabled:false },
        exporting: { enabled:false },
        stockTools: { gui: { enabled:false } },
        series: [{ name: title, colorByPoint: true, data: seriesData }]
    });
}

// ------------- Interactive row click handling -------------
function setupRowInteractions() {
    // Add hover styling via JS to guarantee presence
    const styleTag = document.createElement('style');
    styleTag.textContent = `.simple-table.interactive tbody tr:hover { background:#eef4ff; cursor:pointer; }`;
    document.head.appendChild(styleTag);

    document.body.addEventListener('click', async (e) => {
        const tr = e.target.closest('tr[data-ticker]');
        if (!tr) return;
        const ticker = tr.getAttribute('data-ticker');
        if (!ticker) return;
        await openChartModal(ticker);
    });
}

// --------- Modal + Chart logic ----------
let chartModalEl = null;
let chartContainerEl = null;
let currentDailyChart = null;

async function openChartModal(ticker) {
    if (!chartModalEl) {
        initChartModal();
    }
    chartModalEl.classList.add('show');
    await ensureHighstockLoaded();
    await createDailyCandleChart(ticker);
}

function initChartModal() {
    chartModalEl = document.createElement('div');
    chartModalEl.className = 'modal';
    chartModalEl.innerHTML = `
        <div class="modal-content" style="width:90%; max-width:1100px; height:80%; display:flex; flex-direction:column;">
            <span id="chartModalClose" style="align-self:flex-end; cursor:pointer; font-size:1.4rem;">&times;</span>
            <h3 id="chartModalTitle" style="margin-bottom:0.5rem;">Loading…</h3>
            <div id="dailyChartContainer" style="flex:1; min-height:500px;"></div>
        </div>`;
    document.body.appendChild(chartModalEl);
    chartContainerEl = document.getElementById('dailyChartContainer');
    document.getElementById('chartModalClose').addEventListener('click', () => {
        chartModalEl.classList.remove('show');
        if (currentDailyChart) {
            currentDailyChart.destroy();
            currentDailyChart = null;
        }
    });
    // clicking outside modal-content closes
    chartModalEl.addEventListener('click', (ev)=>{
        if(ev.target===chartModalEl){
            chartModalEl.classList.remove('show');
            if (currentDailyChart) { currentDailyChart.destroy(); currentDailyChart=null; }
        }
    });
}

async function ensureHighstockLoaded() {
    if (window.Highcharts && typeof Highcharts.stockChart === 'function') {
        return;
    }
    // load stock/highstock.js dynamically
    await loadScript('https://code.highcharts.com/stock/highstock.js');
    // Optional modules for consistency but not blocking if fail
    loadScript('https://code.highcharts.com/stock/modules/annotations-advanced.js');
    loadScript('https://code.highcharts.com/stock/modules/full-screen.js');
    loadScript('https://code.highcharts.com/stock/modules/price-indicator.js');
}

function loadScript(src) {
    return new Promise((resolve, reject) => {
        const existing = document.querySelector(`script[src="${src}"]`);
        if (existing) {
            existing.addEventListener('load', () => resolve());
            if (existing.complete) resolve();
            return;
        }
        const s = document.createElement('script');
        s.src = src;
        s.onload = () => resolve();
        s.onerror = () => reject(new Error('Failed to load script ' + src));
        document.head.appendChild(s);
    });
}

async function createDailyCandleChart(symbol) {
    if (!window.Highcharts || !Highcharts.stockChart) return;

    document.getElementById('chartModalTitle').textContent = `${symbol} – Price History`;
    chartContainerEl.innerHTML = '<p style="padding:1rem;">Loading chart…</p>';

    try {
        const res = await fetch(`/api/ticker/${symbol}?type=price`);
        if (!res.ok) throw new Error('Server responded ' + res.status);
        const data = await res.json();
        if(!Array.isArray(data) || data.length===0){ throw new Error('No data'); }

        const ohlcData = [];
        const volumeData = [];
        data.forEach(item=>{
            const ts = item.timestamp ? item.timestamp : Date.parse(item.date);
            ohlcData.push([ts, item.open, item.high, item.low, item.close]);
            volumeData.push([ts, item.volume || 0]);
        });

        if (currentDailyChart) {
            currentDailyChart.destroy();
            currentDailyChart = null;
        }

        currentDailyChart = Highcharts.stockChart('dailyChartContainer', {
            yAxis: [{
                labels: { align: 'right', x: -3 },
                title: { text: 'Price' },
                height: '70%',
                lineWidth: 1,
                resize: { enabled:true }
            }, {
                labels: { align: 'right', x: -3 },
                title: { text: 'Volume' },
                top: '75%',
                height: '25%',
                offset: 0,
                lineWidth: 1
            }],
            series: [{
                id: 'main',
                type: 'candlestick',
                name: symbol,
                color: '#FF6F6F',
                upColor: '#6FB76F',
                data: ohlcData,
                dataGrouping: { enabled:false }
            }, {
                type: 'column',
                name: 'Volume',
                id: 'volume',
                data: volumeData,
                yAxis: 1,
                color: 'rgba(0, 0, 150, 0.3)',
                dataGrouping: { enabled:false }
            }],
            stockTools: {
                gui: {
                    enabled: true,
                    buttons: ['indicators', 'separator', 'simpleShapes', 'lines', 'measure', 'advanced', 'separator', 'toggleAnnotations', 'verticalLabels', 'flags', 'separator', 'zoomChange', 'fullScreen', 'typeChange', 'separator', 'currentPriceIndicator']
                }
            },
            tooltip: { split: false },
            rangeSelector: { selected: 1 },
            navigator: { enabled: true },
            scrollbar: { enabled: true },
            title: { text: `${symbol} – Iraqi Stock Exchange` },
            credits: { enabled:false },
            chart: { backgroundColor: 'rgba(255,255,255,0.95)' }
        });
    } catch (err) {
        chartContainerEl.innerHTML = `<p style="color:red; padding:1rem;">Failed to load chart: ${err.message}</p>`;
    }
}
