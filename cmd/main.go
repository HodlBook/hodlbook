package main

import (
	"log"
	"os"

	"hodlbook/internal/handler"
	"hodlbook/internal/repo"
	"hodlbook/pkg/database"
	"hodlbook/pkg/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	utils.LoadEnv()

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

	r := gin.Default()
	r.Static("/static", "./static")

	h, err := handler.New(r, repository)
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

	log.Printf("Starting HodlBook on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
