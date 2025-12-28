# HodlBook

A self-hosted cryptocurrency portfolio tracker.

## Features

- Register asset deposits/withdrawals
- Register exchanges from asset A to asset B
- Track current value of assets and value increase/decrease
- Show wallet share by asset (allocation)
- Show wallet value and profit by currency of reference

## What HodlBook is NOT

- No automatic exchange API integrations
- No tax reporting or compliance features
- Not a hosted service or SaaS product

## Tech Stack

### Backend
- **Language**: Go 1.24
- **Framework**: Gin
- **Database**: SQLite (GORM)
- **Price Data**: Binance API (primary), CoinGecko (fallback)

### Frontend
- **Approach**: Server-side rendering with HTMX + Alpine.js
- **Styling**: Custom CSS (dark theme)
- **Charts**: Chart.js

### Deployment
- Docker Compose
- Umbrel App Store compatible

## Project Structure

```
hodlbook/
├── cmd/main.go                 # Application entry point
├── internal/
│   ├── controller/             # API request handlers (JSON)
│   ├── handler/                # API route setup
│   ├── models/                 # Data models
│   ├── repo/                   # Database repository
│   ├── service/                # Background services
│   └── ui/                     # Web UI (HTML)
│       ├── handler/            # UI route handlers
│       ├── static/             # CSS, JS assets
│       └── templates/          # HTML templates
├── pkg/
│   ├── database/               # SQLite initialization
│   ├── integrations/           # External services
│   └── types/                  # Interfaces
└── docs/                       # Swagger documentation
```

## UI Architecture

### Core Views

| View | Status | Description |
|------|--------|-------------|
| Dashboard | Done | Portfolio summary, allocation chart, holdings, recent transactions |
| Portfolio | Pending | Holdings table, performance metrics, historical chart |
| Assets | Pending | Asset management, add/delete assets |
| Transactions | Pending | Transaction list with filters, CRUD operations |
| Exchanges | Pending | Exchange list with filters, CRUD operations |
| Prices | Pending | Live price ticker with SSE updates |

### UI Components

| Component | Status | Purpose |
|-----------|--------|---------|
| Sidebar | Done | Navigation between views |
| Summary Card | Done | Display key metrics |
| Data Table | Done | Sortable, filterable tables with pagination |
| Modal/Dialog | Done | Forms for CRUD operations |
| Charts | Done | Pie (allocation), Line (history) |
| Form Inputs | Done | Dropdowns, date pickers, number inputs |
| Toast/Alerts | Done | Success/error feedback |
| Loading States | Done | Skeletons, spinners |
| Empty States | Done | When no data exists |

### View Details

#### 1. Dashboard (Home)
- Portfolio summary card: Total value, P/L percentage
- Asset allocation pie/donut chart
- Quick stats: Asset count, best performer
- Recent transactions list (last 5)
- Holdings preview (top 5)

#### 2. Portfolio View
- Holdings table: Asset, Amount, Price, Value, Change, Allocation %
- Performance metrics: Total invested, current value, unrealized P/L
- Historical chart: Portfolio value over time (7d/30d/90d/1y)
- Filter/sort controls

#### 3. Assets Management
- Assets list/table: Symbol, Name, Price, Holdings, Actions
- Add asset modal: Symbol input, Name
- Asset detail view: Price history chart, transactions, delete

#### 4. Transactions View
- Transactions table: Date, Type, Asset, Amount, Notes, Actions
- Filters: By asset, type (buy/sell/deposit/withdraw), date range
- Pagination controls
- Add/Edit transaction modals

#### 5. Exchanges View
- Exchanges table: Date, From, To, Fee, Notes, Actions
- Filters: By asset, date range
- Add/Edit exchange modals

#### 6. Live Prices View
- Real-time price ticker via SSE
- Price table with live updates
- Sparkline charts (24h movement)

## API Endpoints

### Assets
- `GET /api/assets` - List all assets
- `POST /api/assets` - Create asset
- `GET /api/assets/:id` - Get asset
- `DELETE /api/assets/:id` - Delete asset

### Transactions
- `GET /api/transactions` - List transactions (filterable)
- `POST /api/transactions` - Create transaction
- `GET /api/transactions/:id` - Get transaction
- `PUT /api/transactions/:id` - Update transaction
- `DELETE /api/transactions/:id` - Delete transaction

### Exchanges
- `GET /api/exchanges` - List exchanges (filterable)
- `POST /api/exchanges` - Create exchange
- `GET /api/exchanges/:id` - Get exchange
- `PUT /api/exchanges/:id` - Update exchange
- `DELETE /api/exchanges/:id` - Delete exchange

### Portfolio
- `GET /api/portfolio/summary` - Portfolio summary
- `GET /api/portfolio/allocation` - Asset allocation
- `GET /api/portfolio/performance` - P/L per asset
- `GET /api/portfolio/history` - Historical values

### Prices
- `GET /api/prices` - Current prices
- `GET /api/prices/:symbol` - Price for symbol
- `GET /api/prices/history/:id` - Price history
- `GET /api/prices/stream` - SSE live prices

## Running

```bash
# Development
go run cmd/main.go

# Access
http://localhost:8080        # Web UI
http://localhost:8080/swagger # API docs
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| APP_PORT | 8080 | Server port |
| DB_PATH | ./data/hodlbook.db | SQLite database path |
