async function loadReport() {
    try {
        const res = await fetch('/api/daily_report');
        if (!res.ok) throw new Error(res.statusText || 'request failed');
        const data = await res.json();
        renderReport(data);
    } catch (err) {
        const div = document.getElementById('reportContent');
        if (div) {
            div.innerHTML = `<p class="error">Failed to load report: ${err.message}</p>`;
        }
        console.error('Failed to load daily report', err);
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const btn = document.querySelector('[data-tab="reportTab"]');
    if (btn) btn.addEventListener('click', loadReport);
    const dl = document.getElementById('downloadReport');
    if (dl) dl.addEventListener('click', () => { window.location = '/api/daily_report_excel'; });
});

function renderReport(data) {
    const div = document.getElementById('reportContent');
    if (!div) return;
    let html = `<h2>Daily Report - ${data.date}</h2>`;
    html += buildTopTable('Top 5 by Volume', data.top_volume, 'chartVol');
    html += buildTopTable('Top 5 by Value', data.top_value, 'chartVal');
    html += buildTopTable('Top 5 Gainers', data.top_gain, 'chartGain');
    html += buildTopTable('Top 5 Losers', data.top_loss, 'chartLoss');
    html += '<h3>Traded Companies</h3>' + buildCompanyTable(data.traded);
    html += '<h3>Non Traded Companies</h3>' + buildCompanyTable(data.non_traded);
    html += `<button id="downloadReport" class="btn btn-secondary">Download Excel</button>`;
    div.innerHTML = html;
    attachCharts(data);
    const dl = document.getElementById('downloadReport');
    if (dl) dl.addEventListener('click', () => { window.location = '/api/daily_report_excel'; });
}

function buildTopTable(title, rows, chartId) {
    let html = `<h3>${title}</h3><div id="${chartId}" style="height:300px"></div>`;
    html += '<table class="simple-table"><thead><tr><th>Ticker</th><th>Name</th><th>Close</th><th>Change%</th><th>Volume</th></tr></thead><tbody>';
    rows.forEach(r => {
        html += `<tr><td>${r.ticker}</td><td>${r.name}</td><td>${r.close}</td><td>${r.change_pct.toFixed(2)}</td><td>${r.volume}</td></tr>`;
    });
    html += '</tbody></table>';
    return html;
}

function buildCompanyTable(rows) {
    let html = '<table class="simple-table"><thead><tr><th>Code</th><th>Name</th><th>Open</th><th>High</th><th>Low</th><th>Avg</th><th>Prev Avg</th><th>Close</th><th>Prev Close</th><th>Change%</th><th>Trades</th><th>Volume</th><th>Value</th></tr></thead><tbody>';
    rows.forEach(r => {
        html += `<tr><td>${r.code}</td><td>${r.name}</td><td>${r.open}</td><td>${r.high}</td><td>${r.low}</td><td>${r.avg_price}</td><td>${r.prev_avg_price}</td><td>${r.close}</td><td>${r.prev_close}</td><td>${r.change_pct.toFixed(2)}</td><td>${r.trades}</td><td>${r.volume}</td><td>${r.value}</td></tr>`;
    });
    html += '</tbody></table>';
    return html;
}

function attachCharts(data) {
    renderBar('chartVol', data.top_volume, 'volume', 'Volume');
    renderBar('chartVal', data.top_value, 'value', 'Value');
    renderBar('chartGain', data.top_gain, 'change_pct', 'Gain %');
    renderBar('chartLoss', data.top_loss, 'change_pct', 'Loss %');
}

function renderBar(id, rows, field, yTitle) {
    const categories = rows.map(r => r.ticker);
    const seriesData = rows.map(r => r[field]);
    Highcharts.chart(id, { chart: { type: 'column' }, title: { text: '' }, xAxis: { categories }, yAxis: { title: { text: yTitle } }, series: [{ data: seriesData, name: yTitle }] });
}
