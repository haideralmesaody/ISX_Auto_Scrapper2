// Global variables
let currentChart = null;
let currentVolumeChart = null;
let currentRSIChart = null;
let currentMACDChart = null;
let currentStochasticChart = null;
let currentOBVChart = null;
let currentCMFChart = null;
let currentATRChart = null;
let tickersData = [];
let selectedTicker = null;
let selectedStrategy = 'OBV';
let currentTimeframe = '1M';
let showVolume = true;
let priceData = {};
let indicatorsData = {};
let strategiesData = {};
let activeIndicators = {
    RSI: false,
    MACD: false,
    SMA: false,
    EMA: false,
    STOCH: false,
    BB: false,
    PSAR: false,
    OBV: false,
    CMF: false,
    ATR: false
};

// Debug function
function debugLog(message) {
    console.log(message);
    const debugInfo = document.getElementById('debugInfo');
    const debugContent = document.getElementById('debugContent');
    
    if (debugInfo && debugContent) {
        debugInfo.style.display = 'block';
        const time = new Date().toLocaleTimeString();
        debugContent.innerHTML += `<div>${time}: ${message}</div>`;
        debugContent.scrollTop = debugContent.scrollHeight;
        
        // Keep only last 20 messages
        const lines = debugContent.children;
        if (lines.length > 20) {
            debugContent.removeChild(lines[0]);
        }
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', function() {
    debugLog('=== ISX Auto Scrapper Dashboard Loading... ===');
    
    // Check Chart.js availability
    if (typeof Chart === 'undefined') {
        debugLog('ERROR: Chart.js not loaded');
        return;
    }
    
    debugLog('Chart.js version: ' + Chart.version);
    debugLog('Available chart types: ' + Object.keys(Chart.registry.controllers.items).join(', '));
    
    // Check canvas element
    const canvas = document.getElementById('priceChart');
    debugLog('Canvas element found: ' + !!canvas);
    
    // Wait a bit for all plugins to load
    setTimeout(() => {
        debugLog('Chart types after plugins loaded: ' + Object.keys(Chart.registry.controllers.items).join(', '));
        debugLog('Starting app initialization...');
        initializeApp();
    }, 500);
});

async function initializeApp() {
    showLoading(true);
    
    try {
        console.log('Initializing dashboard...');
        await loadTickers();
        setupEventListeners();
        
        console.log('Dashboard initialized successfully with', tickersData.length, 'tickers');
    } catch (error) {
        console.error('Failed to initialize dashboard:', error);
        showError('Failed to load dashboard data');
    } finally {
        showLoading(false);
    }
}

// Data loading functions
async function loadTickers() {
    try {
        // Load ticker list from API endpoint
        const response = await fetch('/api/tickers');
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        tickersData = await response.json();
        console.log('Loaded tickers:', tickersData.length);
        
        renderTickerDropdown();
        
    } catch (error) {
        console.error('Error loading tickers from API:', error);
        
        // Fallback: try to load from CSV directly
        try {
            const response = await fetch('../TICKERS.csv');
            const csvText = await response.text();
            
            const lines = csvText.split('\n').filter(line => line.trim());
            
            tickersData = lines.slice(1).map(line => {
                const values = line.split(',');
                return {
                    symbol: values[0]?.trim(),
                    name: values[1]?.trim() || values[0]?.trim(),
                    companyName: values[2]?.trim() || values[1]?.trim() || values[0]?.trim(),
                    price: 0,
                    change: 0,
                    volume: 0
                };
            }).filter(ticker => ticker.symbol);
            
            console.log('Loaded tickers from CSV fallback:', tickersData.length);
            renderTickerDropdown();
            await loadTickerPrices();
            
        } catch (csvError) {
            console.error('Error loading tickers from CSV fallback:', csvError);
            // Use sample data as last resort
            tickersData = [
                { symbol: 'TASC', name: 'Tabuk Agricultural', companyName: 'Tabuk Agricultural Development Co', price: 12.21, change: 2.34, volume: 15430 },
                { symbol: 'IIEW', name: 'Iraq Engineering Works', companyName: 'Iraq Engineering Works Co', price: 9.40, change: -1.20, volume: 8920 },
                { symbol: 'HASH', name: 'Al-Hashimiya', companyName: 'Al-Hashimiya General Trading Co', price: 20.70, change: 1.80, volume: 12500 },
                { symbol: 'BMNS', name: 'Babylon Media', companyName: 'Babylon Media & Advertisement Co', price: 1.91, change: 0.15, volume: 45200 },
                { symbol: 'NGIR', name: 'National Insurance', companyName: 'National Insurance Company', price: 0.43, change: 0.06, volume: 67800 }
            ];
            console.log('Using sample data:', tickersData.length);
            renderTickerDropdown();
        }
    }
}

async function loadTickerPrices() {
    // Try to load recent prices from raw CSV files
    for (const ticker of tickersData) {
        try {
            const response = await fetch(`../raw_${ticker.symbol}.csv`);
            if (response.ok) {
                const csvText = await response.text();
                const lines = csvText.split('\n').filter(line => line.trim());
                
                if (lines.length > 1) {
                    const lastLine = lines[lines.length - 1];
                    const values = lastLine.split(',');
                    
                    ticker.price = parseFloat(values[4]) || 0; // Close price
                    ticker.volume = parseInt(values[5]) || 0;
                    
                    // Calculate change (mock calculation)
                    if (lines.length > 2) {
                        const prevLine = lines[lines.length - 2];
                        const prevValues = prevLine.split(',');
                        const prevClose = parseFloat(prevValues[4]) || 0;
                        ticker.change = ticker.price - prevClose;
                    }
                }
            }
        } catch (error) {
            console.warn(`Could not load data for ${ticker.symbol}`);
        }
    }
    
    renderTickerDropdown();
}

async function loadStrategiesData() {
    try {
        const response = await fetch('../Strategy_Summary.json');
        if (response.ok) {
            strategiesData = await response.json();
            updateSignals();
        }
    } catch (error) {
        console.warn('Could not load strategies data:', error);
    }
}

async function loadTickerData(symbol) {
    console.log('=== loadTickerData() called for symbol:', symbol, '===');
    showLoading(true);
    
    try {
        // Load price data
        console.log('Loading price data...');
        await loadPriceData(symbol);
        console.log('Price data loaded successfully');
        
        // Load indicators
        console.log('Loading indicators...');
        await loadIndicators(symbol);
        console.log('Indicators loaded successfully');
        
        // Update chart
        console.log('Updating chart...');
        updateChart();
        console.log('Chart update completed');
        
    } catch (error) {
        console.error(`Error loading data for ${symbol}:`, error);
        showError(`Failed to load data for ${symbol}`);
    } finally {
        showLoading(false);
    }
}

async function loadPriceData(symbol) {
    try {
        debugLog(`Loading price data for ${symbol} via API`);
        const response = await fetch(`/api/ticker/${symbol}?type=price`);
        if (!response.ok) throw new Error(`Price data not found: ${response.status}`);
        
        const apiData = await response.json();
        
        if (!apiData || apiData.length === 0) throw new Error('No price data returned');
        
        debugLog(`Processing ${apiData.length} data points from API for ${symbol}`);
        
        const data = apiData.map(item => {
            // Parse date from API response
            const date = new Date(item.date);
            
            if (isNaN(date.getTime())) {
                console.warn(`Invalid date: "${item.date}"`);
                return null;
            }
            
            return {
                date: date,
                close: item.close || 0,
                open: item.open || 0,
                high: item.high || 0,
                low: item.low || 0,
                volume: item.volume || 0
            };
        }).filter(row => row !== null);
        
        debugLog(`Successfully loaded ${data.length} data points for ${symbol} from API`);
        
        priceData[symbol] = data;
        
    } catch (error) {
        debugLog(`Error loading price data for ${symbol}: ${error.message}`);
        debugLog(`Generating sample data for ${symbol}`);
        // Generate sample data
        priceData[symbol] = generateSamplePriceData(symbol);
    }
}

async function loadIndicators(symbol) {
    try {
        debugLog(`Loading indicators for ${symbol} via API`);
        const response = await fetch(`/api/ticker/${symbol}?type=indicators`);
        if (!response.ok) throw new Error('Indicators data not found via API');
        
        const indicators = await response.json();
        
        if (!indicators || Object.keys(indicators).length === 0) {
            throw new Error('No indicators returned');
        }
        
        debugLog(`Successfully loaded ${Object.keys(indicators).length} indicators for ${symbol}`);
        indicatorsData[symbol] = indicators;
        
    } catch (error) {
        debugLog(`Error loading indicators for ${symbol}: ${error.message}`);
        debugLog(`Generating sample indicators for ${symbol}`);
        // Generate sample indicators
        indicatorsData[symbol] = generateSampleIndicators();
    }
}

// Rendering functions
function renderTickerDropdown() {
    const tickerDropdown = document.getElementById('tickerDropdown');
    if (!tickerDropdown) return;
    
    console.log('Rendering dropdown with', tickersData.length, 'tickers');
    
    // Keep the default option and add all tickers
    const defaultOption = '<option value="">-- Select a Ticker --</option>';
    const tickerOptions = tickersData.map(ticker => {
        // Use the correct field names from API response
        const displayName = ticker.name || ticker.companyName || '';
        const optionText = displayName && displayName !== ticker.symbol 
            ? `${ticker.symbol} - ${displayName}` 
            : ticker.symbol;
        return `<option value="${ticker.symbol}">${optionText}</option>`;
    }).join('');
    
    tickerDropdown.innerHTML = defaultOption + tickerOptions;
    console.log('Dropdown populated with options:', tickerDropdown.children.length);
}

