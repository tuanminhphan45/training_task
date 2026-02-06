package service

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"task1/internal/model"
	"task1/internal/repository"
)

type ImportJob struct {
	FilePath string
}

type ImportWorker struct {
	repo           repository.HashRepository
	batchSize      int
	totalImported  *int64
	service        *CrawlService
	md5Pattern     *regexp.Regexp
}

func NewImportWorker(repo repository.HashRepository, batchSize int, totalImported *int64, service *CrawlService, md5Pattern *regexp.Regexp) *ImportWorker {
	return &ImportWorker{
		repo:          repo,
		batchSize:     batchSize,
		totalImported: totalImported,
		service:       service,
		md5Pattern:    md5Pattern,
	}
}

func (w *ImportWorker) Process(job ImportJob) error {
	w.service.mu.Lock()
	w.service.progress.CurrentFile = fmt.Sprintf("%s (importing)", filepath.Base(job.FilePath))
	w.service.mu.Unlock()

	imported := w.importFile(job.FilePath)
	atomic.AddInt64(w.totalImported, imported)

	w.service.mu.Lock()
	w.service.progress.Imported = atomic.LoadInt64(w.totalImported)
	w.service.mu.Unlock()

	return nil
}

func (w *ImportWorker) importFile(filePath string) int64 {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	sourceFile := strings.TrimSuffix(filepath.Base(filePath), ".md5")

	var batch []*model.Hash
	var imported int64

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !w.md5Pattern.MatchString(line) {
			continue
		}

		batch = append(batch, &model.Hash{
			MD5Hash:    strings.ToLower(line),
			SourceFile: sourceFile,
			CreatedAt:  time.Now(),
		})

		if len(batch) >= w.batchSize {
			w.repo.CreateBatch(context.Background(), batch)
			imported += int64(len(batch))
			batch = nil
		}
	}

	if len(batch) > 0 {
		w.repo.CreateBatch(context.Background(), batch)
		imported += int64(len(batch))
	}

	return imported
}
