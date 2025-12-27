# HodlBook - Development Status

Self-hosted crypto portfolio management application.
First phase focuses on backend API development.

**Stack:** Go/Gin backend, htmx frontend (pending), SQLite database

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

## Pending

### Testing
- [ ] Integration tests for API endpoints
- [x] Service-level tests

### Infrastructure
- [ ] `GET /api/health` endpoint
- [ ] Hot-reload for development (Air)

### Frontend (htmx)
- [ ] Layout component
- [ ] Asset management pages
- [ ] Transaction forms
- [ ] Exchange forms
- [ ] Dashboard with charts
- [ ] Settings page
- [ ] Live price updates via pubsub/websocket

### Documentation
- [ ] Local setup instructions in README

## Out of Scope
- Automatic exchange API integrations
- Tax reporting or compliance features
- Hosted service / SaaS
