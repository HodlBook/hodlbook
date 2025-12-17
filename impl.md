# HodlBook - Implementation Steps

## Phase 1: Project Setup & Technical Stack

### 1.1 Technology Stack Decision
- [x] Choose back-end framework (recommendations: FastAPI/Python, Node.js/Express, or Go/Gin)
    - Go gin
- [x] Choose front-end framework (recommendations: React, Vue.js, or Svelte)
    - Go htmx
- [x] Choose database (SQLite for development, PostgreSQL for production as alternative)
    - Initially sqlite;
    - in the future allow user to chose the db type and set the connection info;
- [x] Choose deployment pipeline (GitHub Actions, GitLab CI, or custom scripts)
    - github actions for open source projects;

### 1.2 Project Initialization
- [x] Initialize Git repository structure
- [x] Set up back-end project scaffolding (Go + Gin + MVC structure)
- [x] Set up front-end scaffolding (htmx templates integrated with back-end)
- [x] Create Docker Compose configuration
- [x] Create .gitignore and environment variable templates

### 1.3 Development Environment
- [x] Create Dockerfile for back-end service
- [x] Set up docker-compose.yml for local development
- [ ] Configure hot-reload for development
- [ ] Document local setup instructions in README

**Note**: Front-end implementation (templates/htmx) will be done after back-end API is complete.

## Phase 2: Database Design & Implementation

### 2.1 Database Schema
- [x] Design `transactions` table (id, type, asset_id, amount, timestamp, notes)
- [x] Design `prices` table (id, asset_id, currency, price, timestamp)
- [ ] Design `assets` table (id, symbol, name, type) - managed externally
- [ ] Design `exchanges` table (id, from_asset_id, to_asset_id, from_amount, to_amount, timestamp) - derived from transactions
- [ ] Design `settings` table (id, key, value) for reference currency preference - future enhancement
- [x] Create database migration files (using GORM auto-migration)
- [x] Implement database initialization script

### 2.2 Database Access Layer
- [x] Implement database connection handler (`pkg/utils/db.go`)
- [x] Create GORM models (`internal/repo/models.go`)
- [x] Implement transaction CRUD operations (`internal/repo/transaction_repo.go`)
- [x] Implement price data storage and retrieval (`internal/repo/price_repo.go`)
- [x] Repository pattern with unified access (`internal/repo/repo.go`)

**Note**: Assets and Exchanges are calculated from transaction data, not stored directly.

## Phase 3: Back-end API Development

### 3.1 Core API Endpoints - Assets
- [ ] POST /api/assets - Create new asset
- [ ] GET /api/assets - List all assets
- [ ] GET /api/assets/:id - Get asset details
- [ ] PUT /api/assets/:id - Update asset
- [ ] DELETE /api/assets/:id - Delete asset

### 3.2 Core API Endpoints - Transactions
- [ ] POST /api/transactions/deposit - Register deposit
- [ ] POST /api/transactions/withdrawal - Register withdrawal
- [ ] GET /api/transactions - List all transactions
- [ ] GET /api/transactions/:id - Get transaction details
- [ ] PUT /api/transactions/:id - Update transaction
- [ ] DELETE /api/transactions/:id - Delete transaction

### 3.3 Core API Endpoints - Exchanges
- [ ] POST /api/exchanges - Register exchange from asset A to B
- [ ] GET /api/exchanges - List all exchanges
- [ ] GET /api/exchanges/:id - Get exchange details
- [ ] PUT /api/exchanges/:id - Update exchange
- [ ] DELETE /api/exchanges/:id - Delete exchange

### 3.4 Price Integration
- [ ] Implement price fetcher service (CoinGecko or CoinMarketCap free API)
- [ ] GET /api/prices/:asset - Get current price for asset
- [ ] Create background job to update prices periodically
- [ ] Implement price caching mechanism

### 3.5 Analytics & Reporting
- [ ] GET /api/portfolio/summary - Calculate total portfolio value
- [ ] GET /api/portfolio/allocation - Calculate wallet share by asset
- [ ] GET /api/portfolio/performance - Calculate profit/loss
- [ ] GET /api/portfolio/history - Portfolio value over time
- [ ] Implement currency conversion logic

## Phase 4: Front-end Development

