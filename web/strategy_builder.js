// Strategy Builder JS

// --- Fake indicator list (Highcharts Stock indicators). Replace with API call later ---
const INDICATORS = [
  // Moving averages & trend
  { id: "SMA", type: "sma", params: [5,10,20,50,100,200], paramCount:1 },
  { id: "EMA", type: "ema", params: [5,10,20,50,100,200], paramCount:1 },
  { id: "WMA", type: "wma", params: [10,20,50,100], paramCount:1 },
  { id: "DEMA", type: "dema", params: [20], paramCount:1 },
  { id: "TEMA", type: "tema", params: [20], paramCount:1 },
  { id: "TRIX", type: "trix", params: [14], paramCount:1 },
  { id: "SuperTrend", type: "supertrend", params: [10,3], paramCount:2, paramNames:["Period","Multiplier"] },
  { id: "PSAR", type: "psar", params: [0.02,0.2], paramCount:2, paramNames:["Step","Max"] },

  // Volatility & bands
  { id: "ATR", type: "atr", params: [14], paramCount:1 },
  { id: "Bollinger", type: "bb", params: [20,2], paramCount:2, paramNames:["Period","StdDev"] },
  { id: "Keltner", type: "keltnerchannels", params: [20], paramCount:1 },
  { id: "PriceEnvelopes", type: "priceenvelopes", params: [20,2], paramCount:2, paramNames:["Period","Percentage"] },
  { id: "RollingStd", type: "standarddeviation", params: [10,50], paramCount:1 },

  // Oscillators
  { id: "RSI", type: "rsi", params: [7,9,14,25], paramCount:1 },
  { id: "Stochastic", type: "stoch", params: [14,3], paramCount:2, paramNames:["%K","%D"] },
  { id: "MACD", type: "macd", params: [12,26,9], paramCount:3, paramNames:["Fast","Slow","Signal"] },
  { id: "Momentum", type: "momentum", params: [10], paramCount:1 },
  { id: "ROC", type: "roc", params: [10], paramCount:1 },
  { id: "WilliamR", type: "williamsr", params: [14], paramCount:1 },
  { id: "CCI", type: "cci", params: [20], paramCount:1 },
  { id: "Aroon", type: "aroon", params: [14], paramCount:1 },
  { id: "APO", type: "apo", params: [12,26], paramCount:2, paramNames:["Short","Long"] },
  { id: "DPO", type: "dpo", params: [20], paramCount:1 },
  { id: "CMF", type: "cmf", params: [20], paramCount:1 },
  { id: "OBV", type: "obv", params: [], paramCount:0 },
  { id: "OBV_RoC", type: "obv_roc", params: [10], paramCount:1 },
  { id: "ChaikinOsc", type: "chaikin", params: [3,10], paramCount:2, paramNames:["Short","Long"] },

  // Patterns & misc
  { id: "PivotPoints", type: "pivotpoints", params: [], paramCount:0 },
  { id: "ZigZag", type: "zigzag", params: [1], paramCount:1 },
  { id: "Ichimoku", type: "ichimoku", params: [9,26,52], paramCount:3, paramNames:["Tenkan","Kijun","Senkou"] }
];

const OPERATORS = [">", "<", ">=", "<=", "==", "!="];

// Constants for dropdown special actions
const ACTION_NEW = "__new__";
const ACTION_DELETE = "__delete__";

// Local in-memory list of strategies (replace later with backend data)
let strategiesCache = [];

// Chart globals
let builderChart = null;
let selectedTicker = "";

// Ensure a dedicated indicator axis exists (below price, above volume)
function ensureIndicatorAxis() {
  if (!builderChart) return 0;
  const existing = builderChart.yAxis.find((a) => a.options && a.options.id === "indicatorAxis");
  if (existing) {
    if (!existing.visible) existing.update({ visible: true }, false);
    return "indicatorAxis";
  }

  const axisIdx = builderChart.yAxis.length;
  builderChart.addAxis(
    {
      id: "indicatorAxis",
      top: "60%",
      height: "15%",
      offset: 0,
      lineWidth: 1,
      labels: { align: "right", x: -3 },
      title: { text: "Indicator" },
    },
    false,
    false
  );
  // Make sure volume axis stays at 75% top if it was pushed down by axis addition
  if (builderChart.yAxis[1]) {
    builderChart.yAxis[1].update({ top: "75%", height: "25%" }, false);
  }
  return "indicatorAxis";
}

