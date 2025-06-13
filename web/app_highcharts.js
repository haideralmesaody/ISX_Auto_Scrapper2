// ISX Auto Scrapper - Simple Highcharts Implementation
let tickersData = [];
let displayedData = [];
let sortColumn = 'date';
let sortAsc = false;
let currentChart = null;
let selectedSymbol = '';
let selectedRow = null;

// Debug function
function debugLog(message) {
    console.log(message);
}

// Initialize the application
document.addEventListener('DOMContentLoaded', async function() {
    debugLog('=== ISX Auto Scrapper Dashboard Loading... ===');
    
    if (typeof Highcharts === 'undefined') {
        debugLog('ERROR: Highcharts not loaded');
    } else {
        debugLog('Highcharts version: ' + Highcharts.version);
    }
    
    // Tab navigation setup
    document.querySelectorAll('.tab-button').forEach(btn => {
        btn.addEventListener('click', () => switchTab(btn.dataset.tab));
    });

    switchTab('dashboardTab');

    // Initialize dashboard
    document.getElementById('fetchBtn').addEventListener('click', openFetchModal);
    document.getElementById('refreshBtn').addEventListener('click', runRefresh);
    document.getElementById('calcBtn').addEventListener('click', runCalculate);
    document.getElementById('calcNumBtn').addEventListener('click', runCalculateNum);
    document.getElementById('liqBtn').addEventListener('click', runLiquidity);
    document.getElementById('stratBtn').addEventListener('click', runStrategies);
    document.getElementById('backtestBtn').addEventListener('click', runBacktest);

    document.getElementById('fetchCancelBtn').addEventListener('click', closeFetchModal);
    document.getElementById('fetchConfirmBtn').addEventListener('click', confirmFetch);

    await initializeDashboard();
});

// Initialize dashboard
async function initializeDashboard() {
    debugLog('Initializing dashboard...');
    showLoading(true);
    
    try {
        // Load tickers
        const response = await fetch('/api/tickers');
        if (!response.ok) {
            throw new Error('Failed to load tickers');
        }
        
        tickersData = await response.json();
        displayedData = applyDefaultSort([...tickersData]);
        populateTickerTable(displayedData);
        
        document.getElementById('tickerSearch').addEventListener('input', filterTickers);
        
        showLoading(false);
        debugLog('Dashboard initialized successfully');
        
    } catch (error) {
        debugLog('Error initializing dashboard: ' + error.message);
        showLoading(false);
    }
}

// Helper formatters
function formatNumber(n) {
    if (n === undefined || n === null) return '-';
    if (n >= 1e9) return (n / 1e9).toFixed(1) + 'B';
    if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M';
    if (n >= 1e3) return (n / 1e3).toFixed(1) + 'K';
    return n.toLocaleString();
}

function formatPrice(n) {
    if (n === undefined || n === null) return '-';
    return n % 1 ? n.toFixed(2) : n.toFixed(0);
}

// -------------------------------------------------------------
// Lightweight sparkline generator using inline SVG (no library)
// -------------------------------------------------------------
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
    const range = max - min || 1; // protect against div-by-zero
    const stepX = width / (data.length - 1);

    let d = '';
    data.forEach((v, i) => {
        const x = (i * stepX).toFixed(1);
        const y = (height - ((v - min) / range) * height).toFixed(1);
        d += (i === 0 ? 'M' : 'L') + x + ' ' + y + ' ';
    });

    // Build visible dots and larger invisible hit-area circles
    let dotsSvg = '';
    data.forEach((v, i) => {
        const x = (i * stepX).toFixed(1);
        const y = (height - ((v - min) / range) * height).toFixed(1);
        const dotId = `${containerId}-dot-${i}`;
        // visible smaller dot
        dotsSvg += `<circle id="${dotId}" class="spark-dot" cx="${x}" cy="${y}" r="2" fill="${isPositive ? '#4ade80' : '#f87171'}" />`;
        // invisible bigger circle for hover / tooltip
        dotsSvg += `<circle class="spark-hit" data-dot="${dotId}" cx="${x}" cy="${y}" r="8" fill="transparent"><title>${v.toFixed(2)}</title></circle>`;
    });

    const stroke = isPositive ? '#4ade80' : '#f87171';
    const svg = `<svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg">
        <path d="${d.trim()}" stroke="${stroke}" stroke-width="1" fill="none" stroke-linejoin="round" stroke-linecap="round"/>
        ${dotsSvg}
    </svg>`;

    const target = document.getElementById(containerId);
    if (target) {
        target.innerHTML = svg;
        // attach hover events to enlarge dots
        const svgEl = target.querySelector('svg');
        if (svgEl) {
            svgEl.querySelectorAll('.spark-hit').forEach(hit => {
                const dotId = hit.getAttribute('data-dot');
                const dot = svgEl.querySelector(`#${CSS.escape(dotId)}`);
                if (dot) {
                    hit.addEventListener('mouseenter', () => dot.setAttribute('r', '4'));
                    hit.addEventListener('mouseleave', () => dot.setAttribute('r', '2'));
                }
            });
        }
    }
}