function updateChart() {
    debugLog('=== updateChart() called ===');
    
    const priceCanvas = document.getElementById('priceChart');
    const volumeCanvas = document.getElementById('volumeChart');
    
    if (!priceCanvas) {
        debugLog('ERROR: Price canvas element not found!');
        return;
    }
    
    if (!selectedTicker) {
        debugLog('No ticker selected');
        return;
    }
    
    debugLog('Selected ticker: ' + selectedTicker);
    debugLog('Available price data keys: ' + Object.keys(priceData).join(', '));
    
    if (!priceData[selectedTicker]) {
        debugLog('ERROR: No price data for ticker: ' + selectedTicker);
        return;
    }
    
    debugLog('Price data length: ' + priceData[selectedTicker].length);
    
    const data = filterDataByTimeframe(priceData[selectedTicker], currentTimeframe);
    debugLog('Filtered data length: ' + data.length);
    
    // Clean up both charts
    cleanupChart(priceCanvas, 'currentChart');
    if (volumeCanvas && showVolume) {
        cleanupChart(volumeCanvas, 'currentVolumeChart');
    }
    
    const chartTypeElement = document.querySelector('.chart-type.active');
    const chartType = chartTypeElement ? chartTypeElement.dataset.type : 'line';
    debugLog('Chart type: ' + chartType);
    
    try {
        // Create price chart
        const priceCtx = priceCanvas.getContext('2d');
        if (chartType === 'candlestick') {
            debugLog('Creating candlestick chart...');
            createCandlestickChart(priceCtx, data);
        } else {
            debugLog('Creating line chart...');
            createLineChart(priceCtx, data);
        }
        
        // Create volume chart if enabled
        if (volumeCanvas && showVolume) {
            debugLog('Creating volume chart...');
            const volumeCtx = volumeCanvas.getContext('2d');
            createVolumeChart(volumeCtx, data);
        }
        
        // Create indicator charts if enabled
        if (activeIndicators.RSI) {
            const rsiCanvas = document.getElementById('rsiChart');
            if (rsiCanvas) {
                debugLog('Creating RSI chart...');
                cleanupChart(rsiCanvas, 'currentRSIChart');
                const rsiCtx = rsiCanvas.getContext('2d');
                createRSIChart(rsiCtx, data);
            }
        }
        
        if (activeIndicators.MACD) {
            const macdCanvas = document.getElementById('macdChart');
            if (macdCanvas) {
                debugLog('Creating MACD chart...');
                cleanupChart(macdCanvas, 'currentMACDChart');
                const macdCtx = macdCanvas.getContext('2d');
                createMACDChart(macdCtx, data);
            }
        }
        
        if (activeIndicators.STOCH) {
            const stochCanvas = document.getElementById('stochasticChart');
            if (stochCanvas) {
                debugLog('Creating Stochastic chart...');
                cleanupChart(stochCanvas, 'currentStochasticChart');
                const stochCtx = stochCanvas.getContext('2d');
                createStochasticChart(stochCtx, data);
            }
        }
        
        if (activeIndicators.OBV) {
            const obvCanvas = document.getElementById('obvChart');
            if (obvCanvas) {
                debugLog('Creating OBV chart...');
                cleanupChart(obvCanvas, 'currentOBVChart');
                const obvCtx = obvCanvas.getContext('2d');
                createOBVChart(obvCtx, data);
            }
        }
        
        if (activeIndicators.CMF) {
            const cmfCanvas = document.getElementById('cmfChart');
            if (cmfCanvas) {
                debugLog('Creating CMF chart...');
                cleanupChart(cmfCanvas, 'currentCMFChart');
                const cmfCtx = cmfCanvas.getContext('2d');
                createCMFChart(cmfCtx, data);
            }
        }
        
        if (activeIndicators.ATR) {
            const atrCanvas = document.getElementById('atrChart');
            if (atrCanvas) {
                debugLog('Creating ATR chart...');
                cleanupChart(atrCanvas, 'currentATRChart');
                const atrCtx = atrCanvas.getContext('2d');
                createATRChart(atrCtx, data);
            }
        }
        
        debugLog('Chart creation completed successfully');
    } catch (error) {
        debugLog('ERROR creating chart: ' + error.message);
        debugLog('Falling back to basic line chart...');
        try {
            const priceCtx = priceCanvas.getContext('2d');
            createBasicLineChart(priceCtx, data);
        } catch (fallbackError) {
            debugLog('ERROR with basic chart: ' + fallbackError.message);
            debugLog('Using minimal chart as last resort...');
            const priceCtx = priceCanvas.getContext('2d');
            createMinimalChart(priceCtx, data);
        }
    }
}

function cleanupChart(canvas, chartVarName) {
    try {
        const ctx = canvas.getContext('2d');
        
        // Method 1: Destroy by Chart.getChart
        const existingChart = Chart.getChart(canvas);
        if (existingChart) {
            debugLog('Destroying existing Chart.js instance for ' + canvas.id);
            existingChart.destroy();
        }
        
        // Method 2: Destroy chart reference
        let chartVar;
        switch (chartVarName) {
            case 'currentChart':
                chartVar = currentChart;
                break;
            case 'currentVolumeChart':
                chartVar = currentVolumeChart;
                break;
            case 'currentRSIChart':
                chartVar = currentRSIChart;
                break;
            case 'currentMACDChart':
                chartVar = currentMACDChart;
                break;
            case 'currentStochasticChart':
                chartVar = currentStochasticChart;
                break;
            case 'currentOBVChart':
                chartVar = currentOBVChart;
                break;
            case 'currentCMFChart':
                chartVar = currentCMFChart;
                break;
            case 'currentATRChart':
                chartVar = currentATRChart;
                break;
            default:
                chartVar = null;
        }
        
        if (chartVar) {
            debugLog('Destroying chart variable reference for ' + canvas.id);
            try {
                chartVar.destroy();
            } catch (e) {
                debugLog('Error destroying chart var: ' + e.message);
            }
        }
        
        // Method 3: Clear canvas manually
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        
        // Method 4: Remove Chart.js internal references
        if (canvas.chartjs) {
            delete canvas.chartjs;
        }
        canvas.chartjs = undefined;
        
        // Method 5: Remove all Chart.js related properties
        for (let prop in canvas) {
            if (prop.startsWith('_chart') || prop.includes('chart')) {
                try {
                    delete canvas[prop];
                } catch (e) {}
            }
        }
        
    } catch (e) {
        debugLog('Error during chart cleanup for ' + canvas.id + ': ' + e.message);
    }
    
    // Reset global variables
    switch (chartVarName) {
        case 'currentChart':
            currentChart = null;
            break;
        case 'currentVolumeChart':
            currentVolumeChart = null;
            break;
        case 'currentRSIChart':
            currentRSIChart = null;
            break;
        case 'currentMACDChart':
            currentMACDChart = null;
            break;
        case 'currentStochasticChart':
            currentStochasticChart = null;
            break;
        case 'currentOBVChart':
            currentOBVChart = null;
            break;
        case 'currentCMFChart':
            currentCMFChart = null;
            break;
        case 'currentATRChart':
            currentATRChart = null;
            break;
    }
}