function addIndicatorToChart(indType, params) {
  if (!builderChart) return;
  // Avoid duplicates: use type+period id
  const uid = `${indType}_${params}`;
  if (builderChart.series.find((s) => s.options.id === uid)) return;
  try {
    const overlayTypes = [
      "sma",
      "ema",
      "wma",
      "dema",
      "tema",
      "bb",
      "supertrend",
      "psar",
      "pivotpoints",
      "zigzag",
      "keltnerchannels",
    ];

    const indOptions = {
      id: uid,
      type: indType,
      linkedTo: "main",
      dataGrouping: { enabled: false },
      showInLegend: true,
    };

    if (!overlayTypes.includes(indType)) {
      // oscillator or separate-indicator: use dedicated axis id
      indOptions.yAxis = ensureIndicatorAxis();
    }

    if (params && params !== "" && !isNaN(Number(params))) {
      indOptions.params = { period: Number(params) };
    }
    builderChart.addSeries(indOptions);
  } catch (err) {
    console.warn("Unable to add indicator", indType, err.message);
  }
}

function clearIndicatorsFromChart() {
  if (!builderChart) return;
  // Remove all series except main and volume (id main/volume)
  [...builderChart.series].forEach((s) => {
    if (s.options.id !== "main" && s.options.id !== "volume") s.remove(false);
  });
  builderChart.redraw();
}

function updateChartIndicators() {
  clearIndicatorsFromChart();
  const rows = document.querySelectorAll("#buyRulesTable tbody tr, #sellRulesTable tbody tr");
  rows.forEach((tr) => {
    const indSel = tr.querySelector("td:nth-child(1) select");
    const paramSel = tr.querySelector("td:nth-child(2) select, td:nth-child(2) input, td:nth-child(2) span");
    if (!indSel) return;
    const indObj = INDICATORS.find((i) => i.id === indSel.value);
    if (!indObj) return;
    const paramVal = paramSel ? (paramSel.value || paramSel.textContent || "") : undefined;
    const indType = indObj.type || indObj.id.toLowerCase();
    addIndicatorToChart(indType, paramVal);

    // process target indicator
    const tgtSel = tr.querySelector("td:nth-child(4) select");
    if (tgtSel && tgtSel.value && !["Value","Open","Close","High","Low"].includes(tgtSel.value)) {
      const tgtMeta = INDICATORS.find((i)=> i.id === tgtSel.value);
      if(tgtMeta){
        const valTd = tr.children[4];
        const pEl = valTd ? valTd.querySelector("select,input,span") : null;
        const tgtParamVal = pEl ? (pEl.value || pEl.textContent || "") : undefined;
        const tgtType = tgtMeta.type || tgtMeta.id.toLowerCase();
        addIndicatorToChart(tgtType, tgtParamVal);
      }
    }
  });
  highlightZones();
  renderLogic("buyLogic", buyRules());
  renderLogic("sellLogic", sellRules());

  builderChart && builderChart.redraw();
}

// Attach both change & input handlers so any element updates redraw
function attachUpdate(el){
  if(!el) return;
  ["change","input"].forEach(evt=>el.addEventListener(evt, updateChartIndicators));
}

// Replace existing wrapper definition
const _origAddRuleRow = addRuleRow;
addRuleRow = function (tableBody, r = {}) {
  _origAddRuleRow(tableBody, r);
  const lastRow = tableBody.lastElementChild;
  if (lastRow) {
    lastRow.querySelectorAll("select,input").forEach(attachUpdate);
    const del = lastRow.querySelector("button");
    if (del) del.addEventListener("click", () => { updateChartIndicators(); builderChart && builderChart.redraw();});
  }
  updateChartIndicators();
  builderChart && builderChart.redraw();
};

function attachRowListenersExist() {
  document.querySelectorAll("#buyRulesTable tbody tr select, #sellRulesTable tbody tr select").forEach((s) => s.addEventListener("change", updateChartIndicators));
}

async function fetchTickersList() {
  try {
    const res = await fetch("/api/tickers");
    return await res.json();
  } catch {
    return [];
  }
}