// Build ticker table with sparkline charts
function populateTickerTable(data = displayedData) {
    const table = document.getElementById('tickerTable');
    table.innerHTML = '';

    const thead = document.createElement('thead');
    thead.innerHTML = `
        <tr>
            <th data-col="symbol">Symbol</th>
            <th data-col="date">Date</th>
            <th data-col="price">Close</th>
            <th data-col="change">Chg%</th>
            <th></th>
        </tr>`;
    table.appendChild(thead);

    const tbody = document.createElement('tbody');
    table.appendChild(tbody);

    data.forEach(ticker => {
        const row = document.createElement('tr');
        row.classList.add('ticker-row');
        row.innerHTML = `
            <td>${ticker.symbol}</td>
            <td>${ticker.date || ''}</td>
            <td>${formatPrice(ticker.price)}</td>
            <td class="${ticker.change >= 0 ? 'positive' : 'negative'}">${ticker.change.toFixed(2)}</td>
            <td><div class="sparkline" id="spark-${ticker.symbol}"></div></td>`;
        row.addEventListener('click', () => {
            selectTicker(ticker.symbol);
            if (selectedRow) selectedRow.classList.remove('selected');
            row.classList.add('selected');
            selectedRow = row;
        });
        if (ticker.symbol === selectedSymbol) {
            row.classList.add('selected');
            selectedRow = row;
        }
        tbody.appendChild(row);

        if (ticker.sparkline && ticker.sparkline.length > 0) {
            renderSparkline(`spark-${ticker.symbol}`, ticker.sparkline, ticker.change >= 0);
        }
    });

    table.querySelectorAll('th[data-col]').forEach(th => {
        th.addEventListener('click', () => sortTickers(th.dataset.col));
    });
}

function sortTickers(column) {
    if (sortColumn === column) {
        sortAsc = !sortAsc;
    } else {
        sortColumn = column;
        sortAsc = true;
    }
    sortAndDisplay();
}

function filterTickers() {
    const q = document.getElementById('tickerSearch').value.trim().toUpperCase();
    displayedData = tickersData.filter(t => t.symbol.toUpperCase().startsWith(q));
    sortAndDisplay();
}

function sortAndDisplay() {
    displayedData = applyDefaultSort(displayedData);
    populateTickerTable(displayedData);
}

// Handle ticker selection
async function selectTicker(symbol) {
    selectedSymbol = symbol;
    debugLog('Selected ticker: ' + symbol);
    
    if (!symbol) {
        clearChart();
        return;
    }
    
    // Find ticker info
    const ticker = tickersData.find(t => t.symbol === symbol);
    if (!ticker) {
        debugLog('Ticker not found: ' + symbol);
        return;
    }
    
    // Update UI
    updateSelectedTickerInfo(ticker);
    document.getElementById('chartTitle').textContent = `${ticker.symbol} - ${ticker.name}`;
    
    // Create chart using the exact pattern from the Highcharts sample
    await createChart(symbol);

    // Fetch latest strategy signals
    try {
        const res = await fetch(`/api/ticker/${symbol}?type=strategies`);
        if (res.ok) {
            const data = await res.json();
            renderStrategySignals(data.signals);
        }
    } catch (err) {
        console.log('Strategy fetch error', err);
    }
}