function createCandlestickChart(ctx, data) {
    debugLog('Creating candlestick chart with ' + data.length + ' data points');
    
    // Check if financial charts are available
    debugLog('Available chart types: ' + Object.keys(Chart.registry.controllers.items).join(', '));
    
    // Create simplified data for candlestick chart (no date objects)
    const candlestickData = data.map((item, index) => ({
        x: index, // Simple index
        o: item.open,
        h: item.high,
        l: item.low,
        c: item.close
    }));
    
    // Simple labels (avoid complex date formatting)
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            // Show only every Nth label to avoid clutter
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    debugLog('Sample candlestick data: ' + candlestickData.length + ' points');
    
    // Prepare datasets array
    const datasets = [{
        label: selectedTicker + ' Price',
        data: candlestickData,
        borderColor: '#00d4ff',
        borderWidth: 1,
        color: {
            up: '#4ade80',
            down: '#f87171',
            unchanged: '#a0a0a0'
        }
    }];
    
    // Add Simple Moving Averages if enabled
    if (activeIndicators.SMA) {
        debugLog('Adding Simple Moving Averages to candlestick chart');
        const closePrices = data.map(d => d.close);
        
        const sma10 = calculateSMA(closePrices, 10);
        const sma50 = calculateSMA(closePrices, 50);
        const sma200 = calculateSMA(closePrices, 200);
        
        datasets.push({
            label: 'SMA 10',
            data: sma10.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#84cc16', // Light green
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'SMA 50',
            data: sma50.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#fbbf24', // Yellow
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'SMA 200',
            data: sma200.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#f97316', // Orange
            backgroundColor: 'transparent',
            borderWidth: 3,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            yAxisID: 'y'
        });
        
        debugLog('Added SMA 10, 50, 200 to candlestick chart');
    }

    // Add Exponential Moving Averages if enabled
    if (activeIndicators.EMA) {
        debugLog('Adding Exponential Moving Averages to candlestick chart');
        const closePrices = data.map(d => d.close);
        
        const ema5 = calculateEMA(closePrices, 5);
        const ema10 = calculateEMA(closePrices, 10);
        const ema20 = calculateEMA(closePrices, 20);
        const ema200 = calculateEMA(closePrices, 200);
        
        datasets.push({
            label: 'EMA 5',
            data: ema5.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#06b6d4', // Cyan
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [3, 3],
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'EMA 10',
            data: ema10.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#8b5cf6', // Purple
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [5, 5],
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'EMA 20',
            data: ema20.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#ef4444', // Red
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [5, 5],
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'EMA 200',
            data: ema200.map((val, index) => val !== null ? { x: index, y: val } : null).filter(item => item !== null),
            type: 'line',
            borderColor: '#dc2626', // Dark red
            backgroundColor: 'transparent',
            borderWidth: 3,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [8, 8],
            yAxisID: 'y'
        });
        
        debugLog('Added EMA 5, 10, 20, 200 to candlestick chart');
    }
    
    // Add Bollinger Bands if enabled
    if (activeIndicators.BB) {
        debugLog('Adding Bollinger Bands to candlestick chart');
        const closePrices = data.map(d => d.close);
        const sma20 = calculateSMA(closePrices, 20);
        const std20 = calculateRollingStd(closePrices, 20);
        
        // Calculate Bollinger Bands
        const upperBand = sma20.map((sma, index) => 
            std20[index] !== null && sma !== null ? { x: index, y: sma + (2 * std20[index]) } : null
        ).filter(item => item !== null);
        
        const lowerBand = sma20.map((sma, index) => 
            std20[index] !== null && sma !== null ? { x: index, y: sma - (2 * std20[index]) } : null
        ).filter(item => item !== null);
        
        datasets.push({
            label: 'BB Upper',
            data: upperBand,
            type: 'line',
            borderColor: '#8b5cf6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            borderDash: [5, 5],
            yAxisID: 'y'
        });
        
        datasets.push({
            label: 'BB Lower',
            data: lowerBand,
            type: 'line',
            borderColor: '#8b5cf6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            borderDash: [5, 5],
            yAxisID: 'y'
        });
        
        debugLog('Added Bollinger Bands with ' + upperBand.length + ' points');
    }
    
    // Add Parabolic SAR if enabled
    if (activeIndicators.PSAR) {
        debugLog('Adding Parabolic SAR to candlestick chart');
        const psarData = calculatePSAR(data);
        
        datasets.push({
            label: 'PSAR',
            data: psarData.map((val, index) => ({ x: index, y: val })),
            type: 'scatter',
            backgroundColor: '#ff6b6b',
            borderColor: '#ff6b6b',
            pointRadius: 3,
            pointHoverRadius: 5,
            showLine: false,
            yAxisID: 'y'
        });
        
        debugLog('Added Parabolic SAR with ' + psarData.length + ' points');
    }
    
    currentChart = new Chart(ctx, {
        type: 'candlestick',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false, // Disable animations to avoid issues
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    labels: {
                        color: '#a0a0a0',
                        font: {
                            size: 12
                        },
                        usePointStyle: true,
                        pointStyle: 'line',
                        padding: 15
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#00d4ff',
                    borderWidth: 1,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const point = context.raw;
                            return [
                                `Open: $${point.o.toFixed(2)}`,
                                `High: $${point.h.toFixed(2)}`,
                                `Low: $${point.l.toFixed(2)}`,
                                `Close: $${point.c.toFixed(2)}`
                            ];
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category', // Use category scale instead of time/linear
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    type: 'linear', // Explicit linear scale
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            }
        }
    });
    
    debugLog('Candlestick chart created successfully!');
}

function createOHLCLineChart(ctx, data) {
    console.log('Creating OHLC line chart fallback');
    
    const timestamps = data.map(item => item.date.getTime());
    
    currentChart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Close',
                    data: data.map(item => ({ x: item.date.getTime(), y: item.close })),
                    borderColor: '#00d4ff',
                    backgroundColor: 'rgba(0, 212, 255, 0.1)',
                    borderWidth: 2,
                    fill: false,
                    tension: 0.1
                },
                {
                    label: 'Open',
                    data: data.map(item => ({ x: item.date.getTime(), y: item.open })),
                    borderColor: '#4ade80',
                    backgroundColor: 'rgba(74, 222, 128, 0.1)',
                    borderWidth: 1,
                    fill: false,
                    tension: 0.1
                },
                {
                    label: 'High',
                    data: data.map(item => ({ x: item.date.getTime(), y: item.high })),
                    borderColor: '#f87171',
                    backgroundColor: 'rgba(248, 113, 113, 0.1)',
                    borderWidth: 1,
                    fill: false,
                    tension: 0.1
                },
                {
                    label: 'Low',
                    data: data.map(item => ({ x: item.date.getTime(), y: item.low })),
                    borderColor: '#fbbf24',
                    backgroundColor: 'rgba(251, 191, 36, 0.1)',
                    borderWidth: 1,
                    fill: false,
                    tension: 0.1
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#ffffff'
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#00d4ff',
                    borderWidth: 1,
                    callbacks: {
                        title: function(context) {
                            return new Date(context[0].parsed.x).toLocaleDateString();
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'time',
                    time: {
                        displayFormats: {
                            day: 'MMM dd',
                            week: 'MMM dd',
                            month: 'MMM yyyy'
                        }
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0'
                    }
                },
                y: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            }
        }
    });
}

function createLineChart(ctx, data) {
    debugLog('Creating line chart with ' + data.length + ' data points');
    
    // Use same approach as candlestick chart to avoid date adapter issues
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            // Show only every Nth label to avoid clutter
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    const lineData = data.map(item => item.close);
    
    debugLog('Line chart data points: ' + lineData.length);
    
    // Prepare datasets array
    const datasets = [{
        label: selectedTicker + ' Close Price',
        data: lineData,
        borderColor: '#00d4ff',
        backgroundColor: 'rgba(0, 212, 255, 0.1)',
        borderWidth: 3,
        fill: true,
        tension: 0.4,
        pointRadius: 0,
        pointHoverRadius: 6,
        pointHoverBackgroundColor: '#00d4ff',
        pointHoverBorderColor: '#ffffff',
        pointHoverBorderWidth: 2
    }];
    
    // Add Simple Moving Averages if enabled
    if (activeIndicators.SMA) {
        debugLog('Adding Simple Moving Averages to line chart');
        const closePrices = data.map(d => d.close);
        
        const sma10 = calculateSMA(closePrices, 10);
        const sma50 = calculateSMA(closePrices, 50);
        const sma200 = calculateSMA(closePrices, 200);
        
        datasets.push({
            label: 'SMA 10',
            data: sma10.map(val => val !== null ? val : null),
            borderColor: '#84cc16',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            spanGaps: true
        });
        
        datasets.push({
            label: 'SMA 50',
            data: sma50.map(val => val !== null ? val : null),
            borderColor: '#fbbf24',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            spanGaps: true
        });
        
        datasets.push({
            label: 'SMA 200',
            data: sma200.map(val => val !== null ? val : null),
            borderColor: '#f97316',
            backgroundColor: 'transparent',
            borderWidth: 3,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            spanGaps: true
        });
        
        debugLog('Added SMA 10, 50, 200 to line chart');
    }

    // Add Exponential Moving Averages if enabled
    if (activeIndicators.EMA) {
        debugLog('Adding Exponential Moving Averages to line chart');
        const closePrices = data.map(d => d.close);
        
        const ema5 = calculateEMA(closePrices, 5);
        const ema10 = calculateEMA(closePrices, 10);
        const ema20 = calculateEMA(closePrices, 20);
        const ema200 = calculateEMA(closePrices, 200);
        
        datasets.push({
            label: 'EMA 5',
            data: ema5.map(val => val !== null ? val : null),
            borderColor: '#06b6d4',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [3, 3],
            spanGaps: true
        });
        
        datasets.push({
            label: 'EMA 10',
            data: ema10.map(val => val !== null ? val : null),
            borderColor: '#8b5cf6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [5, 5],
            spanGaps: true
        });
        
        datasets.push({
            label: 'EMA 20',
            data: ema20.map(val => val !== null ? val : null),
            borderColor: '#ef4444',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [5, 5],
            spanGaps: true
        });
        
        datasets.push({
            label: 'EMA 200',
            data: ema200.map(val => val !== null ? val : null),
            borderColor: '#dc2626',
            backgroundColor: 'transparent',
            borderWidth: 3,
            fill: false,
            pointRadius: 0,
            pointHoverRadius: 4,
            tension: 0.1,
            borderDash: [8, 8],
            spanGaps: true
        });
        
        debugLog('Added EMA 5, 10, 20, 200 to line chart');
    }
    
    // Add Bollinger Bands if enabled
    if (activeIndicators.BB) {
        debugLog('Adding Bollinger Bands to line chart');
        const closePrices = data.map(d => d.close);
        const sma20 = calculateSMA(closePrices, 20);
        const std20 = calculateRollingStd(closePrices, 20);
        
        // Calculate Bollinger Bands
        const upperBand = sma20.map((sma, index) => 
            std20[index] !== null && sma !== null ? sma + (2 * std20[index]) : null
        );
        
        const lowerBand = sma20.map((sma, index) => 
            std20[index] !== null && sma !== null ? sma - (2 * std20[index]) : null
        );
        
        datasets.push({
            label: 'BB Upper',
            data: upperBand,
            borderColor: '#8b5cf6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            borderDash: [5, 5],
            spanGaps: true
        });
        
        datasets.push({
            label: 'BB Lower',
            data: lowerBand,
            borderColor: '#8b5cf6',
            backgroundColor: 'transparent',
            borderWidth: 2,
            fill: false,
            pointRadius: 0,
            borderDash: [5, 5],
            spanGaps: true
        });
        
        debugLog('Added Bollinger Bands to line chart');
    }
    
    // Add Parabolic SAR if enabled (simplified for line chart)
    if (activeIndicators.PSAR) {
        debugLog('Adding Parabolic SAR to line chart');
        const psarData = calculatePSAR(data);
        
        datasets.push({
            label: 'PSAR',
            data: psarData,
            borderColor: '#ff6b6b',
            backgroundColor: '#ff6b6b',
            borderWidth: 0,
            pointRadius: 3,
            pointHoverRadius: 5,
            showLine: false,
            fill: false
        });
        
        debugLog('Added Parabolic SAR to line chart');
    }
    
    currentChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false, // Disable animations for consistency
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: {
                    display: true,
                    position: 'top',
                    labels: {
                        color: '#a0a0a0',
                        font: {
                            size: 12
                        },
                        usePointStyle: true,
                        pointStyle: 'line',
                        padding: 15
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#00d4ff',
                    borderWidth: 1,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const index = context.dataIndex;
                            const point = data[index];
                            return [
                                `Close: $${point.close.toFixed(2)}`,
                                `Volume: ${point.volume.toLocaleString()}`
                            ];
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category', // Use category scale like candlestick chart
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    type: 'linear', // Explicit linear scale
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            }
        }
    });
    
    debugLog('Line chart created successfully!');
}