async function createBuilderChart(rawSymbol) {
  const symbol = rawSymbol.trim().split(/[\s-]+/)[0]; // take first token before space or dash
  console.log("[Builder] Loading chart for", symbol);
  selectedTicker = symbol;
  const container = document.getElementById("builderChart");

  // Clean up previous chart safely
  if (builderChart) {
    try {
      builderChart.destroy();
    } catch (cleanupErr) {
      console.warn("Highcharts destroy error", cleanupErr);
    }
    builderChart = null;
  }

  container.innerHTML = "Loading...";

  try {
    const resp = await fetch(`/api/ticker/${encodeURIComponent(symbol)}?type=price`);
    if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
    const data = await resp.json();
    if (!Array.isArray(data) || data.length === 0) throw new Error("No data");

    const ohlc = [];
    const volume = [];
    data.forEach((d) => {
      const ts = d.timestamp ? d.timestamp : Date.parse(d.date);
      ohlc.push([ts, d.open, d.high, d.low, d.close]);
      volume.push([ts, d.volume || 0]);
    });

    builderChart = Highcharts.stockChart("builderChart", {
      yAxis: [
        {
          labels: { align: "right", x: -3 },
          title: { text: "Price" },
          top: "0%",
          height: "55%",
          lineWidth: 1,
          resize: { enabled: true },
        },
        {
          id: "indicatorAxis", // placeholder for overlays created later
          labels: { align: "right", x: -3 },
          title: { text: "Indicator" },
          top: "55%",
          height: "20%",
          offset: 0,
          lineWidth: 1,
          visible: false,
        },
        {
          labels: { align: "right", x: -3 },
          title: { text: "Volume" },
          top: "75%",
          height: "25%",
          offset: 0,
          lineWidth: 1,
        },
      ],
      colors: ["#2d5016", "#6b9b37", "#FF6F6F", "#8bc34a", "#4a7c23"],
      chart: { backgroundColor: "rgba(0,0,0,0)", plotBackgroundColor: "rgba(0,0,0,0)" },
      series: [
        { id: "main", type: "candlestick", color: "#FF6F6F", upColor: "#6FB76F", data: ohlc, dataGrouping: { enabled: false } },
        { id: "volume", type: "column", data: volume, yAxis: 2, dataGrouping: { enabled: false } },
      ],
      rangeSelector: { selected: 1 },
      navigator: { enabled: true },
      scrollbar: { enabled: true },
      stockTools: { gui: { enabled: true } },
      title: { text: `${symbol} – Strategy Preview` },
      credits: { enabled: false },
    });

    // Apply indicators only after chart is live
    updateChartIndicators();
    attachRowListenersExist();
    // Recalculate zones on range changes
    if (builderChart && builderChart.xAxis[0]) {
      Highcharts.addEvent(builderChart.xAxis[0], "afterSetExtremes", () => highlightZones());
    }
  } catch (err) {
    container.innerHTML = `<p style='color:#e74c3c;text-align:center;padding:1rem;'>${err.message}</p>`;
    console.error("[Builder] Chart error", err);
  }
}

// --- DOM helpers ---
function el(tag, attrs = {}, children = []) {
  const e = document.createElement(tag);
  Object.entries(attrs).forEach(([k, v]) => (e[k] = v));
  children.forEach((c) => e.appendChild(typeof c === "string" ? document.createTextNode(c) : c));
  return e;
}

function indicatorSelect() {
  const select = el("select");
  INDICATORS.forEach((ind) => {
    const opt = el("option", { value: ind.id, innerText: ind.id });
    select.appendChild(opt);
  });
  return select;
}

function paramElement(indId, current="") {
  const meta = INDICATORS.find((i) => i.id === indId) || { params: [], paramCount: 0 };

  if (meta.paramCount === 0) {
    // no params – return a dash placeholder span
    return el("span", { innerText: "—" });
  }

  // Single-param → dropdown
  if (meta.paramCount === 1) {
    const sel = el("select");
    meta.params.forEach((p) => sel.appendChild(el("option", { value: p, innerText: p })));
    if (current) sel.value = current;
    return sel;
  }

  // Multi-param → free text input (comma-sep)
  const inp = el("input", {
    type: "text",
    style: "width:120px;",
    value: current || meta.params.join(","),
    placeholder: (meta.paramNames || []).join(",") || "params",
  });
  return inp;
}

// Back-compat alias so rest of code need minimal updates
const paramSelect = paramElement;

function operatorSelect() {
  const s = el("select");
  OPERATORS.forEach((op) => s.appendChild(el("option", { value: op, innerText: op })));
  return s;
}

