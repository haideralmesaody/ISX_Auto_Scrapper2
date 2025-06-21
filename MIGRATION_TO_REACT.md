# Migration Plan: Static HTML → React (Vite + TypeScript)

_Last updated: 2025-06-11_

---

## 1 Purpose
This document describes how to migrate the front-end of **ISX Auto Scrapper** from scattered static HTML/JS pages to a modern React application powered by **Vite** while keeping the existing Go backend for all API logic and static file hosting.

Goals :
* Minimal disruption – production stays usable during the transition.
* Re-use as much existing JavaScript (Highcharts logic) as practical.
* Deliver working slices (dashboard, builder, daily report) incrementally.
* Zero changes to Go business logic; only extra routes for static assets.

---

## 2 Prerequisites
| Tool | Version | Notes |
|------|---------|-------|
| Node.js | ≥ 18 | for Vite build/dev server |
| npm / pnpm | latest | package manager |
| Go | ≥ 1.20 | unchanged |

Install globally if missing:
```bash
npm i -g npm  # optional upgrade
```

---

## 3 High-level Steps
1. Bootstrap `webapp/` Vite project (React + TS).
2. Add React Router & top-level navigation shell.
3. Temporarily embed legacy HTML inside iframes (keeps site live).
4. Port pages one by one, starting with Strategy Builder.
5. Replace dashboard ticker table & Highcharts code.
6. Port daily report page.
7. Introduce TailwindCSS (optional but recommended).
8. Build pipeline & Dockerfile updates.
9. Remove legacy static pages once parity is confirmed.

Estimated effort **≈ 11 working days**.

---

## 4 Detailed Task List
### Phase 0 – Project Bootstrap (½ day)
1. Create a working branch `react-migration`.
2. Initialise Vite project:
   ```bash
   npx create-vite@latest webapp -- --template react-ts
   cd webapp && npm i
   ```
3. Configure proxy in `vite.config.ts`:
   ```ts
   export default defineConfig({
     server: {
       proxy: {
         '/api': 'http://localhost:8080'  // Go server
       }
     }
   });
   ```
4. Update Go `web_server.go` to serve built assets:
   ```go
   fs := http.FileServer(http.Dir("webapp/dist"))
   r.PathPrefix("/").Handler(fs) // after API routes
   ```

---
### Phase 1 – Shell & Routing (1 day)
* Install router: `npm i react-router-dom`.
* Create `src/pages/{Dashboard,DailyReport,Builder}.tsx` (empty stubs).
* Build nav bar in `App.tsx` using `<NavLink>`.
* Verify routes switch with hot-reload.

---
### Phase 2 – Legacy Embedding (0.5 day)
* Each stub page temporarily returns:
  ```tsx
  export default () => <iframe src="/index.html" style={{width:'100%',height:'100%',border:'none'}}/>;
  ```
* Site functions exactly as before through iframes.

---
### Phase 3 – Port Strategy Builder (2 days)
* Convert `strategy_builder.html/js` into React components:
  * `RulesTable`, `RuleRow`, `IndicatorSelect`.
* Local `useState` arrays replace direct DOM manipulation.
* Fetch indicators from new endpoint `/api/indicators` (placeholder JSON now).
* Implement CRUD via `fetch('/api/strategies')`.
* Remove iframe for `/builder` route.

---
### Phase 4 – Port Dashboard (3 days)
* Install Highcharts wrapper:
  ```bash
  npm i highcharts highcharts-react-official
  ```
* Components:
  * `TickerTable.tsx` – sorting, filtering, sparkline (reuse SVG code).
  * `MainChart.tsx` – wrap existing `createChart` logic inside `useEffect`.
* Fetch `/api/tickers` and `/api/ticker/{symbol}` just like today.
* Delete dashboard iframe.

---
### Phase 5 – Port Daily Report (2 days)
* Use MUI DataGrid or plain table.
* Fetch `Strategy_Summary.json`.
* Provide CSV download.

---
### Phase 6 – Styling (1 day)
* Install TailwindCSS:
  ```bash
  npm i -D tailwindcss postcss autoprefixer
  npx tailwindcss init
  ```
* Gradually swap old classes for Tailwind utilities.

---
### Phase 7 – Build & Deployment (½ day)
* Add `npm run build` in GitHub Actions before Go build.
* Embed or copy `webapp/dist` into final container.
* Example multi-stage Dockerfile included in repo (`docs/docker/webapp.Dockerfile`).

---
### Phase 8 – Clean-up & Documentation (½ day)
* Remove `web/*.html` & legacy `.js` once migrated.
* Update README with new dev instructions:
  ```bash
  # Front-end
  cd webapp && npm run dev  # proxy to :8080

  # Back-end
  go run cmd/server/main.go --mode web
  ```

---

## 5 Rollback Strategy
If unforeseen issues arise, switch back to the old static pages by:
1. Serving `web/` folder again (`git checkout main -- web`).
2. Disabling Vite build step.
No database migrations are involved, so risk is minimal.

---

## 6 Owner & Reviewers
* **Owner:** _front-end lead TBD_
* **Backend liaison:** `@go-maintainer`
* **Reviewers:** QA, UX, Ops.

---

## 7 Glossary
* **Vite** – lightning-fast dev server & bundler.
* **SPA** – Single-Page Application.
* **Highcharts** – charting library already used for candlestick graphs.

---

_Approved by: ___________________  Date: ___________ 