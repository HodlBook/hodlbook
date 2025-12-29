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

| View         | Status  | Description                                                                          |
|--------------|---------|--------------------------------------------------------------------------------------|
| Dashboard    | Done    | Portfolio summary, allocation chart, holdings, recent transactions                   |
| Portfolio    | Done    | Full holdings table, performance metrics, historical chart, filters                  |
| Assets       | Done    | Asset list with prices/holdings/value, add modal, delete modal                       |
| Transactions | Done    | Transaction table, filters (asset, type, date), pagination, add/edit/delete modals   |
| Exchanges    | Done    | Exchange table, filters (from/to asset, date), pagination, add/edit/delete modals    |
| Prices       | Done    | Live price cards with SSE updates, holdings/value display, connection status         |

### UI Components

| Component      | Status | Purpose                                      |
|----------------|--------|----------------------------------------------|
| Sidebar        | Done   | Navigation between views                     |
| Summary Card   | Done   | Display key metrics                          |
| Data Table     | Done   | Sortable, filterable tables with pagination  |
| Modal/Dialog   | Done   | Forms for CRUD operations                    |
| Charts         | Done   | Pie (allocation), Line (history)             |
| Form Inputs    | Done   | Dropdowns, date pickers, number inputs       |
| Toast/Alerts   | Done   | Success/error feedback                       |
| Loading States | Done   | Skeletons, spinners                          |
| Empty States   | Done   | When no data exists                          |

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

## Running the Application

### Without Docker

1. Install Go 1.24 and SQLite development libraries.
2. Clone the repository:
   ```bash
   git clone https://github.com/HodlBook/hodlbook.git
   cd hodlbook
   ```
3. Run the application:
   ```bash
   make run
   ```
4. Access the application at:
   - Web UI: `http://localhost:2008`
   - Swagger API Docs: `http://localhost:2008/swagger/index.html`

### With Docker

1. Ensure Docker and Docker Compose are installed.
2. Clone the repository:
   ```bash
   git clone https://github.com/HodlBook/hodlbook.git
   cd hodlbook
   ```
3. Build and run the Docker container:
   ```bash
   make run-docker
   ```
4. Access the application at:
   - Web UI: `http://localhost:2008`
   - Swagger API Docs: `http://localhost:2008/swagger`

Data is persisted automatically using a Docker named volume (`hodlbook_data`).

### With Docker (Ephemeral)

For testing or development without data persistence:
```bash
make run-docker-ephemeral
```

Data will be lost when the container is removed.