function targetSelect() {
  const s = el("select");

  // Price fields group
  const priceGroup = el("optgroup", { label: "Price" });
  ["Open", "Close", "High", "Low"].forEach((t) => priceGroup.appendChild(el("option", { value: t, innerText: t })));
  s.appendChild(priceGroup);

  // Indicators group (only IDs – period will be chosen separately)
  const indGroup = el("optgroup", { label: "Indicators" });
  INDICATORS.forEach((ind) => {
    indGroup.appendChild(el("option", { value: ind.id, innerText: ind.id }));
  });

  // Numeric value placeholder
  s.appendChild(el("option", { value: "Value", innerText: "(Numeric Value)" }));

  s.appendChild(indGroup);

  return s;
}

function addRuleRow(tableBody, r = {}) {
  const tr = el("tr");
  const indSel = indicatorSelect();
  const paramSel = paramSelect(indSel.value);
  indSel.addEventListener("change", () => {
    const newEl = paramSelect(indSel.value);
    tr.children[1].replaceChild(newEl, tr.children[1].firstChild);
    attachUpdate(newEl);
    updateChartIndicators();
  });

  // Initial listeners
  attachUpdate(indSel);
  attachUpdate(paramSel);

  tr.appendChild(el("td", {}, [indSel]));
  tr.appendChild(el("td", {}, [paramSel]));
  tr.appendChild(el("td", {}, [operatorSelect()]));

  // Target selector
  const targetTd = el("td");
  const valueTd = el("td");
  const tgtSel = targetSelect();
  targetTd.appendChild(tgtSel);

  const valInput = el("input", { type: "number", step: "any", style: "width:80px; display:none;" });
  valueTd.appendChild(valInput);

  if (typeof r.target === "number") {
    tgtSel.value = "Value";
    valInput.value = r.target;
    valInput.style.display = "inline-block";
  } else if (r.target !== undefined) {
    tgtSel.value = r.target;
    valInput.style.display = "none";
  }

  const handleTargetChange = () => {
    // reset valueTd
    valueTd.innerHTML = "";
    if (tgtSel.value === "Value") {
      valueTd.appendChild(valInput);
      valInput.style.display = "inline-block";
    } else if (["Open","Close","High","Low"].includes(tgtSel.value)) {
      // nothing extra
    } else {
      // indicator – need its params
      const paramEl = paramSelect(tgtSel.value);
      valueTd.appendChild(paramEl);
      attachUpdate(paramEl);
    }
    updateChartIndicators();
  };
  tgtSel.addEventListener("change", handleTargetChange);
  // run once for default state
  handleTargetChange();

  tr.appendChild(targetTd);
  tr.appendChild(valueTd);

  const delBtn = el("button", { innerText: "✕", className: "btn btn-secondary btn-small", style: "position:relative;z-index:1;" });
  delBtn.addEventListener("click", () => {tr.remove(); updateChartIndicators();});
  const btnTd = el("td", { style: "min-width:44px;" }, [delBtn]);
  tr.appendChild(btnTd);

  const linkTd = el("td");
  const linkSel = el("select");
  ["AND","OR"].forEach(op=>linkSel.appendChild(el("option",{value:op,innerText:op})));
  if(r.link) linkSel.value=r.link;
  linkTd.appendChild(linkSel);
  tr.appendChild(linkTd);

  // Hide link dropdown for first row
  if (tableBody.querySelectorAll("tr").length === 0) {
    linkSel.style.visibility = "hidden";
  }

  tableBody.appendChild(tr);
}

// Strategy CRUD stubs
async function loadStrategies() {
  // TODO: replace with backend GET
  try {
    const list = JSON.parse(localStorage.getItem("strategies") || "[]");
    strategiesCache = list;
    return list;
  } catch {
    return [];
  }
}

async function persistStrategies() {
  // TODO: replace with backend POST/PUT
  localStorage.setItem("strategies", JSON.stringify(strategiesCache));
}

async function saveStrategy(strat) {
  const idx = strategiesCache.findIndex((s) => s.id === strat.id);
  if (idx >= 0) {
    strategiesCache[idx] = strat; // update
  } else {
    strategiesCache.push(strat); // add new
  }
  await persistStrategies();
}

async function deleteStrategy(id) {
  strategiesCache = strategiesCache.filter((s) => s.id !== id);
  await persistStrategies();
}

