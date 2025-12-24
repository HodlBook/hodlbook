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
- [x] GORM auto-migration
- [x] Repository pattern with tests

### REST API
- [x] Converted to pure REST/JSON API 
- [x] All routes under `/api` prefix
- [x] HTTP handling in `/internal/handler`

### Assets API
- [x] `POST /api/assets` - Create asset
- [x] `GET /api/assets` - List all assets
- [x] `GET /api/assets/:id` - Get asset details
- [x] `PUT /api/assets/:id` - Update asset
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

## In Progress

### Asset historic values
Add asset historic value tracking for accurate portfolio valuation over time.
- all assets in portfolio will have a new price entry added daily at GMT-0 midnight
- this addition will be done via background job
- [ ] `GET /api/prices/:asset/history` - Get historic prices for asset

### Portfolio Analytics
- [ ] `GET /api/portfolio/summary` - Total portfolio value
- [ ] `GET /api/portfolio/allocation` - Wallet share by asset
- [ ] `GET /api/portfolio/performance` - Profit/loss calculations
- [ ] `GET /api/portfolio/history` - Portfolio value over time

### Price Integration
- [ ] `GET /api/prices/:asset` - Get current price
- [ ] Background job for periodic updates
- [ ] Price caching mechanism

## Pending

### Testing
- [ ] Handler-level tests
- [ ] Integration tests for API endpoints

### Infrastructure
- [ ] `GET /api/health` endpoint
- [ ] Graceful shutdown handling
- [ ] Hot-reload for development (Air)
- [ ] Structured logging
- [ ] Standardized error messages

### Frontend (htmx)
- [ ] Layout component
- [ ] Asset management pages
- [ ] Transaction forms
- [ ] Exchange forms
- [ ] Dashboard with charts
- [ ] Settings page

### Documentation
- [ ] Local setup instructions in README
- [ ] API documentation (OpenAPI/Swagger)

## Out of Scope
- Automatic exchange API integrations
- Tax reporting or compliance features
- Hosted service / SaaS