function createBasicLineChart(ctx, data) {
    debugLog('Creating basic line chart fallback');
    
    // Simple array format for basic Chart.js
    const labels = data.map(item => item.date.toLocaleDateString());
    const prices = data.map(item => item.close);
    
    debugLog('Basic chart data points: ' + labels.length);
    
    currentChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: selectedTicker + ' Price',
                data: prices,
                borderColor: '#00d4ff',
                backgroundColor: 'rgba(0, 212, 255, 0.1)',
                borderWidth: 2,
                fill: true,
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    labels: {
                        color: '#ffffff'
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#00d4ff',
                    borderWidth: 1
                }
            },
            scales: {
                x: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0'
                    }
                },
                y: {
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return '$' + value.toFixed(2);
                        }
                    }
                }
            }
        }
    });
    
    debugLog('Basic line chart created successfully!');
}

function createMinimalChart(ctx, data) {
    debugLog('Creating minimal chart as last resort');
    
    // Simplest possible chart with no advanced features
    const labels = [];
    const prices = [];
    
    for (let i = 0; i < Math.min(data.length, 30); i++) {
        labels.push('Day ' + (i + 1));
        prices.push(data[i].close);
    }
    
    debugLog('Minimal chart: ' + labels.length + ' data points');
    
    currentChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: (selectedTicker || 'Stock') + ' Price',
                data: prices,
                borderColor: '#00d4ff',
                backgroundColor: 'rgba(0, 212, 255, 0.1)',
                borderWidth: 2
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: true,
                    labels: { color: '#ffffff' }
                }
            },
            scales: {
                x: {
                    ticks: { color: '#a0a0a0' }
                },
                y: {
                    ticks: { color: '#a0a0a0' }
                }
            }
        }
    });
    
    debugLog('Minimal chart created successfully!');
}

function createVolumeChart(ctx, data) {
    debugLog('Creating volume chart with ' + data.length + ' data points');
    
    // Create volume bar chart data
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    const volumeData = data.map(item => item.volume);
    
    // Color bars based on price movement (green if close > open, red otherwise)
    const backgroundColors = data.map(item => {
        return item.close >= item.open ? 'rgba(74, 222, 128, 0.7)' : 'rgba(248, 113, 113, 0.7)';
    });
    
    const borderColors = data.map(item => {
        return item.close >= item.open ? 'rgba(74, 222, 128, 1)' : 'rgba(248, 113, 113, 1)';
    });
    
    currentVolumeChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: labels,
            datasets: [{
                label: 'Volume',
                data: volumeData,
                backgroundColor: backgroundColors,
                borderColor: borderColors,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#00d4ff',
                    borderWidth: 1,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const volume = context.parsed.y;
                            return 'Volume: ' + volume.toLocaleString();
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        maxTicksLimit: 8
                    }
                },
                y: {
                    type: 'linear',
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return value.toLocaleString();
                        }
                    }
                }
            }
        }
    });
    
    debugLog('Volume chart created successfully!');
}

function createRSIChart(ctx, data) {
    debugLog('Creating RSI chart with ' + data.length + ' data points');
    
    // Generate RSI data (simplified calculation for demo)
    const rsiData = data.map((item, index) => {
        if (index < 14) return 50; // Need 14 periods for RSI
        
        // Simplified RSI calculation (in real app, use proper RSI formula)
        const recentPrices = data.slice(index - 14, index).map(d => d.close);
        const avgGain = recentPrices.reduce((sum, price, i) => {
            if (i === 0) return sum;
            const change = price - recentPrices[i - 1];
            return sum + (change > 0 ? change : 0);
        }, 0) / 14;
        
        const avgLoss = recentPrices.reduce((sum, price, i) => {
            if (i === 0) return sum;
            const change = price - recentPrices[i - 1];
            return sum + (change < 0 ? Math.abs(change) : 0);
        }, 0) / 14;
        
        const rs = avgGain / (avgLoss || 0.01);
        return 100 - (100 / (1 + rs));
    });
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentRSIChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'RSI (14)',
                data: rsiData,
                borderColor: '#fbbf24',
                backgroundColor: 'rgba(251, 191, 36, 0.1)',
                borderWidth: 3,
                fill: false,
                pointRadius: 0,
                pointHoverRadius: 6,
                pointBackgroundColor: '#fbbf24',
                pointBorderColor: '#ffffff',
                pointBorderWidth: 2,
                segment: {
                    borderColor: function(ctx) {
                        const value = ctx.p1.parsed.y;
                        if (value > 70) return '#f87171'; // Overbought - red
                        if (value < 30) return '#4ade80'; // Oversold - green
                        return '#fbbf24'; // Neutral - yellow
                    }
                }
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.9)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#fbbf24',
                    borderWidth: 2,
                    cornerRadius: 8,
                    displayColors: false,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const rsiValue = context.parsed.y;
                            let signal = '';
                            
                            if (rsiValue > 70) {
                                signal = ' (Overbought 🔴)';
                            } else if (rsiValue < 30) {
                                signal = ' (Oversold 🟢)';
                            } else if (rsiValue > 50) {
                                signal = ' (Bullish Zone)';
                            } else {
                                signal = ' (Bearish Zone)';
                            }
                            
                            return [`RSI: ${rsiValue.toFixed(2)}${signal}`];
                        },
                        afterLabel: function(context) {
                            const rsiValue = context.parsed.y;
                            if (rsiValue > 70) {
                                return ['', 'Consider selling - Stock may be overvalued'];
                            } else if (rsiValue < 30) {
                                return ['', 'Consider buying - Stock may be undervalued'];
                            } else {
                                return ['', 'Normal trading range'];
                            }
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { 
                        color: 'rgba(255, 255, 255, 0.1)',
                        drawOnChartArea: true
                    },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    min: 0,
                    max: 100,
                    grid: { 
                        color: 'rgba(255, 255, 255, 0.1)',
                        drawOnChartArea: true
                    },
                    ticks: {
                        color: '#a0a0a0',
                        stepSize: 10,
                        callback: function(value) {
                            if (value === 30 || value === 70) {
                                return value + (value === 30 ? ' (Oversold)' : ' (Overbought)');
                            }
                            if (value === 0 || value === 50 || value === 100) return value;
                            return value;
                        }
                    }
                }
            },
            plugins: [{
                id: 'rsiReferenceLines',
                afterDraw: function(chart) {
                    const ctx = chart.ctx;
                    const yAxis = chart.scales.y;
                    const chartArea = chart.chartArea;
                    
                    ctx.save();
                    
                    // Draw background zones
                    ctx.globalAlpha = 0.15;
                    
                    // Overbought zone (70-100)
                    ctx.fillStyle = '#f87171';
                    const y70 = yAxis.getPixelForValue(70);
                    const y100 = yAxis.getPixelForValue(100);
                    ctx.fillRect(chartArea.left, y100, chartArea.right - chartArea.left, y70 - y100);
                    
                    // Oversold zone (0-30)
                    ctx.fillStyle = '#4ade80';
                    const y30 = yAxis.getPixelForValue(30);
                    const y0 = yAxis.getPixelForValue(0);
                    ctx.fillRect(chartArea.left, y30, chartArea.right - chartArea.left, y0 - y30);
                    
                    ctx.globalAlpha = 1.0;
                    
                    // Draw reference lines
                    ctx.setLineDash([8, 4]);
                    ctx.lineWidth = 3;
                    
                    // Overbought line (70)
                    ctx.strokeStyle = '#f87171';
                    ctx.beginPath();
                    ctx.moveTo(chartArea.left, y70);
                    ctx.lineTo(chartArea.right, y70);
                    ctx.stroke();
                    
                    // Oversold line (30)
                    ctx.strokeStyle = '#4ade80';
                    ctx.beginPath();
                    ctx.moveTo(chartArea.left, y30);
                    ctx.lineTo(chartArea.right, y30);
                    ctx.stroke();
                    
                    // Midline (50)
                    ctx.setLineDash([4, 4]);
                    ctx.lineWidth = 2;
                    ctx.strokeStyle = 'rgba(160, 160, 160, 0.8)';
                    const y50 = yAxis.getPixelForValue(50);
                    ctx.beginPath();
                    ctx.moveTo(chartArea.left, y50);
                    ctx.lineTo(chartArea.right, y50);
                    ctx.stroke();
                    
                    // Add labels
                    ctx.fillStyle = '#f87171';
                    ctx.font = 'bold 12px Arial';
                    ctx.textAlign = 'left';
                    ctx.fillText('Overbought (70)', chartArea.left + 5, y70 - 5);
                    
                    ctx.fillStyle = '#4ade80';
                    ctx.fillText('Oversold (30)', chartArea.left + 5, y30 + 15);
                    
                    ctx.fillStyle = 'rgba(160, 160, 160, 0.8)';
                    ctx.font = '10px Arial';
                    ctx.textAlign = 'right';
                    ctx.fillText('50', chartArea.right - 5, y50 - 3);
                    
                    ctx.restore();
                }
            }]
        }
    });
    
    debugLog('RSI chart created successfully!');
}