function collectRules(tbody) {
  const rules = [];
  tbody.querySelectorAll("tr").forEach((tr) => {
    const selects = tr.querySelectorAll("select");
    if (selects.length < 5) return; // skip incomplete rows
    const indSel = selects[0];
    const paramEl = tr.querySelector("td:nth-child(2) select, td:nth-child(2) input, td:nth-child(2) span");
    const opSel = selects[2], targetSel = selects[3], linkSel = selects[4];
    let paramValRaw = paramEl ? (paramEl.value || paramEl.textContent || "") : "";
    if (paramValRaw === "—") paramValRaw = "";
    const numInput = tr.querySelector("input[type='number']");
    let targetVal;
    if (targetSel.value === "Value") {
      targetVal = Number(numInput.value || "");
    } else if (["Open","Close","High","Low"].includes(targetSel.value)) {
      targetVal = targetSel.value;
    } else {
      // indicator target – fetch its parameter element inside valueTd
      const valTd = tr.children[4];
      const pEl = valTd ? valTd.querySelector("select,input,span") : null;
      const pVal = pEl ? (pEl.value || pEl.textContent || "") : "";
      targetVal = targetSel.value + (pVal ? `_${pVal}` : "");
    }
    rules.push({ link: linkSel ? linkSel.value : null, indicator: indSel.value + (paramValRaw ? `_${paramValRaw}` : ""), operator: opSel.value, target: targetVal });
  });
  return rules;
}

function clearRulesTables() {
  document.querySelectorAll("#buyRulesTable tbody, #sellRulesTable tbody").forEach((tbody) => (tbody.innerHTML = ""));
  updateChartIndicators();
}

function populateRules(tbody, rules) {
  rules.forEach((r) => {
    const tr = document.createElement("tr");

    const indSel = indicatorSelect();
    indSel.value = r.indicator.split("_")[0];

    const paramVal = r.indicator.split("_")[1] || "";
    const paramSel = paramSelect(indSel.value, paramVal);

    tr.appendChild(el("td", {}, [indSel]));
    tr.appendChild(el("td", {}, [paramSel]));
    attachUpdate(indSel);
    attachUpdate(paramSel);

    const opSel = operatorSelect();
    opSel.value = r.operator;
    tr.appendChild(el("td", {}, [opSel]));

    const targetTd = el("td");
    const valueTd = el("td");
    const tgtSel = targetSelect();
    targetTd.appendChild(tgtSel);

    const valInput = el("input", { type: "number", step: "any", style: "width:80px; display:none;" });
    valueTd.appendChild(valInput);

    if (typeof r.target === "number") {
      tgtSel.value = "Value";
      valInput.value = r.target;
      valInput.style.display = "inline-block";
    } else if (r.target !== undefined) {
      const parts = String(r.target).split("_");
      tgtSel.value = parts[0];
      valInput.style.display = "none";
      // if param exists we will set it after control created in handleTargetChange2
    }

    const handleTargetChange2 = () => {
      valueTd.innerHTML = "";
      if (tgtSel.value === "Value") {
        valueTd.appendChild(valInput);
        valInput.style.display = "inline-block";
      } else if (["Open","Close","High","Low"].includes(tgtSel.value)) {
        // nothing extra
      } else {
        const paramEl = paramSelect(tgtSel.value);
        valueTd.appendChild(paramEl);
        attachUpdate(paramEl);
        // if original target had param, set it
        if(r.target && typeof r.target!="number"){
          const seg=r.target.split("_")[1]||"";
          if(seg){
             if(paramEl.tagName==="SELECT") paramEl.value=seg;
             else if(paramEl.tagName==="INPUT") paramEl.value=seg;
          }
        }
      }
      updateChartIndicators();
    };
    tgtSel.addEventListener("change", handleTargetChange2);
    handleTargetChange2();
    attachUpdate(tgtSel);

    tr.appendChild(targetTd);
    tr.appendChild(valueTd);

    const delBtn = el("button", { innerText: "✕", className: "btn btn-secondary btn-small" });
    delBtn.addEventListener("click", () => {tr.remove(); updateChartIndicators();});
    tr.appendChild(el("td", {}, [delBtn]));

    const linkTd = el("td");
    const linkSel = el("select");
    ["AND","OR"].forEach(op=>linkSel.appendChild(el("option",{value:op,innerText:op})));
    if(r.link) linkSel.value=r.link;
    linkTd.appendChild(linkSel);
    tr.appendChild(linkTd);

    tbody.appendChild(tr);
  });
}

