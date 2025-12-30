# HodlBook - Development Status

Self-hosted crypto portfolio management application.

**Stack:** Go/Gin backend, HTMX + Alpine.js frontend, SQLite database

## Current Status

### Completed
- **Backend**:
  - REST API for assets, transactions, exchanges, portfolio, and prices.
  - SQLite database with GORM auto-migration.
  - Price fetcher service with Binance and CoinGecko integrations.
  - Background jobs for price updates and historic value storage.
  - Structured logging, graceful shutdown, and in-memory caching.

- **Frontend**:
  - Server-side rendering with Go templates.
  - Dynamic updates using HTMX and Alpine.js.
  - Charts for portfolio allocation and historical performance.
  - Fully functional views: Dashboard, Portfolio, Assets, Transactions, Exchanges, Prices.

- **Infrastructure**:
  - Docker and Docker Compose setup.
  - GitHub Actions for CI/CD (build, release, docker push to GHCR).
  - Swagger documentation for API.
  - Local setup instructions completed.
  - Umbrel app store files prepared (see below).

### Pending
- **Development**:
  - Hot-reload for development (e.g., Air).
  - Additional input validation and error handling.

### Desirable Future Enhancements
- Add deposit/exchange tags/wallet names for better organization
- Advanced filtering and sorting options in tables.
- Export/import functionality for data backup.
    - import thru CSV/JSON
    - import from transaction link for popular blockchains
    - export to CSV/JSON/PDF
- web-browser wallets integration
- User authentication and multi-user support.
- Additional price data sources.
- Mobile-friendly UI improvements.

### Out of Scope
- Automatic centralized exchange API integrations.
- Tax reporting or compliance features.
- Hosted service / SaaS.

