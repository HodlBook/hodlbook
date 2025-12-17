package main

import (
	"log"
	"os"

	"hodlbook/internal/controllers"
	"hodlbook/internal/repo"
	"hodlbook/pkg/database"
	"hodlbook/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	utils.LoadEnv()

	// Initialize database
	dbPath := utils.GetEnv("DB_PATH", "./data/hodlbook.db")
	db, err := database.New(database.WithPath(dbPath))
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Run migrations
	if err := repo.Migrate(db.Get()); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Initialize repositories with options pattern
	txRepo, err := repo.NewTransaction(repo.WithDB(db.Get()))
	if err != nil {
		log.Fatal("Failed to create transaction repository:", err)
	}

	priceRepo, err := repo.NewPrice(repo.WithPriceDB(db.Get()))
	if err != nil {
		log.Fatal("Failed to create price repository:", err)
	}

	// Initialize main repository with options pattern
	repository, err := repo.New(
		repo.WithTransactionRepository(txRepo),
		repo.WithPriceRepository(priceRepo),
	)
	if err != nil {
		log.Fatal("Failed to create repository:", err)
	}

	// Initialize Gin router
	r := gin.Default()

	// Load HTML templates
	r.LoadHTMLGlob("templates/*")

	// Serve static files
	r.Static("/static", "./static")

	// Setup routes with repository
	setupRoutes(r, repository)

	// Get port from environment or use default
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	log.Printf("Starting HodlBook on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRoutes(r *gin.Engine, repository *repo.Repository) {
	// Initialize controllers with options pattern
	assetCtrl, err := controllers.NewAsset()
	if err != nil {
		log.Fatal("Failed to create asset controller:", err)
	}

	transactionCtrl, err := controllers.NewTransaction(controllers.WithTransactionRepo(repository.Transaction))
	if err != nil {
		log.Fatal("Failed to create transaction controller:", err)
	}

	portfolioCtrl, err := controllers.NewPortfolio(controllers.WithRepository(repository))
	if err != nil {
		log.Fatal("Failed to create portfolio controller:", err)
	}

	// Home/Dashboard routes
	r.GET("/", portfolioCtrl.Dashboard)

	// Asset routes
	assets := r.Group("/assets")
	{
		assets.GET("", assetCtrl.List)
		assets.GET("/new", assetCtrl.New)
		assets.POST("", assetCtrl.Create)
		assets.GET("/:id", assetCtrl.Show)
		assets.GET("/:id/edit", assetCtrl.Edit)
		assets.PUT("/:id", assetCtrl.Update)
		assets.DELETE("/:id", assetCtrl.Delete)
	}

	// Transaction routes
	transactions := r.Group("/transactions")
	{
		transactions.GET("", transactionCtrl.List)
		transactions.GET("/new", transactionCtrl.New)
		transactions.POST("", transactionCtrl.Create)
		transactions.GET("/:id", transactionCtrl.Show)
		transactions.GET("/:id/edit", transactionCtrl.Edit)
		transactions.PUT("/:id", transactionCtrl.Update)
		transactions.DELETE("/:id", transactionCtrl.Delete)
	}

	// API routes for HTMX
	api := r.Group("/api")
	{
		api.GET("/portfolio/summary", portfolioCtrl.Summary)
		api.GET("/portfolio/allocation", portfolioCtrl.Allocation)
	}
}