// --- Highlight buy/sell/hold zones as plotBands ---
function evalCondition(value, operator, target) {
  switch (operator) {
    case ">":
      return value > target;
    case "<":
      return value < target;
    case ">=":
      return value >= target;
    case "<=":
      return value <= target;
    case "==":
      return value === target;
    case "!=":
      return value !== target;
    default:
      return false;
  }
}

function highlightZones(){
  if(!builderChart) return;

  const buy = buyRules()[0];
  const sell = sellRules()[0];
  if(!buy||!sell) return;

  function parseInd(str){
    const [id,...rest]=str.split("_");
    return {id,param:rest.join("_")};
  }

  const baseMeta = INDICATORS.find(i=>i.id===buy.indicator.split("_")[0]);
  if(!baseMeta) return;
  const baseType = baseMeta.type||baseMeta.id.toLowerCase();

  // helper to get series points by indicator string "EMA_20"
  function getSeriesData(indStr){
    const {id,param}=parseInd(indStr);
    const meta = INDICATORS.find(i=>i.id===id);
    if(!meta) return null;
    const type = meta.type||meta.id.toLowerCase();
    const uid=`${type}_${param}`;
    const s = builderChart.series.find(se=>se.options.id===uid);
    return s ? s.points : null;
  }

  let mainPts = getSeriesData(buy.indicator);
  if(!mainPts||mainPts.length===0) return;

  // Determine comparator arrays/values flags BEFORE used
  const buyTargetIsNum = !isNaN(Number(buy.target));
  const sellTargetIsNum = !isNaN(Number(sell.target));

  let buyTargetPts = buyTargetIsNum ? null : getSeriesData(buy.target);
  let sellTargetPts = sellTargetIsNum ? null : getSeriesData(sell.target);

  // Skip candles before both comparator values exist (start zones when comparison is evaluable)
  let startIdx = mainPts.findIndex((pt, i)=>{
      const buyReady  = buyTargetIsNum  ? true : (buyTargetPts && buyTargetPts[i] !== undefined);
      const sellReady = sellTargetIsNum ? true : (sellTargetPts && sellTargetPts[i] !== undefined);
      return buyReady && sellReady;
  });
  if(startIdx===-1) return; // never comparable in visible range
  if(startIdx>0){
     // trim arrays to comparable part
     mainPts = mainPts.slice(startIdx);
     if(buyTargetPts)  buyTargetPts  = buyTargetPts.slice(startIdx);
     if(sellTargetPts) sellTargetPts = sellTargetPts.slice(startIdx);
  }

  if((!buyTargetIsNum && !buyTargetPts)||(!sellTargetIsNum && !sellTargetPts)) return;

  const xAxis = builderChart.xAxis[0];
  (xAxis.plotLinesAndBands||[]).forEach(b=>{ if(b.id&&b.id.startsWith("zone-")) b.destroy();});

  // Always draw a baseline neutral band across current visible range
  const xMin = xAxis.min ?? builderChart.xAxis[0].dataMin;
  const xMax = xAxis.max ?? builderChart.xAxis[0].dataMax;
  xAxis.addPlotBand({
     id: 'zone-base-hold',
     from: xMin,
     to: xMax,
     color: 'rgba(255,235,59,0.08)',
     zIndex: 1
  });

  const segments=[];
  let curr="hold";
  let segStart=mainPts[0].x;

  mainPts.forEach((p,idx)=>{
    const val=p.y;
    const compareBuy = buyTargetIsNum ? Number(buy.target) : (buyTargetPts[idx]?buyTargetPts[idx].y:undefined);
    const compareSell = sellTargetIsNum ? Number(sell.target) : (sellTargetPts[idx]?sellTargetPts[idx].y:undefined);
    if(compareBuy===undefined||compareSell===undefined) return;

    let state="hold";
    if(evalCondition(val,buy.operator,compareBuy)) state="buy";
    else if(evalCondition(val,sell.operator,compareSell)) state="sell";

    if(state!==curr){
      segments.push({from:segStart,to:p.x,state:curr});
      segStart=p.x;
      curr=state;
    }
    if(idx===mainPts.length-1){
      segments.push({from:segStart,to:p.x,state:curr});
    }
  });

  if(segments.length===0){
    segments.push({from:mainPts[0].x, to: mainPts[mainPts.length-1].x, state: "hold"});
  }

  const colors={buy:"rgba(76,175,80,0.25)",sell:"rgba(244,67,54,0.25)",hold:"rgba(255,235,59,0.18)"};
  segments.forEach((seg,i)=>{ xAxis.addPlotBand({id:`zone-${i}`,from:seg.from,to:seg.to,color:colors[seg.state]||colors.hold,zIndex:2});});
}