function createMACDChart(ctx, data) {
    debugLog('Creating MACD chart with ' + data.length + ' data points');
    
    // Generate MACD data (simplified calculation)
    const ema12 = calculateEMA(data.map(d => d.close), 12);
    const ema26 = calculateEMA(data.map(d => d.close), 26);
    const macdLine = ema12.map((val, i) => val - ema26[i]);
    const signalLine = calculateEMA(macdLine, 9);
    const histogram = macdLine.map((val, i) => val - signalLine[i]);
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentMACDChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'MACD',
                data: macdLine,
                borderColor: '#00d4ff',
                backgroundColor: 'transparent',
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                yAxisID: 'y'
            }, {
                label: 'Signal',
                data: signalLine,
                borderColor: '#f87171',
                backgroundColor: 'transparent',
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                yAxisID: 'y'
            }, {
                label: 'Histogram',
                data: histogram,
                borderColor: 'transparent',
                backgroundColor: histogram.map(val => val > 0 ? 'rgba(74, 222, 128, 0.7)' : 'rgba(248, 113, 113, 0.7)'),
                type: 'bar',
                yAxisID: 'y'
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0' }
                }
            }
        }
    });
    
    // Add MACD zero line plugin
    const macdZeroLinePlugin = {
        id: 'macdZeroLine',
        afterDraw: function(chart) {
            if (chart.canvas.id === 'macdChart') {
                const ctx = chart.ctx;
                const chartArea = chart.chartArea;
                const yAxis = chart.scales.y;
                
                ctx.save();
                ctx.strokeStyle = 'rgba(160, 160, 160, 0.8)';
                ctx.lineWidth = 2;
                ctx.setLineDash([3, 3]);
                
                // Zero line
                const y0 = yAxis.getPixelForValue(0);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y0);
                ctx.lineTo(chartArea.right, y0);
                ctx.stroke();
                
                ctx.restore();
            }
        }
    };
    
    // Register the plugin for this chart
    Chart.register(macdZeroLinePlugin);
    
    debugLog('MACD chart created successfully!');
}

function createStochasticChart(ctx, data) {
    debugLog('Creating Stochastic chart with ' + data.length + ' data points');
    
    // Generate Stochastic data (simplified calculation)
    const stochasticData = data.map((item, index) => {
        if (index < 14) return 50; // Need 14 periods
        
        const period = data.slice(index - 14, index);
        const highest = Math.max(...period.map(d => d.high));
        const lowest = Math.min(...period.map(d => d.low));
        const current = item.close;
        
        return ((current - lowest) / (highest - lowest)) * 100;
    });
    
    // Simple 3-period moving average for %D line
    const stochasticD = stochasticData.map((val, index) => {
        if (index < 2) return val;
        const sum = stochasticData.slice(index - 2, index + 1).reduce((a, b) => a + b, 0);
        return sum / 3;
    });
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentStochasticChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '%K',
                data: stochasticData,
                borderColor: '#00d4ff',
                backgroundColor: 'transparent',
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                pointHoverRadius: 4
            }, {
                label: '%D',
                data: stochasticD,
                borderColor: '#f87171',
                backgroundColor: 'transparent',
                borderWidth: 2,
                fill: false,
                pointRadius: 0,
                pointHoverRadius: 4
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const label = context.dataset.label;
                            return `${label}: ${context.parsed.y.toFixed(2)}`;
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    min: 0,
                    max: 100,
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            if (value === 20 || value === 80) return value;
                            if (value === 0 || value === 50 || value === 100) return value;
                            return '';
                        }
                    }
                }
            }
        }
    });
    
    // Add Stochastic reference lines plugin
    const stochLinesPlugin = {
        id: 'stochLines',
        afterDraw: function(chart) {
            if (chart.canvas.id === 'stochasticChart') {
                const ctx = chart.ctx;
                const chartArea = chart.chartArea;
                const yAxis = chart.scales.y;
                
                ctx.save();
                ctx.setLineDash([5, 5]);
                ctx.lineWidth = 2;
                
                // Overbought line (80)
                ctx.strokeStyle = 'rgba(248, 113, 113, 0.8)';
                const y80 = yAxis.getPixelForValue(80);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y80);
                ctx.lineTo(chartArea.right, y80);
                ctx.stroke();
                
                // Oversold line (20)
                ctx.strokeStyle = 'rgba(74, 222, 128, 0.8)';
                const y20 = yAxis.getPixelForValue(20);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y20);
                ctx.lineTo(chartArea.right, y20);
                ctx.stroke();
                
                // Midline (50)
                ctx.strokeStyle = 'rgba(160, 160, 160, 0.5)';
                const y50 = yAxis.getPixelForValue(50);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y50);
                ctx.lineTo(chartArea.right, y50);
                ctx.stroke();
                
                ctx.restore();
            }
        }
    };
    
    // Register the plugin for this chart
    Chart.register(stochLinesPlugin);
    
    debugLog('Stochastic chart created successfully!');
}

function createOBVChart(ctx, data) {
    debugLog('Creating OBV chart with ' + data.length + ' data points');
    
    // Calculate OBV data
    const obvData = calculateOBV(data);
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentOBVChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'OBV',
                data: obvData,
                borderColor: '#10b981',
                backgroundColor: 'rgba(16, 185, 129, 0.1)',
                borderWidth: 3,
                fill: true,
                pointRadius: 0,
                pointHoverRadius: 6,
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.9)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#10b981',
                    borderWidth: 2,
                    cornerRadius: 8,
                    displayColors: false,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const obvValue = context.parsed.y;
                            let signal = '';
                            
                            // Calculate OBV trend
                            const index = context.dataIndex;
                            if (index > 10) {
                                const current = obvData[index];
                                const previous = obvData[index - 10];
                                const change = ((current - previous) / Math.abs(previous)) * 100;
                                
                                if (change > 5) {
                                    signal = ' (Strong Accumulation 🟢)';
                                } else if (change < -5) {
                                    signal = ' (Strong Distribution 🔴)';
                                } else if (change > 0) {
                                    signal = ' (Mild Accumulation)';
                                } else {
                                    signal = ' (Mild Distribution)';
                                }
                            }
                            
                            return [`OBV: ${obvValue.toLocaleString()}${signal}`];
                        },
                        afterLabel: function(context) {
                            const index = context.dataIndex;
                            if (index > 10) {
                                const current = obvData[index];
                                const previous = obvData[index - 10];
                                const change = ((current - previous) / Math.abs(previous)) * 100;
                                
                                if (change > 5) {
                                    return ['', 'Smart money is buying - bullish signal'];
                                } else if (change < -5) {
                                    return ['', 'Smart money is selling - bearish signal'];
                                } else {
                                    return ['', 'Neutral volume trend'];
                                }
                            }
                            return [];
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return value.toLocaleString();
                        }
                    }
                }
            }
        }
    });
    
    debugLog('OBV chart created successfully!');
}