// Update selected ticker information
function updateSelectedTickerInfo(ticker) {
    const infoDiv = document.getElementById('selectedTickerInfo');
    if (!ticker) {
        infoDiv.style.display = 'none';
        return;
    }
    
    const changeClass = ticker.change >= 0 ? 'positive' : 'negative';
    const changeSign = ticker.change >= 0 ? '+' : '';
    
    infoDiv.innerHTML = `
        <div class="ticker-details">
            <div>
                <div class="ticker-symbol">${ticker.symbol}</div>
                <div class="ticker-name">${ticker.name}</div>
            </div>
            <div>
                <div class="ticker-price">${ticker.price.toFixed(2)} IQD</div>
                <div class="ticker-change ${changeClass}">${changeSign}${ticker.change.toFixed(2)}%</div>
                <div class="ticker-ohlc">O:${ticker.open.toFixed(2)} H:${ticker.high.toFixed(2)} L:${ticker.low.toFixed(2)} V:${ticker.volume}</div>
            </div>
        </div>
    `;
    
    infoDiv.classList.add('show');
}

// Create chart - EXACTLY matching the Highcharts sample pattern
async function createChart(symbol) {
    debugLog('Creating chart for: ' + symbol);
    showLoading(true);
    
    try {
        // Clear existing chart
        if (currentChart) {
            currentChart.destroy();
            currentChart = null;
        }
        
        // Fetch price data
        const data = await fetch(`/api/ticker/${symbol}?type=price`).then(r=>r.json());
        debugLog('Loaded data points: ' + data.length);
        if(!Array.isArray(data) || data.length===0) throw new Error('No data');

        const ohlcData = [];
        const volumeData = [];
        data.forEach(item=>{
            const ts = item.timestamp ? item.timestamp : Date.parse(item.date);
            ohlcData.push([ts, item.open, item.high, item.low, item.close]);
            volumeData.push([ts, item.volume || 0]);
        });
        
        currentChart = Highcharts.stockChart('container', {
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
                color: '#FF6F6F',
                upColor: '#6FB76F',
                data: ohlcData,
                dataGrouping: { enabled:false }
            }, {
                id: 'volume',
                type: 'column',
                name: 'Volume',
                data: volumeData,
                yAxis: 1,
                color: 'rgba(0, 0, 150, 0.3)',
                dataGrouping: { enabled:false }
            }],
            stockTools: {
                gui: {
                    enabled: true,
                    buttons: ['indicators', 'separator', 'simpleShapes', 'lines', 'crookedLines', 'measure', 'advanced', 'toggleAnnotations', 'separator', 'verticalLabels', 'flags', 'separator', 'zoomChange', 'fullScreen', 'typeChange', 'separator', 'currentPriceIndicator']
                }
            },
            title: { text: `${symbol} - Iraqi Stock Exchange` },
            rangeSelector: { selected: 1 },
            navigator: { enabled:true },
            scrollbar: { enabled:true },
            credits: { enabled:false },
            colors: ['#2d5016', '#6b9b37', '#FF6F6F', '#8bc34a', '#4a7c23'],
            chart: { backgroundColor:'rgba(0,0,0,0)' }
        });

        debugLog('Chart created successfully');
        // Expand StockTools toolbar
        setTimeout(()=>{
            const wrapper=document.querySelector('.highcharts-stocktools-wrapper');
            if(wrapper){wrapper.setAttribute('aria-hidden','false');const btn=wrapper.querySelector('.highcharts-toggle-toolbar');if(btn&&btn.classList.contains('highcharts-arrow-left'))btn.click();}
        },100);

        showLoading(false);
    } catch(error) {
        debugLog('Error creating chart: ' + error.message);
        showLoading(false);
        document.getElementById('container').innerHTML = `<div style="text-align:center; padding:50px; color:#e74c3c;"><h3>Error loading chart</h3><p>${error.message}</p></div>`;
    }
}

// Clear chart
function clearChart() {
    if (currentChart) {
        currentChart.destroy();
        currentChart = null;
    }
    
    document.getElementById('container').innerHTML = '';
    document.getElementById('chartTitle').textContent = 'Select a ticker to view chart';
    document.getElementById('selectedTickerInfo').classList.remove('show');
}

// Show/hide loading overlay
function showLoading(show) {
    const overlay = document.getElementById('loadingOverlay');
    if (show) {
        overlay.classList.add('show');
    } else {
        overlay.classList.remove('show');
    }
}

