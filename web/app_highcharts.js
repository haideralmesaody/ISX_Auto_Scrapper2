// ISX Auto Scrapper - Simple Highcharts Implementation
let tickersData = [];
let displayedData = [];
let sortColumn = 'symbol';
let sortAsc = true;
let currentChart = null;
let selectedSymbol = '';

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
        displayedData = [...tickersData];
        debugLog('Loaded tickers: ' + tickersData.length);

        document.getElementById('tickerSearch').addEventListener('input', filterTickers);

        // Populate ticker table
        sortAndDisplay();
        
        showLoading(false);
        debugLog('Dashboard initialized successfully');
        
    } catch (error) {
        debugLog('Error initializing dashboard: ' + error.message);
        showLoading(false);
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
            <th data-col="open">Open</th>
            <th data-col="high">High</th>
            <th data-col="low">Low</th>
            <th data-col="change">Chg%</th>
            <th data-col="volume">Vol</th>
            <th data-col="value">Val</th>
            <th></th>
        </tr>`;
    table.appendChild(thead);

    const tbody = document.createElement('tbody');

    data.forEach(ticker => {
        const row = document.createElement('tr');
        row.classList.add('ticker-row');
        row.innerHTML = `
            <td>${ticker.symbol}</td>
            <td>${ticker.date || ''}</td>
            <td>${ticker.price.toFixed(2)}</td>
            <td>${ticker.open.toFixed(2)}</td>
            <td>${ticker.high.toFixed(2)}</td>
            <td>${ticker.low.toFixed(2)}</td>
            <td class="${ticker.change >= 0 ? 'positive' : 'negative'}">${ticker.change.toFixed(2)}</td>
            <td>${ticker.volume}</td>
            <td>${ticker.value.toFixed(2)}</td>
            <td><div class="sparkline" id="spark-${ticker.symbol}"></div></td>`;
        row.addEventListener('click', () => selectTicker(ticker.symbol));
        tbody.appendChild(row);

        if (typeof Highcharts !== 'undefined' && ticker.sparkline && ticker.sparkline.length > 0) {
            Highcharts.chart(`spark-${ticker.symbol}`, {
                chart: {
                    backgroundColor: 'transparent',
                    borderWidth: 0,
                    type: 'line',
                    height: 40,
                    width: 100,
                    margin: [2, 0, 2, 0],
                    style: { overflow: 'visible' },
                    skipClone: true
                },
                title: { text: null },
                credits: { enabled: false },
                xAxis: { visible: false },
                yAxis: { visible: false },
                tooltip: { enabled: false },
                legend: { enabled: false },
                series: [{ data: ticker.sparkline, color: ticker.change >= 0 ? '#4ade80' : '#f87171', lineWidth: 1, marker: { enabled: false } }]
            });
        }
    });

    table.appendChild(tbody);

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
    displayedData.sort((a, b) => {
        if (a[sortColumn] < b[sortColumn]) return sortAsc ? -1 : 1;
        if (a[sortColumn] > b[sortColumn]) return sortAsc ? 1 : -1;
        return 0;
    });
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
        
        // Fetch data using the exact pattern from the sample
        const data = await fetch(`/api/ticker/${symbol}?type=price`)
            .then(response => response.json());
        
        debugLog('Loaded data points: ' + data.length);
        
        // Convert data to OHLC format for candlestick
        const ohlcData = data.map(item => {
            const ts = item.timestamp ? item.timestamp : Date.parse(item.date);
            return [ts, item.open, item.high, item.low, item.close];
        });
        
        // Build yAxis configuration dynamically
        const yAxisConfig = {};
        
        // before creating chart, build annotations array
        const annotationsConfig = [];
        
        // Create chart using the exact configuration from the sample
        currentChart = Highcharts.stockChart('container', {
            yAxis: yAxisConfig,
            annotations: annotationsConfig,
            series: [{
                id: 'main',
                type: 'candlestick',
                color: '#FF6F6F',
                upColor: '#6FB76F',
                data: ohlcData,
                dataGrouping: {
                    enabled: false
                }
            }],
            
            // Add stock tools
            stockTools: {
                gui: {
                    enabled: true,
                    buttons: ['indicators', 'separator', 'simpleShapes', 'lines', 
                             'crookedLines', 'measure', 'advanced', 'toggleAnnotations', 
                             'separator', 'verticalLabels', 'flags', 'separator', 
                             'zoomChange', 'fullScreen', 'typeChange', 'separator', 
                             'currentPriceIndicator']
                }
            },
            
            // Basic configuration
            title: {
                text: `${symbol} - Iraqi Stock Exchange`
            },
            
            rangeSelector: {
                selected: 1
            },
            
            navigator: {
                enabled: true
            },
            
            scrollbar: {
                enabled: true
            },
            
            credits: {
                enabled: false
            },
            
            colors: ['#2d5016', '#6b9b37', '#FF6F6F', '#8bc34a', '#4a7c23'],
            chart: { backgroundColor: 'rgba(0,0,0,0)' },
        });
        
        debugLog('Chart created successfully');
        // Ensure StockTools toolbar is expanded
        setTimeout(() => {
            const wrapper = document.querySelector('.highcharts-stocktools-wrapper');
            if (wrapper) {
                wrapper.setAttribute('aria-hidden', 'false');
                const toggleBtn = wrapper.querySelector('.highcharts-toggle-toolbar');
                if (toggleBtn && toggleBtn.classList.contains('highcharts-arrow-left')) {
                    toggleBtn.click();
                }
            }
        }, 100);
        
        showLoading(false);
        
    } catch (error) {
        debugLog('Error creating chart: ' + error.message);
        showLoading(false);
        
        // Show error message
        document.getElementById('container').innerHTML = 
            '<div style="text-align: center; padding: 50px; color: #e74c3c;">' +
            '<h3>Error loading chart</h3>' +
            '<p>' + error.message + '</p>' +
            '</div>';
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

 