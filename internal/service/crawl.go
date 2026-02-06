package service

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"

	"task1/internal/config"
	"task1/internal/repository"
)

type CrawlService struct {
	repo      repository.HashRepository
	cfg       *config.CrawlConfig
	importing bool
	progress  CrawlProgress
	mu        sync.RWMutex
}

var _ Crawler = (*CrawlService)(nil)

func NewCrawlService(repo repository.HashRepository, cfg *config.CrawlConfig) Crawler {
	return &CrawlService{repo: repo, cfg: cfg}
}

var md5Pattern = regexp.MustCompile(`^[a-fA-F0-9]{32}$`)

func (s *CrawlService) Start() (CrawlProgress, error) {
	s.mu.Lock()
	if s.importing {
		s.mu.Unlock()
		return s.progress, fmt.Errorf("crawl already in progress")
	}
	s.importing = true
	s.progress = CrawlProgress{IsRunning: true, Phase: "starting", Total: s.cfg.MaxFiles + 1}
	s.mu.Unlock()

	go s.doCrawlAndImport()

	return s.progress, nil
}

func (s *CrawlService) Status() CrawlProgress {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.progress
}

func (s *CrawlService) doCrawlAndImport() {
	defer func() {
		s.mu.Lock()
		s.importing = false
		s.progress.IsRunning = false
		s.progress.Phase = "done"
		s.mu.Unlock()
	}()

	os.MkdirAll(s.cfg.OutDir, 0755)

	s.mu.Lock()
	s.progress.Phase = "processing"
	s.progress.Total = s.cfg.MaxFiles + 1
	s.mu.Unlock()

	downloadedCh := make(chan string, s.cfg.MaxConcurrent)
	var downloaded int32
	var totalImported int64

	downloadWorker := NewDownloadWorker(s.cfg.BaseURL, downloadedCh, &downloaded, s)
	downloadPool := NewWorkerPool(s.cfg.MaxConcurrent, downloadWorker)
	downloadPool.Start()

	importWorker := NewImportWorker(s.repo, s.cfg.BatchSize, &totalImported, s, md5Pattern)
	importPool := NewWorkerPool(s.cfg.MaxImportWorkers, importWorker)
	importPool.Start()

	go func() {
		defer downloadPool.Close()
		for i := 0; i <= s.cfg.MaxFiles; i++ {
			name := fmt.Sprintf("VirusShare_%05d.md5", i)
			path := filepath.Join(s.cfg.OutDir, name)

			if _, err := os.Stat(path); err == nil {
				atomic.AddInt32(&downloaded, 1)
				s.mu.Lock()
				s.progress.Current = int(atomic.LoadInt32(&downloaded))
				s.mu.Unlock()
				downloadedCh <- path
				continue
			}

			downloadPool.Submit(DownloadJob{
				FileName: name,
				FilePath: path,
			})
		}
	}()

	go func() {
		downloadPool.Wait()
		close(downloadedCh)
	}()

	go func() {
		defer importPool.Close()
		for filePath := range downloadedCh {
			importPool.Submit(ImportJob{FilePath: filePath})
		}
	}()

	importPool.Wait()
}