### 4.1 Core UI Components
- [ ] Create layout component (header, sidebar, main content)
- [ ] Create navigation component
- [ ] Create asset list component
- [ ] Create transaction form component
- [ ] Create exchange form component
- [ ] Create data table component

### 4.2 Asset Management Pages
- [ ] Assets overview page (list all assets)
- [ ] Add/Edit asset modal or page
- [ ] Asset detail page (show transactions and exchanges)

### 4.3 Transaction Management Pages
- [ ] Deposit form page/modal
- [ ] Withdrawal form page/modal
- [ ] Transaction history page with filtering
- [ ] Transaction detail view

### 4.4 Exchange Management Pages
- [ ] Exchange form page/modal (asset A → asset B)
- [ ] Exchange history page
- [ ] Exchange detail view

### 4.5 Dashboard & Analytics
- [ ] Dashboard overview page
- [ ] Portfolio value card/widget
- [ ] Asset allocation chart (pie/donut chart)
- [ ] Profit/loss indicator
- [ ] Portfolio history chart (line chart)
- [ ] Currency selector for reference currency

### 4.6 Settings & Configuration
- [ ] Settings page
- [ ] Reference currency selector
- [ ] Theme selector (optional)
- [ ] Data export functionality (optional)

## Phase 5: Integration & Testing

### 5.1 API Integration
- [ ] Set up API client/service layer in front-end
- [ ] Implement error handling and loading states
- [ ] Connect all forms to API endpoints
- [ ] Connect all data displays to API endpoints
- [ ] Implement optimistic updates where appropriate

### 5.2 Testing
- [ ] Write unit tests for back-end services
- [ ] Write unit tests for front-end components
- [ ] Write integration tests for API endpoints
- [ ] Write E2E tests for critical user flows
- [ ] Test Docker Compose deployment locally

## Phase 6: Deployment Preparation

### 6.1 Production Configuration
- [ ] Create production Dockerfile for back-end
- [ ] Create production Dockerfile for front-end
- [ ] Create production docker-compose.yml
- [ ] Set up environment variable management
- [ ] Configure logging and monitoring
- [ ] Set up database backup strategy

### 6.2 Deployment Pipeline
- [ ] Create CI/CD pipeline configuration
- [ ] Set up automated testing in pipeline
- [ ] Set up Docker image building
- [ ] Configure deployment triggers
- [ ] Document deployment process

### 6.3 Documentation
- [ ] Write comprehensive README
- [ ] Document API endpoints (OpenAPI/Swagger)
- [ ] Create user guide
- [ ] Document deployment steps
- [ ] Document database schema
- [ ] Add troubleshooting guide

## Phase 7: Optional Enhancements

### 7.1 Nice-to-Have Features
- [ ] Dark mode support
- [ ] Data import/export (CSV, JSON)
- [ ] Multiple portfolio support
- [ ] Notes/tags for transactions
- [ ] Search and advanced filtering
- [ ] Notification system
- [ ] Mobile-responsive design improvements

### 7.2 Performance Optimizations
- [ ] Implement pagination for large datasets
- [ ] Add database indexes
- [ ] Implement caching strategies
- [ ] Optimize bundle size
- [ ] Lazy loading for front-end routes

## Notes

### Technology Recommendations

**Back-end Stack Options:**
1. **Python + FastAPI** - Fast, modern, auto-generated docs, great type hints
2. **Node.js + Express** - JavaScript full-stack, large ecosystem
3. **Go + Gin** - High performance, compiled binary, minimal dependencies

**Front-end Stack Options:**
1. **React + Vite** - Popular, large ecosystem, good tooling
2. **Vue.js + Vite** - Easier learning curve, good documentation
3. **Svelte + SvelteKit** - Minimal bundle size, intuitive syntax

**Database:**
- SQLite for simplicity and self-contained deployment
- PostgreSQL if more advanced features needed

**Deployment:**
- Docker Compose for easy self-hosting
- Single command deployment: `docker-compose up -d`

### Development Order Priority

1. **High Priority**: Database schema, core API (transactions, exchanges), basic UI
2. **Medium Priority**: Portfolio analytics, price integration, dashboard
3. **Low Priority**: Advanced filtering, export features, theming
