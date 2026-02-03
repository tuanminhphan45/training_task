package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"task1/internal/dto"
	"task1/internal/model"
	"task1/internal/repository"
	"task1/internal/service"
)

type Handler struct {
	repo    repository.HashRepository
	crawler service.Crawler
}

func New(repo repository.HashRepository, crawler service.Crawler) *Handler {
	return &Handler{
		repo:    repo,
		crawler: crawler,
	}
}

func (h *Handler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	sourceFile := c.Query("source_file")

	if page < 1 {
		page = 1
	}

	hashes, count, err := h.repo.List(c.Request.Context(), page, size, sourceFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": hashes, "total": count, "page": page})
}

func (h *Handler) Get(c *gin.Context) {
	hash, err := h.repo.GetByMD5(c.Request.Context(), c.Param("hash"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, hash)
}

func (h *Handler) Create(c *gin.Context) {
	var input dto.CreateHashRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hash := &model.Hash{MD5Hash: input.MD5Hash, CreatedAt: time.Now()}
	if err := h.repo.Create(c.Request.Context(), hash); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, hash)
}

func (h *Handler) Stats(c *gin.Context) {
	count, _ := h.repo.Count(c.Request.Context())
	c.JSON(http.StatusOK, gin.H{"total": count})
}

func (h *Handler) Crawl(c *gin.Context) {
	progress, err := h.crawler.Start()
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "progress": progress})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "crawl started", "progress": progress})
}

func (h *Handler) CrawlStatus(c *gin.Context) {
	c.JSON(http.StatusOK, h.crawler.Status())
}
