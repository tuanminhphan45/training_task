package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"task1/internal/config"
	"task1/internal/db"
	"task1/internal/handler"
	"task1/internal/repository"
	"task1/internal/service"
)

func main() {
	cfg := config.Load()

	database := db.Init(&cfg.DB)
	defer database.Close()

	repo := repository.NewHashRepository(database)
	crawler := service.NewCrawlService(repo, &cfg.Crawl)
	h := handler.New(repo, crawler)

	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api/v1")
	{
		api.GET("/hashes", h.List)
		api.GET("/hashes/:hash", h.Get)
		api.POST("/hashes", h.Create)
		api.GET("/stats", h.Stats)
		api.POST("/crawl", h.Crawl)
		api.GET("/crawl/status", h.CrawlStatus)
	}

	log.Printf("Server running on :%s", cfg.Server.Port)
	r.Run(":" + cfg.Server.Port)
}
