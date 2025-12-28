# HodlBook - Development Status

Self-hosted crypto portfolio management application.

**Stack:** Go/Gin backend, HTMX + Alpine.js frontend, SQLite database

## Completed

### Project Setup
- [x] Go + Gin + MVC structure
- [x] Docker/Docker Compose configuration
- [ ] Git repository and CI (GitHub Actions)

### Database Layer
- [x] `transactions` table (id, type, asset_id, amount, timestamp, notes)
- [x] `prices` table (id, asset_id, currency, price, timestamp)
- [x] `assets` table (id, symbol, name, type, decimals)
- [x] `exchanges` table (id, from_asset_id, to_asset_id, from_amount, to_amount, fee, timestamp, notes)
- [x] `asset_historic_values` table (asset_id, value, timestamp)
- [x] GORM auto-migration
- [x] Repository pattern with tests

### REST API
- [x] Converted to pure REST/JSON API
- [x] All routes under `/api` prefix
- [x] HTTP handling in `/internal/handler`
- [x] Swagger documentation at `/swagger/index.html`

### Assets API
- [x] `POST /api/assets` - Create asset
- [x] `GET /api/assets` - List all assets
- [x] `GET /api/assets/:id` - Get asset details
- [x] `DELETE /api/assets/:id` - Delete asset

### Transactions API
- [x] `POST /api/transactions` - Create transaction
- [x] `GET /api/transactions` - List with pagination/filtering
- [x] `GET /api/transactions/:id` - Get transaction details
- [x] `PUT /api/transactions/:id` - Update transaction
- [x] `DELETE /api/transactions/:id` - Delete transaction
- [x] Input validation with request DTOs

### Exchanges API
- [x] `POST /api/exchanges` - Register exchange
- [x] `GET /api/exchanges` - List with pagination/filtering
- [x] `GET /api/exchanges/:id` - Get exchange details
- [x] `PUT /api/exchanges/:id` - Update exchange
- [x] `DELETE /api/exchanges/:id` - Delete exchange

### Price Integration
- [x] Price fetcher service (Binance API with CoinGecko fallback)
- [x] Fetch single asset price
- [x] Fetch multiple asset prices
- [x] Background job for periodic updates (1-minute interval)
- [x] Price caching mechanism (in-memory cache)
- [x] PubSub for live price broadcasting

### Asset Historic Values
- [x] Daily price storage via background scheduler
- [x] Historic value repository

### Price API
- [x] `GET /api/prices` - List all current prices from cache
- [x] `GET /api/prices/:symbol` - Get current price for specific asset
- [x] `GET /api/prices/history/:id` - Get historic prices for asset
- [x] `GET /api/prices/stream` - SSE endpoint for live price updates

### Infrastructure
- [x] Graceful shutdown handling (signal handling)
- [x] Structured logging (slog)
- [x] Scheduler package (ticker-based)
- [x] In-memory pubsub package
- [x] In-memory cache package
- [x] Repository interface abstraction (`pkg/types/repo`)
- [x] Cache interface abstraction (`pkg/types/cache`)

### Portfolio Analytics
- [x] `GET /api/portfolio/summary` - Total portfolio value with holdings breakdown
- [x] `GET /api/portfolio/allocation` - Wallet share by asset with percentages
- [x] `GET /api/portfolio/performance` - Profit/loss calculations per asset
- [x] `GET /api/portfolio/history` - Portfolio value over time (configurable days)

## Frontend (HTMX + Alpine.js)

### Architecture
- [x] Server-side rendering with Go templates
- [x] HTMX for dynamic HTML fragment updates
- [x] Alpine.js for client-side interactivity
- [x] Chart.js for data visualization
- [x] Custom CSS dark theme

### UI Components
- [x] Sidebar navigation (collapsible)
- [x] Navbar with live connection indicator
- [x] Summary cards
- [x] Data tables (sortable, filterable, paginated)
- [x] Modal/Dialog system
- [x] Form inputs (text, select, date, number)
- [x] Toast notifications
- [x] Loading states (skeletons, spinners)
- [x] Empty states
- [x] Charts (pie/donut, line)

### Views

| View | Status | Features |
|------|--------|----------|
| Dashboard | Done | Summary cards, allocation chart, holdings table, recent transactions |
| Portfolio | Done | Full holdings table, performance metrics, historical chart, filters |
| Assets | Done | Asset list with prices/holdings/value, add modal, delete modal |
| Transactions | Done | Transaction table, filters (asset, type, date), pagination, add/edit/delete modals |
| Exchanges | Done | Exchange table, filters (from/to asset, date), pagination, add/edit/delete modals |
| Prices | Done | Live price cards with SSE updates, holdings/value display, connection status |

### Dashboard Features (Completed)
- [x] Portfolio summary (total value, asset count, P/L, best performer)
- [x] Portfolio value line chart (7d/30d/90d/1y toggle)
- [x] Asset allocation donut chart with legend
- [x] Top 5 holdings table
- [x] Last 5 transactions list
- [x] Auto-refresh via HTMX (60s interval)
- [x] Live connection status indicator

### Portfolio Features (Completed)
- [x] Summary cards (total invested, current value, unrealized P/L)
- [x] Portfolio history chart (7d/30d/90d/1y toggle)
- [x] Holdings table with sorting (by value, amount, change, allocation, symbol)
- [x] Performance table (cost basis, P/L per asset)
- [x] Auto-refresh via HTMX (60s interval)

### Assets Features (Completed)
- [x] Asset list table with current price, holdings, value
- [x] Add asset modal (symbol, name)
- [x] Delete asset confirmation modal
- [x] Auto-uppercase symbol input
- [x] Sorted by portfolio value

### Transactions Features (Completed)
- [x] Transaction table with date, type, asset, amount, notes
- [x] Type badges (buy/sell/deposit/withdraw)
- [x] Filters (asset dropdown, type dropdown, date range)
- [x] Pagination (20 per page)
- [x] Add transaction modal
- [x] Edit transaction modal
- [x] Delete transaction confirmation modal

### Exchanges Features (Completed)
- [x] Exchange table with date, from/to amounts, rate, fee, notes
- [x] From/To asset badges
- [x] Filters (from asset, to asset, date range)
- [x] Pagination (20 per page)
- [x] Add exchange modal (from/to asset/amount, fee, timestamp, notes)
- [x] Edit exchange modal
- [x] Delete exchange confirmation modal

### Prices Features (Completed)
- [x] Price cards grid layout
- [x] Real-time SSE price updates with visual feedback (green/red flash)
- [x] Connection status indicator
- [x] Holdings and value display per asset
- [x] Auto-reconnect on disconnect

### Startup Improvements
- [x] Immediate price fetch on service startup (no wait for first tick)

## Pending

### Infrastructure
- [x] `GET /api/health` endpoint (used by navbar live indicator)
- [x] Favicon (icon.png)
- [ ] Hot-reload for development (Air)

### Documentation
- [x] README with tech stack and structure
- [ ] Local setup instructions

## Out of Scope
- Automatic exchange API integrations
- Tax reporting or compliance features
- Hosted service / SaaS
