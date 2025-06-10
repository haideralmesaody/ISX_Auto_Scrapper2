async function init() {
    const tickRes = await fetch('/api/tickers');
    const tickers = await tickRes.json();
    const tSel = document.getElementById('strategyTicker');
    tickers.forEach(t => {
        const opt = document.createElement('option');
        opt.value = t.symbol; opt.textContent = t.symbol;
        tSel.appendChild(opt);
    });
    tSel.addEventListener('change', loadStrategies);
}

async function loadStrategies() {
    const sym = document.getElementById('strategyTicker').value;
    if (!sym) return;
    const res = await fetch(`/api/ticker/${sym}?type=strategies&full=1`);
    const data = await res.json();
    const sSel = document.getElementById('strategySelect');
    sSel.innerHTML = '';
    const strategies = Object.keys(data.signals);
    strategies.forEach(s => {
        const opt = document.createElement('option');
        opt.value = s; opt.textContent = s;
        sSel.appendChild(opt);
    });
    sSel.onchange = () => displaySignal(data, sSel.value);
    if (strategies.length) displaySignal(data, strategies[0]);
}

function displaySignal(data, strategy) {
    const div = document.getElementById('strategyResult');
    const history = data.history || [];
    let html = `<h3>${strategy}</h3><ul>`;
    history.forEach(row => {
        html += `<li>${row.Date} : ${row[strategy.replace(/ /g,'')] || row[strategy]}</li>`;
    });
    html += '</ul>';
    div.innerHTML = html;
}

document.addEventListener('DOMContentLoaded', init);