// Helper to extract rule arrays for each panel
function buyRules(){return extractRules(document.querySelectorAll("#buyRulesTable tbody tr"));}
function sellRules(){return extractRules(document.querySelectorAll("#sellRulesTable tbody tr"));}
function extractRules(rows){
  return Array.from(rows).map(tr=>{
    const selects = tr.querySelectorAll("select");
    if (selects.length < 5) return null;
    const indSel = selects[0];
    const paramEl = tr.querySelector("td:nth-child(2) select, td:nth-child(2) input, td:nth-child(2) span");
    const opSel = selects[2], targetSel = selects[3], linkSel = selects[4];
    const input=tr.querySelector("input[type='number']");
    let paramValRaw = paramEl ? (paramEl.value || paramEl.textContent || "") : "";
    if (paramValRaw === "—") paramValRaw = "";
    return {
      link: linkSel ? linkSel.value : null,
      indicator: indSel.value,
      period: paramValRaw,
      operator: opSel.value,
      target: (()=>{
        if(targetSel.value === "Value") return input.value;
        if(["Open","Close","High","Low"].includes(targetSel.value)) return targetSel.value;
        const valTd = tr.children[4];
        const pEl = valTd ? valTd.querySelector("select,input,span") : null;
        const pVal = pEl ? (pEl.value || pEl.textContent || "") : "";
        return targetSel.value + (pVal ? `_${pVal}` : "");
      })()
    };
  }).filter(Boolean);
}

// Render simple SVG logic diagram
function renderLogic(svgId,rules){
  const svg=document.getElementById(svgId);
  if(!svg)return;
  svg.innerHTML="";
  const h=40,vGap=20, rectW=260;
  rules.forEach((r,i)=>{
    const y=i*(h+vGap)+10;
    const rect=document.createElementNS("http://www.w3.org/2000/svg","rect");
    rect.setAttribute("x",10);rect.setAttribute("y",y);
    rect.setAttribute("width",rectW);rect.setAttribute("height",h);
    rect.setAttribute("rx",6);rect.setAttribute("ry",6);
    rect.setAttribute("fill","#fff");rect.setAttribute("stroke","#999");
    svg.appendChild(rect);
    const txt=document.createElementNS("http://www.w3.org/2000/svg","text");
    txt.setAttribute("x",20);txt.setAttribute("y",y+25);txt.setAttribute("font-size","12");
    txt.textContent=`${r.indicator} ${r.operator} ${r.target}`;
    svg.appendChild(txt);
    if(i>0){
      const gate=document.createElementNS("http://www.w3.org/2000/svg","text");
      gate.setAttribute("x",rectW+25);gate.setAttribute("y",y+25);
      gate.setAttribute("font-size","14");gate.textContent=r.link==="OR"?"∨":"∧";
      svg.appendChild(gate);
      const line=document.createElementNS("http://www.w3.org/2000/svg","line");
      line.setAttribute("x1",rectW+10);line.setAttribute("y1",y-(vGap)+h);
      line.setAttribute("x2",rectW+10);line.setAttribute("y2",y);
      line.setAttribute("stroke","#666");
      svg.appendChild(line);
    }
  });
  svg.setAttribute("height", rules.length*(h+vGap)+10);
}

