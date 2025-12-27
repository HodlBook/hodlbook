package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	_ "hodlbook/docs"
	"hodlbook/internal/handler"
	"hodlbook/internal/repo"
	"hodlbook/internal/service"
	"hodlbook/pkg/database"
	"hodlbook/pkg/integrations/memcache"
	"hodlbook/pkg/integrations/prices"
	"hodlbook/pkg/integrations/wmPubsub"
	"hodlbook/pkg/utils"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title HodlBook API
// @version 1.0
// @description Cryptocurrency portfolio tracking API

// @host localhost:8080
// @BasePath /

func main() {
	utils.LoadEnv()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbPath := utils.GetEnv("DB_PATH", "./data/hodlbook.db")
	db, err := database.New(database.WithPath(dbPath))
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	repository, err := repo.New(db.Get())
	if err != nil {
		log.Fatal("Failed to create repository:", err)
	}

	if err := repository.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	priceFetcher := prices.NewPriceService()
	priceCache := memcache.New[string, float64]()
	priceCh := make(chan []byte, 10)
	sseCh := make(chan []byte, 10)
	pricePublisher := wmPubsub.New(
		wmPubsub.WithChannel(priceCh),
		wmPubsub.WithContext(ctx),
		wmPubsub.WithTopic("prices"),
		wmPubsub.WithLogger(logger),
		wmPubsub.WithHandler(func(data []byte) error {
			select {
			case sseCh <- data:
			default:
				logger.Warn("sseCh full, dropping message")
			}
			return nil
		}),
	)
	if err := pricePublisher.Subscribe(); err != nil {
		log.Fatal("Failed to start price subscriber:", err)
	}

	livePriceSvc, err := service.NewLivePriceService(
		service.WithLivePriceContext(ctx),
		service.WithLivePriceLogger(logger),
		service.WithLivePriceCache(priceCache),
		service.WithLivePriceFetcher(priceFetcher),
		service.WithLivePricePublisher(pricePublisher),
		service.WithLivePriceRepo(repository),
	)
	if err != nil {
		log.Fatal("Failed to create live price service:", err)
	}

	historicPriceSvc, err := service.NewHistoricPriceService(
		service.WithHistoricPriceContext(ctx),
		service.WithHistoricPriceLogger(logger),
		service.WithHistoricPriceFetcher(priceFetcher),
		service.WithHistoricPriceRepo(repository),
	)
	if err != nil {
		log.Fatal("Failed to create historic price service:", err)
	}

	assetCreatedCh := make(chan []byte, 10)
	assetHistoricSvc, err := service.NewAssetHistoricService(
		service.WithAssetHistoricContext(ctx),
		service.WithAssetHistoricLogger(logger),
		service.WithAssetHistoricFetcher(priceFetcher),
		service.WithAssetHistoricRepo(repository),
		service.WithAssetHistoricChannel(assetCreatedCh),
	)
	if err != nil {
		log.Fatal("Failed to create asset historic service:", err)
	}

	if err := livePriceSvc.Start(); err != nil {
		log.Fatal("Failed to start live price service:", err)
	}
	if err := historicPriceSvc.Start(); err != nil {
		log.Fatal("Failed to start historic price service:", err)
	}
	if err := assetHistoricSvc.Start(); err != nil {
		log.Fatal("Failed to start asset historic service:", err)
	}

	r := gin.Default()
	r.Static("/static", "./static")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	h, err := handler.New(
		handler.WithEngine(r),
		handler.WithRepository(repository),
		handler.WithPriceChannel(sseCh),
		handler.WithPriceCache(priceCache),
		handler.WithAssetCreatedPublisher(assetHistoricSvc.Publisher()),
	)
	if err != nil {
		log.Fatal("Failed to create handler:", err)
	}
	if err := h.Setup(); err != nil {
		log.Fatal("Failed to setup routes:", err)
	}

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down...")
		cancel()
		livePriceSvc.Stop()
		historicPriceSvc.Stop()
		os.Exit(0)
	}()

	logger.Info("starting HodlBook", "port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