function createCMFChart(ctx, data) {
    debugLog('Creating CMF chart with ' + data.length + ' data points');
    
    // Calculate CMF data
    const cmfData = calculateCMF(data, 20);
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentCMFChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'CMF (20)',
                data: cmfData,
                borderColor: '#f59e0b',
                backgroundColor: 'rgba(245, 158, 11, 0.1)',
                borderWidth: 3,
                fill: true,
                pointRadius: 0,
                pointHoverRadius: 6,
                segment: {
                    borderColor: function(ctx) {
                        const value = ctx.p1.parsed.y;
                        if (value > 0.1) return '#10b981'; // Strong positive - green
                        if (value < -0.1) return '#ef4444'; // Strong negative - red
                        if (value > 0) return '#84cc16'; // Mild positive - light green
                        return '#f97316'; // Mild negative - orange
                    }
                }
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.9)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#f59e0b',
                    borderWidth: 2,
                    cornerRadius: 8,
                    displayColors: false,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const cmfValue = context.parsed.y;
                            let signal = '';
                            
                            if (cmfValue > 0.1) {
                                signal = ' (Strong Buying Pressure 🟢)';
                            } else if (cmfValue < -0.1) {
                                signal = ' (Strong Selling Pressure 🔴)';
                            } else if (cmfValue > 0) {
                                signal = ' (Mild Buying)';
                            } else if (cmfValue < 0) {
                                signal = ' (Mild Selling)';
                            } else {
                                signal = ' (Neutral)';
                            }
                            
                            return [`CMF: ${cmfValue.toFixed(4)}${signal}`];
                        },
                        afterLabel: function(context) {
                            const cmfValue = context.parsed.y;
                            if (cmfValue > 0.1) {
                                return ['', 'Institutional accumulation - consider buying'];
                            } else if (cmfValue < -0.1) {
                                return ['', 'Institutional distribution - consider selling'];
                            } else {
                                return ['', 'Normal money flow conditions'];
                            }
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return value.toFixed(3);
                        }
                    }
                }
            }
        }
    });
    
    // Add CMF reference lines plugin
    const cmfLinesPlugin = {
        id: 'cmfLines',
        afterDraw: function(chart) {
            if (chart.canvas.id === 'cmfChart') {
                const ctx = chart.ctx;
                const chartArea = chart.chartArea;
                const yAxis = chart.scales.y;
                
                ctx.save();
                ctx.setLineDash([5, 5]);
                ctx.lineWidth = 2;
                
                // Strong buying line (0.1)
                ctx.strokeStyle = 'rgba(16, 185, 129, 0.8)';
                const y01 = yAxis.getPixelForValue(0.1);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y01);
                ctx.lineTo(chartArea.right, y01);
                ctx.stroke();
                
                // Strong selling line (-0.1)
                ctx.strokeStyle = 'rgba(239, 68, 68, 0.8)';
                const yNeg01 = yAxis.getPixelForValue(-0.1);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, yNeg01);
                ctx.lineTo(chartArea.right, yNeg01);
                ctx.stroke();
                
                // Zero line
                ctx.strokeStyle = 'rgba(160, 160, 160, 0.8)';
                const y0 = yAxis.getPixelForValue(0);
                ctx.beginPath();
                ctx.moveTo(chartArea.left, y0);
                ctx.lineTo(chartArea.right, y0);
                ctx.stroke();
                
                ctx.restore();
            }
        }
    };
    
    Chart.register(cmfLinesPlugin);
    
    debugLog('CMF chart created successfully!');
}

function createATRChart(ctx, data) {
    debugLog('Creating ATR chart with ' + data.length + ' data points');
    
    // Calculate ATR data
    const atrData = calculateATR(data, 14);
    
    const labels = data.map((item, index) => {
        if (index % Math.ceil(data.length / 10) === 0) {
            return item.date.toLocaleDateString();
        }
        return '';
    });
    
    currentATRChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: 'ATR (14)',
                data: atrData,
                borderColor: '#8b5cf6',
                backgroundColor: 'rgba(139, 92, 246, 0.1)',
                borderWidth: 3,
                fill: true,
                pointRadius: 0,
                pointHoverRadius: 6,
                tension: 0.1
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            animation: false,
            interaction: {
                intersect: false,
                mode: 'index'
            },
            plugins: {
                legend: { display: false },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.9)',
                    titleColor: '#ffffff',
                    bodyColor: '#ffffff',
                    borderColor: '#8b5cf6',
                    borderWidth: 2,
                    cornerRadius: 8,
                    displayColors: false,
                    callbacks: {
                        title: function(context) {
                            const index = context[0].dataIndex;
                            return data[index].date.toLocaleDateString();
                        },
                        label: function(context) {
                            const atrValue = context.parsed.y;
                            let signal = '';
                            
                            // Calculate relative volatility
                            const index = context.dataIndex;
                            if (index > 20) {
                                const recent = atrData.slice(index - 20, index);
                                const avgATR = recent.reduce((a, b) => a + b, 0) / recent.length;
                                
                                if (atrValue > avgATR * 1.2) {
                                    signal = ' (High Volatility 🔴)';
                                } else if (atrValue < avgATR * 0.8) {
                                    signal = ' (Low Volatility 🟢)';
                                } else {
                                    signal = ' (Normal Volatility)';
                                }
                            }
                            
                            return [`ATR: ${atrValue.toFixed(4)}${signal}`];
                        },
                        afterLabel: function(context) {
                            const atrValue = context.parsed.y;
                            const index = context.dataIndex;
                            
                            if (index > 20) {
                                const recent = atrData.slice(index - 20, index);
                                const avgATR = recent.reduce((a, b) => a + b, 0) / recent.length;
                                
                                if (atrValue > avgATR * 1.2) {
                                    return ['', 'High risk period - use smaller position sizes'];
                                } else if (atrValue < avgATR * 0.8) {
                                    return ['', 'Low risk period - good for position building'];
                                } else {
                                    return ['', 'Normal market conditions'];
                                }
                            }
                            return [];
                        }
                    }
                }
            },
            scales: {
                x: {
                    type: 'category',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: { color: '#a0a0a0', maxTicksLimit: 6 }
                },
                y: {
                    type: 'linear',
                    grid: { color: 'rgba(255, 255, 255, 0.1)' },
                    ticks: {
                        color: '#a0a0a0',
                        callback: function(value) {
                            return value.toFixed(4);
                        }
                    }
                }
            }
        }
    });
    
    debugLog('ATR chart created successfully!');
}

// Helper function to calculate EMA
function calculateEMA(prices, period) {
    const multiplier = 2 / (period + 1);
    const ema = [prices[0]];
    
    for (let i = 1; i < prices.length; i++) {
        ema[i] = (prices[i] * multiplier) + (ema[i - 1] * (1 - multiplier));
    }
    
    return ema;
}

// Helper function to calculate SMA
function calculateSMA(prices, period) {
    const sma = [];
    
    for (let i = 0; i < prices.length; i++) {
        if (i < period - 1) {
            sma[i] = null; // Not enough data points
        } else {
            const sum = prices.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0);
            sma[i] = sum / period;
        }
    }
    
    return sma;
}

// Helper function to calculate Rolling Standard Deviation
function calculateRollingStd(prices, period) {
    const std = [];
    
    for (let i = 0; i < prices.length; i++) {
        if (i < period - 1) {
            std[i] = null; // Not enough data points
        } else {
            // Calculate mean
            const slice = prices.slice(i - period + 1, i + 1);
            const mean = slice.reduce((a, b) => a + b, 0) / period;
            
            // Calculate variance
            const variance = slice.reduce((acc, price) => {
                const diff = price - mean;
                return acc + (diff * diff);
            }, 0) / period;
            
            // Standard deviation is square root of variance
            std[i] = Math.sqrt(variance);
        }
    }
    
    return std;
}

// Helper function to calculate Parabolic SAR
function calculatePSAR(data) {
    if (data.length < 2) return [];
    
    const psar = [];
    let acceleration = 0.02;
    const maxAcceleration = 0.2;
    let af = acceleration;
    let ep = data[0].high; // Extreme Point
    let isUptrend = true;
    
    psar[0] = data[0].low;
    
    for (let i = 1; i < data.length; i++) {
        // Calculate PSAR
        let newPsar = psar[i - 1] + af * (ep - psar[i - 1]);
        
        if (isUptrend) {
            if (data[i].low < newPsar) {
                // Trend reversal
                isUptrend = false;
                newPsar = ep;
                ep = data[i].low;
                af = acceleration;
            } else {
                if (data[i].high > ep) {
                    ep = data[i].high;
                    af = Math.min(af + acceleration, maxAcceleration);
                }
            }
        } else {
            if (data[i].high > newPsar) {
                // Trend reversal
                isUptrend = true;
                newPsar = ep;
                ep = data[i].high;
                af = acceleration;
            } else {
                if (data[i].low < ep) {
                    ep = data[i].low;
                    af = Math.min(af + acceleration, maxAcceleration);
                }
            }
        }
        
        psar[i] = newPsar;
    }
    
    return psar;
}

// Helper function to calculate OBV
function calculateOBV(data) {
    if (data.length === 0) return [];
    
    const obv = [];
    obv[0] = data[0].volume;
    
    for (let i = 1; i < data.length; i++) {
        if (data[i].close > data[i - 1].close) {
            obv[i] = obv[i - 1] + data[i].volume;
        } else if (data[i].close < data[i - 1].close) {
            obv[i] = obv[i - 1] - data[i].volume;
        } else {
            obv[i] = obv[i - 1];
        }
    }
    
    return obv;
}