window.addEventListener("DOMContentLoaded", async () => {
  const buyBody = document.querySelector("#buyRulesTable tbody");
  const sellBody = document.querySelector("#sellRulesTable tbody");

  document.getElementById("addBuyRule").addEventListener("click", () => addRuleRow(buyBody));
  document.getElementById("addSellRule").addEventListener("click", () => addRuleRow(sellBody));

  const sel = document.getElementById("strategySelect");

  const delBtn = document.getElementById("delStrategyBtn");

  function toggleDeleteVisibility() {
    if (sel.value && sel.value !== ACTION_NEW && sel.value !== ACTION_DELETE && sel.value !== "") {
      delBtn.style.display = "inline-block";
    } else {
      delBtn.style.display = "none";
    }
  }

  delBtn.addEventListener("click", async () => {
    const currentId = sel.value;
    if (!currentId || currentId === ACTION_NEW) return;
    if (confirm("Delete selected strategy?")) {
      await deleteStrategy(currentId);
      clearRulesTables();
      document.getElementById("statusMsg").innerText = "Deleted ✔";
      refreshDropdown();
    }
  });

  function refreshDropdown() {
    sel.innerHTML = "";
    // Placeholder
    sel.appendChild(el("option", { value: "", innerText: "-- Select strategy --", disabled: true, selected: true }));

    strategiesCache.forEach((s) => sel.appendChild(el("option", { value: s.id, innerText: s.name })));

    // Optgroup with actions
    const actionsGroup = el("optgroup", { label: "Actions" });
    actionsGroup.appendChild(el("option", { value: ACTION_NEW, innerText: "➕ New strategy…" }));
    if (strategiesCache.length) {
      actionsGroup.appendChild(el("option", { value: ACTION_DELETE, innerText: "🗑 Delete selected" }));
    }
    sel.appendChild(actionsGroup);

    toggleDeleteVisibility();
  }

  // Handle dropdown change
  sel.addEventListener("change", async () => {
    const val = sel.value;
    toggleDeleteVisibility();
    if (val === ACTION_NEW) {
      const name = prompt("Enter new strategy name:");
      if (!name) {
        sel.selectedIndex = 0;
        return;
      }
      clearRulesTables();
      addRuleRow(buyBody); // start with one empty row
      const newStrat = { id: name.toLowerCase().replace(/\s+/g, "_"), name, buy_rules: [], sell_rules: [] };
      await saveStrategy(newStrat);
      refreshDropdown();
      sel.value = newStrat.id;
      sel.dispatchEvent(new Event("change"));
      document.getElementById("statusMsg").innerText = "New strategy created – don't forget to Save";
    } else if (val === ACTION_DELETE) {
      const currentId = sel.options[sel.selectedIndex - 1]?.value; // previous option before Actions group
      if (!currentId || [ACTION_NEW, ACTION_DELETE].includes(currentId)) {
        alert("Select a strategy first to delete");
        return;
      }
      if (confirm("Delete selected strategy?")) {
        await deleteStrategy(currentId);
        clearRulesTables();
        document.getElementById("statusMsg").innerText = "Deleted ✔";
        refreshDropdown();
        sel.selectedIndex = 0;
      }
    } else {
      // Load existing strategy rules
      const strat = strategiesCache.find((s) => s.id === val);
      clearRulesTables();
      if (strat) {
        populateRules(buyBody, strat.buy_rules || []);
        populateRules(sellBody, strat.sell_rules || []);
      }
      updateChartIndicators();
      document.getElementById("statusMsg").innerText = "Loaded";
    }
  });

  // Save button
  document.getElementById("saveStrategyBtn").addEventListener("click", async () => {
    const currentId = sel.value;
    if (!currentId || [ACTION_NEW, ACTION_DELETE].includes(currentId)) {
      alert("Select or create a strategy to save");
      return;
    }
    const strat = {
      id: currentId,
      name: sel.options[sel.selectedIndex].innerText,
      buy_rules: collectRules(buyBody),
      sell_rules: collectRules(sellBody),
    };
    await saveStrategy(strat);
    document.getElementById("statusMsg").innerText = "Saved ✔";
  });

  // Initial load
  await loadStrategies();
  refreshDropdown();

  // Populate ticker dropdown and chart
  const tickerDd = document.getElementById("tickerDropdown");
  const tickers = await fetchTickersList();
  tickers.forEach((t) => {
    const opt = el("option", {
      value: t.symbol || "",
      innerText: `${t.symbol} - ${t.name}`,
    });
    opt.dataset.symbol = (t.symbol || "").trim();
    tickerDd.appendChild(opt);
  });
  if (tickers.length) {
    tickerDd.value = tickers[0].symbol;
    createBuilderChart(tickers[0].symbol);
  }

  tickerDd.addEventListener("change", () => {
    const sym = tickerDd.selectedOptions[0].dataset.symbol || tickerDd.value || "";
    createBuilderChart(sym);
  });

  // trim values added earlier
  tickerDd.querySelectorAll("option").forEach(o=>o.value=o.value.trim());
}); 