async function runCalculate() {
    showLoading(true);
    try {
        const res = await fetch('/api/calculate', { method: 'POST' });
        const data = await res.json();
        debugLog('Calculate: ' + data.status);
    } catch (err) {
        debugLog('Calculate error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

async function runLiquidity() {
    showLoading(true);
    try {
        const res = await fetch('/api/liquidity', { method: 'POST' });
        const data = await res.json();
        debugLog('Liquidity: ' + data.status);
    } catch (err) {
        debugLog('Liquidity error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

async function runStrategies() {
    showLoading(true);
    try {
        const res = await fetch('/api/strategies', { method: 'POST' });
        const data = await res.json();
        debugLog('Strategies: ' + data.status);
    } catch (err) {
        debugLog('Strategies error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

function openFetchModal() {
    populateFetchDropdown();
    document.getElementById('fetchModal').classList.add('show');
}

function closeFetchModal() {
    document.getElementById('fetchModal').classList.remove('show');
}

function populateFetchDropdown() {
    const dd = document.getElementById('fetchTickerDropdown');
    dd.innerHTML = '<option value="">-- Select a ticker --</option>';
    tickersData.forEach(t => {
        const option = document.createElement('option');
        option.value = t.symbol;
        option.textContent = `${t.symbol} - ${t.name}`;
        dd.appendChild(option);
    });
    dd.value = selectedSymbol;
}

async function confirmFetch() {
    const dd = document.getElementById('fetchTickerDropdown');
    const symbol = dd.value;
    if (!symbol) {
        alert('Please select a ticker to fetch');
        return;
    }
    selectedSymbol = symbol;
    closeFetchModal();
    await runFetch(symbol);
    await loadIndicatorData(symbol);
}

async function runFetch(symbol = selectedSymbol) {
    if (!symbol) {
        alert('Select a ticker first');
        return;
    }
    showLoading(true);
    try {
        const res = await fetch(`/api/fetch?ticker=${encodeURIComponent(symbol)}`, { method: 'POST' });
        const data = await res.json();
        debugLog('Fetch: ' + data.status);
    } catch (err) {
        debugLog('Fetch error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

async function loadIndicatorData(symbol = selectedSymbol) {
    if (!symbol) {
        return;
    }
    try {
        const res = await fetch(`/api/ticker/${symbol}?type=indicators`);
        if (res.ok) {
            const data = await res.json();
            debugLog('Indicators reloaded for ' + symbol);
            console.log(data);
        }
    } catch (err) {
        debugLog('Indicator reload error: ' + err.message);
    }
}

async function runRefresh() {
    showLoading(true);
    try {
        const res = await fetch('/api/refresh', { method: 'POST' });
        const data = await res.json();
        debugLog('Auto: ' + data.status);
    } catch (err) {
        debugLog('Auto error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

async function runCalculateNum() {
    showLoading(true);
    try {
        const res = await fetch('/api/calculate_num', { method: 'POST' });
        const data = await res.json();
        debugLog('CalcNum: ' + data.status);
    } catch (err) {
        debugLog('CalcNum error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

async function runBacktest() {
    showLoading(true);
    try {
        const res = await fetch('/api/backtest', { method: 'POST' });
        const data = await res.json();
        debugLog('Backtest: ' + data.status);
    } catch (err) {
        debugLog('Backtest error: ' + err.message);
    } finally {
        showLoading(false);
    }
}

function renderStrategySignals(signals) {
    const container = document.getElementById('strategyRecommendations');
    if (!signals) {
        container.innerHTML = '';
        return;
    }
    let html = '<h4>Latest Signals</h4><ul>';
    for (const [name, sig] of Object.entries(signals)) {
        html += `<li>${name}: <strong>${sig}</strong></li>`;
    }
    html += '</ul>';
    container.innerHTML = html;
}

// Switch active tab and toggle visibility
function switchTab(tabId) {
    document.querySelectorAll('.tab-button').forEach(btn => {
        if (btn.dataset.tab === tabId) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });

    document.querySelectorAll('.tab-content').forEach(content => {
        if (content.id === tabId) {
            content.classList.add('active');
        } else {
            content.classList.remove('active');
        }
    });
}

// Safe date parser: returns a numeric timestamp, or -Infinity on failure
function dateValue(d) {
    const t = Date.parse(d);
    return Number.isFinite(t) ? t : -Infinity;
}

function applyDefaultSort(arr) {
    return arr.sort((a, b) => {
        // 1) date descending
        const tA = dateValue(a.date);
        const tB = dateValue(b.date);
        if (tA !== tB) return tB - tA;

        // 2) symbol ascending
        const symCmp = a.symbol.localeCompare(b.symbol);
        if (symCmp !== 0) return symCmp;

        // 3) change % descending
        return b.change - a.change;
    });
}

 