// Helper function to calculate CMF
function calculateCMF(data, period = 20) {
    if (data.length < period) return [];
    
    const cmf = [];
    
    for (let i = 0; i < data.length; i++) {
        if (i < period - 1) {
            cmf[i] = 0;
            continue;
        }
        
        let sumMFV = 0;
        let sumVolume = 0;
        
        for (let j = i - period + 1; j <= i; j++) {
            // Money Flow Multiplier = ((Close - Low) - (High - Close)) / (High - Low)
            const high = data[j].high;
            const low = data[j].low;
            const close = data[j].close;
            const volume = data[j].volume;
            
            if (high !== low) {
                const mfm = ((close - low) - (high - close)) / (high - low);
                const mfv = mfm * volume;
                sumMFV += mfv;
            }
            sumVolume += volume;
        }
        
        cmf[i] = sumVolume > 0 ? sumMFV / sumVolume : 0;
    }
    
    return cmf;
}

// Helper function to calculate ATR
function calculateATR(data, period = 14) {
    if (data.length < period) return [];
    
    const tr = []; // True Range
    const atr = [];
    
    // Calculate True Range for each period
    for (let i = 0; i < data.length; i++) {
        if (i === 0) {
            tr[i] = data[i].high - data[i].low;
        } else {
            const hl = data[i].high - data[i].low;
            const hpc = Math.abs(data[i].high - data[i - 1].close);
            const lpc = Math.abs(data[i].low - data[i - 1].close);
            tr[i] = Math.max(hl, hpc, lpc);
        }
    }
    
    // Calculate ATR (Simple Moving Average of True Range)
    for (let i = 0; i < data.length; i++) {
        if (i < period - 1) {
            atr[i] = 0;
        } else {
            const sum = tr.slice(i - period + 1, i + 1).reduce((a, b) => a + b, 0);
            atr[i] = sum / period;
        }
    }
    
    return atr;
}

function toggleVolumeChart() {
    const checkbox = document.getElementById('showVolume');
    const volumeContainer = document.getElementById('volumeChartContainer');
    
    showVolume = checkbox.checked;
    
    if (showVolume) {
        volumeContainer.classList.remove('hidden');
        debugLog('Volume chart enabled');
    } else {
        volumeContainer.classList.add('hidden');
        // Destroy volume chart when hidden
        if (currentVolumeChart) {
            currentVolumeChart.destroy();
            currentVolumeChart = null;
        }
        debugLog('Volume chart disabled');
    }
    
    // Update charts if we have data
    if (selectedTicker && priceData[selectedTicker]) {
        updateChart();
    }
}

function toggleIndicator(indicatorType) {
    debugLog(`=== toggleIndicator('${indicatorType}') called ===`);
    
    const checkbox = document.getElementById(`show${indicatorType}`);
    if (!checkbox) {
        debugLog(`ERROR: Checkbox with id 'show${indicatorType}' not found!`);
        return;
    }
    
    activeIndicators[indicatorType] = checkbox.checked;
    debugLog(`${indicatorType} indicator state: ${activeIndicators[indicatorType]}`);
    
    // Handle legacy MA calls by redirecting to SMA
    if (indicatorType === 'MA') {
        indicatorType = 'SMA';
        const smaCheckbox = document.getElementById('showSMA');
        if (smaCheckbox) {
            smaCheckbox.checked = checkbox.checked;
        }
    }

    const containerMap = {
        'RSI': 'rsiChartContainer',
        'MACD': 'macdChartContainer', 
        'STOCH': 'stochasticChartContainer',
        'OBV': 'obvChartContainer',
        'CMF': 'cmfChartContainer',
        'ATR': 'atrChartContainer',
        'SMA': null, // Simple Moving averages are overlays on price chart
        'EMA': null, // Exponential Moving averages are overlays on price chart
        'BB': null, // Bollinger Bands are overlays on price chart
        'PSAR': null // Parabolic SAR are overlays on price chart
    };
    
    const containerId = containerMap[indicatorType];
    
    if (containerId) {
        const container = document.getElementById(containerId);
        if (container) {
            if (activeIndicators[indicatorType]) {
                container.classList.remove('hidden');
                debugLog(`${indicatorType} container shown`);
            } else {
                container.classList.add('hidden');
                // Destroy the specific indicator chart
                switch(indicatorType) {
                    case 'RSI':
                        if (currentRSIChart) {
                            currentRSIChart.destroy();
                            currentRSIChart = null;
                        }
                        break;
                    case 'MACD':
                        if (currentMACDChart) {
                            currentMACDChart.destroy();
                            currentMACDChart = null;
                        }
                        break;
                    case 'STOCH':
                        if (currentStochasticChart) {
                            currentStochasticChart.destroy();
                            currentStochasticChart = null;
                        }
                        break;
                    case 'OBV':
                        if (currentOBVChart) {
                            currentOBVChart.destroy();
                            currentOBVChart = null;
                        }
                        break;
                    case 'CMF':
                        if (currentCMFChart) {
                            currentCMFChart.destroy();
                            currentCMFChart = null;
                        }
                        break;
                    case 'ATR':
                        if (currentATRChart) {
                            currentATRChart.destroy();
                            currentATRChart = null;
                        }
                        break;
                }
                debugLog(`${indicatorType} chart destroyed and container hidden`);
            }
        } else {
            debugLog(`ERROR: Container '${containerId}' not found!`);
        }
    } else {
        debugLog(`${indicatorType} is an overlay indicator (no separate container)`);
    }
    
    // Update charts if we have data
    if (selectedTicker && priceData[selectedTicker]) {
        debugLog(`Updating charts for ${selectedTicker}`);
        updateChart();
    } else {
        debugLog('No ticker selected or no price data available');
    }
}

function updateTimeframe() {
    currentTimeframe = document.getElementById('timeframe').value;
    debugLog('Timeframe changed to: ' + currentTimeframe);
    
    if (selectedTicker) {
        updateChart();
    }
}

function updateIndicatorsPanel() {
    const indicatorsGrid = document.getElementById('indicatorsGrid');
    if (!indicatorsGrid || !selectedTicker || !indicatorsData[selectedTicker]) return;
    
    const indicators = indicatorsData[selectedTicker];
    const indicatorItems = [
        { name: 'RSI', key: 'RSI14', format: (val) => val.toFixed(2) },
        { name: 'MACD', key: 'MACD', format: (val) => val.toFixed(4) },
        { name: 'EMA 50', key: 'EMA50', format: (val) => '$' + val.toFixed(2) },
        { name: 'SMA 20', key: 'SMA20', format: (val) => '$' + val.toFixed(2) },
        { name: 'OBV', key: 'OBV', format: (val) => Math.round(val).toLocaleString() },
        { name: 'CMF', key: 'CMF', format: (val) => val.toFixed(4) }
    ];
    
    indicatorsGrid.innerHTML = indicatorItems.map(item => {
        const value = indicators[item.key];
        if (value === undefined) return '';
        
        let sentiment = 'neutral';
        if (item.name === 'RSI') {
            sentiment = value > 70 ? 'bearish' : value < 30 ? 'bullish' : 'neutral';
        } else if (item.name === 'MACD') {
            sentiment = value > 0 ? 'bullish' : 'bearish';
        }
        
        return `
            <div class="indicator-item">
                <div class="indicator-name">${item.name}</div>
                <div class="indicator-value ${sentiment}">${item.format(value)}</div>
            </div>
        `;
    }).join('');
}

function updateTickerInfo(symbol) {
    const ticker = tickersData.find(t => t.symbol === symbol);
    if (!ticker) return;
    
    document.getElementById('selectedTicker').textContent = symbol;
    document.getElementById('currentPrice').textContent = '$' + ticker.price.toFixed(2);
    
    const changeElement = document.getElementById('priceChange');
    const changePercent = ticker.price > 0 ? (ticker.change / ticker.price * 100) : 0;
    changeElement.textContent = `${ticker.change >= 0 ? '+' : ''}${ticker.change.toFixed(2)} (${changePercent.toFixed(2)}%)`;
    changeElement.className = `change ${ticker.change >= 0 ? 'positive' : 'negative'}`;
    
    document.getElementById('volume').textContent = `Volume: ${ticker.volume.toLocaleString()}`;
}

function updateSignals() {
    const signalsList = document.getElementById('signalsList');
    if (!signalsList) return;
    
    // Sample signals based on current strategy
    const sampleSignals = [
        { ticker: 'TASC', strategy: 'OBV Strategy', action: 'buy', strength: 'Strong' },
        { ticker: 'HASH', strategy: 'RSI Strategy2', action: 'buy', strength: 'Weak' },
        { ticker: 'IIEW', strategy: 'OBV Strategy', action: 'hold', strength: '' },
        { ticker: 'BMNS', strategy: 'RSI Strategy', action: 'sell', strength: 'Strong' }
    ];
    
    signalsList.innerHTML = sampleSignals.map(signal => `
        <div class="signal-item ${signal.action}">
            <div class="signal-ticker">${signal.ticker}</div>
            <div class="signal-strategy">${signal.strategy}</div>
            <div class="signal-action ${signal.action}">${signal.strength} ${signal.action.toUpperCase()}</div>
        </div>
    `).join('');
}

function updateMarketStats() {
    document.getElementById('totalTickers').textContent = tickersData.length;
    document.getElementById('activeStrategies').textContent = '12';
}

function updateTopMovers() {
    const topMovers = document.getElementById('topMovers');
    if (!topMovers) return;
    
    const sortedTickers = [...tickersData]
        .filter(ticker => ticker.price > 0)
        .sort((a, b) => Math.abs(b.change / b.price) - Math.abs(a.change / a.price))
        .slice(0, 5);
    
    topMovers.innerHTML = sortedTickers.map(ticker => {
        const changePercent = (ticker.change / ticker.price * 100);
        return `
            <div class="mover-item" onclick="selectTicker('${ticker.symbol}')" style="cursor: pointer;">
                <div class="mover-ticker">${ticker.symbol}</div>
                <div class="mover-change ${changePercent >= 0 ? 'positive' : 'negative'}">
                    ${changePercent >= 0 ? '+' : ''}${changePercent.toFixed(2)}%
                </div>
            </div>
        `;
    }).join('');
}

function updateLastUpdate() {
    const now = new Date();
    document.getElementById('lastUpdate').textContent = now.toLocaleTimeString();
}

// Event handlers
function selectTicker(symbol) {
    console.log('=== selectTicker() called with symbol:', symbol, '===');
    selectedTicker = symbol;
    
    // Update dropdown selection
    const dropdown = document.getElementById('tickerDropdown');
    if (dropdown) {
        dropdown.value = symbol;
        console.log('Updated dropdown value to:', dropdown.value);
    }
    
    // Update selected ticker info display
    updateSelectedTickerInfo(symbol);
    
    // Load ticker data and update chart
    console.log('Loading ticker data for:', symbol);
    loadTickerData(symbol);
}

function selectTickerFromDropdown() {
    debugLog('=== selectTickerFromDropdown() called ===');
    const dropdown = document.getElementById('tickerDropdown');
    const symbol = dropdown.value;
    
    debugLog('Dropdown value: ' + symbol);
    
    if (symbol) {
        debugLog('Valid symbol selected, calling selectTicker()');
        selectTicker(symbol);
    } else {
        debugLog('No symbol selected, clearing chart');
        // Clear selection
        selectedTicker = null;
        clearSelectedTickerInfo();
        clearChart();
    }
}

function updateSelectedTickerInfo(symbol) {
    const ticker = tickersData.find(t => t.symbol === symbol);
    const infoContainer = document.getElementById('selectedTickerInfo');
    
    if (!ticker || !infoContainer) return;
    
    const changePercent = ticker.price > 0 ? (ticker.change / ticker.price * 100) : 0;
    
    infoContainer.innerHTML = `
        <div class="ticker-details">
            <div class="ticker-symbol">${ticker.symbol}</div>
            <div class="ticker-name">${ticker.name || ticker.companyName || ''}</div>
            <div class="ticker-price">$${ticker.price.toFixed(2)}</div>
            <div class="ticker-change ${ticker.change >= 0 ? 'positive' : 'negative'}">
                ${ticker.change >= 0 ? '+' : ''}${ticker.change.toFixed(2)} (${changePercent.toFixed(2)}%)
            </div>
        </div>
    `;
    
    infoContainer.classList.add('active');
    infoContainer.style.display = 'block';
}

function clearSelectedTickerInfo() {
    const infoContainer = document.getElementById('selectedTickerInfo');
    if (infoContainer) {
        infoContainer.classList.remove('active');
        infoContainer.style.display = 'none';
        infoContainer.innerHTML = '';
    }
}

function clearChart() {
    // Clear all charts
    if (currentChart) {
        currentChart.destroy();
        currentChart = null;
    }
    
    if (currentVolumeChart) {
        currentVolumeChart.destroy();
        currentVolumeChart = null;
    }
    
    if (currentRSIChart) {
        currentRSIChart.destroy();
        currentRSIChart = null;
    }
    
    if (currentMACDChart) {
        currentMACDChart.destroy();
        currentMACDChart = null;
    }
    
    if (currentStochasticChart) {
        currentStochasticChart.destroy();
        currentStochasticChart = null;
    }
    
    if (currentOBVChart) {
        currentOBVChart.destroy();
        currentOBVChart = null;
    }
    
    if (currentCMFChart) {
        currentCMFChart.destroy();
        currentCMFChart = null;
    }
    
    if (currentATRChart) {
        currentATRChart.destroy();
        currentATRChart = null;
    }
    
    // Reset chart header
    document.getElementById('selectedTicker').textContent = 'Select a Ticker from the dropdown above';
    
    // Clear data range display
    const rangeElement = document.getElementById('dataRange');
    if (rangeElement) {
        rangeElement.textContent = '';
    }
}

// Function removed - updateTimeframe now defined above

function setupEventListeners() {
    // Chart type switching
    document.querySelectorAll('.chart-type').forEach(button => {
        button.addEventListener('click', function() {
            document.querySelectorAll('.chart-type').forEach(btn => btn.classList.remove('active'));
            this.classList.add('active');
            if (selectedTicker) {
                updateChart();
            }
        });
    });
}

// Utility functions
function filterDataByTimeframe(data, timeframe) {
    if (!data || data.length === 0) return [];
    
    const now = new Date();
    let startDate = new Date();
    
    switch (timeframe) {
        case '7D':
            startDate.setDate(now.getDate() - 7);
            break;
        case '14D':
            startDate.setDate(now.getDate() - 14);
            break;
        case '1M':
            startDate.setMonth(now.getMonth() - 1);
            break;
        case '3M':
            startDate.setMonth(now.getMonth() - 3);
            break;
        case '6M':
            startDate.setMonth(now.getMonth() - 6);
            break;
        case '1Y':
            startDate.setFullYear(now.getFullYear() - 1);
            break;
        case '2Y':
            startDate.setFullYear(now.getFullYear() - 2);
            break;
        case '5Y':
            startDate.setFullYear(now.getFullYear() - 5);
            break;
        case 'ALL':
        default:
            return data;
    }
    
    const filteredData = data.filter(item => item.date >= startDate);
    
    // Update data range display
    updateDataRangeDisplay(filteredData, timeframe);
    
    return filteredData;
}

function updateDataRangeDisplay(data, timeframe) {
    const rangeElement = document.getElementById('dataRange');
    if (!rangeElement || !data || data.length === 0) return;
    
    const startDate = data[0].date.toLocaleDateString();
    const endDate = data[data.length - 1].date.toLocaleDateString();
    const pointCount = data.length;
    
    rangeElement.textContent = `${pointCount} points (${startDate} to ${endDate})`;
}

function generateSamplePriceData(symbol) {
    const data = [];
    const startDate = new Date();
    startDate.setMonth(startDate.getMonth() - 6);
    
    let price = Math.random() * 20 + 5; // Random starting price between 5-25
    
    for (let i = 0; i < 180; i++) {
        const date = new Date(startDate);
        date.setDate(date.getDate() + i);
        
        const change = (Math.random() - 0.5) * 0.5; // Random change
        price = Math.max(1, price + change); // Minimum price of 1
        
        const open = price;
        const high = price + Math.random() * 0.3;
        const low = price - Math.random() * 0.3;
        const close = low + Math.random() * (high - low);
        const volume = Math.floor(Math.random() * 100000) + 10000;
        
        data.push({
            date: date,
            open: open,
            high: high,
            low: Math.max(0.1, low),
            close: close,
            volume: volume
        });
        
        price = close;
    }
    
    return data;
}

function generateSampleIndicators() {
    return {
        RSI14: Math.random() * 100,
        MACD: (Math.random() - 0.5) * 0.1,
        EMA50: Math.random() * 20 + 5,
        SMA20: Math.random() * 20 + 5,
        OBV: Math.random() * 1000000,
        CMF: (Math.random() - 0.5) * 0.5
    };
}

function showLoading(show) {
    const overlay = document.getElementById('loadingOverlay');
    if (overlay) {
        overlay.classList.toggle('show', show);
    }
}

function showStatus(message, isError = false) {
    const statusDiv = document.getElementById('statusMessage');
    if (!statusDiv) return;
    statusDiv.textContent = message;
    statusDiv.style.backgroundColor = isError ? 'var(--danger-color)' : 'var(--success-color)';
    statusDiv.classList.add('show');
    setTimeout(() => statusDiv.classList.remove('show'), 4000);
}

function showError(message) {
    console.error(message);
    showStatus(message, true);
}

// Button handlers
async function refreshData() {
    showLoading(true);
    try {
        const response = await fetch('/api/refresh', { method: 'POST' });
        const data = await response.json();
        if (!response.ok) throw new Error(data.message || 'Refresh failed');

        await loadTickers();
        await loadStrategiesData();
        if (selectedTicker) {
            await loadTickerData(selectedTicker);
        }
        updateLastUpdate();
        showStatus(data.message || 'Data refreshed');
    } catch (error) {
        console.error('Error refreshing data:', error);
        showError('Failed to refresh data');
    } finally {
        showLoading(false);
    }
}

async function runBacktest() {
    showLoading(true);
    try {
        const response = await fetch('/api/backtest', { method: 'POST' });
        const data = await response.json();
        if (!response.ok) throw new Error(data.message || 'Backtest failed');

        // Reload strategies data
        await loadStrategiesData();
        updateSignals();
        showStatus(data.message || 'Backtest started');
    } catch (error) {
        console.error('Error running backtest:', error);
        showError('Failed to run backtest');
    } finally {
        showLoading(false);
    }
}

// Auto-refresh data every 5 minutes
setInterval(refreshData, 5 * 60 * 1000);

 