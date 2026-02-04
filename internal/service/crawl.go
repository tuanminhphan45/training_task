package service

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"task1/internal/config"
	"task1/internal/model"
	"task1/internal/repository"
)

type CrawlService struct {
	repo repository.HashRepository
	cfg  *config.CrawlConfig

	importing bool
	progress  CrawlProgress
	mu        sync.RWMutex
}


func NewCrawlService(repo repository.HashRepository, cfg *config.CrawlConfig) Crawler {
	return &CrawlService{repo: repo, cfg: cfg}
}

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
	s.progress.Phase = "downloading"
	s.progress.Total = s.cfg.MaxFiles + 1
	s.mu.Unlock()

	sem := make(chan struct{}, s.cfg.MaxConcurrent)
	var downloaded int32

	for i := 0; i <= s.cfg.MaxFiles; i++ {
		name := fmt.Sprintf("VirusShare_%05d.md5", i)
		path := filepath.Join(s.cfg.OutDir, name)

		if _, err := os.Stat(path); err == nil {
			atomic.AddInt32(&downloaded, 1)
			s.mu.Lock()
			s.progress.Current = int(atomic.LoadInt32(&downloaded))
			s.mu.Unlock()
			continue
		}

		sem <- struct{}{}

		go func(fileName, filePath string) {
			defer func() { <-sem }()

			s.mu.Lock()
			s.progress.CurrentFile = fileName
			s.mu.Unlock()

			resp, err := http.Get(s.cfg.BaseURL + fileName)
			if err != nil || resp.StatusCode != 200 {
				if resp != nil {
					resp.Body.Close()
				}
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			os.WriteFile(filePath, body, 0644)

			atomic.AddInt32(&downloaded, 1)
			s.mu.Lock()
			s.progress.Current = int(atomic.LoadInt32(&downloaded))
			s.mu.Unlock()
			fmt.Println("downloaded", fileName)
		}(name, path)
	}
	fmt.Println("import")

	s.mu.Lock()
	s.progress.Phase = "importing"
	s.progress.Current = 0
	s.mu.Unlock()

	files, _ := filepath.Glob(filepath.Join(s.cfg.OutDir, "*.md5"))
	s.mu.Lock()
	s.progress.Total = len(files)
	s.mu.Unlock()

	md5Pattern := regexp.MustCompile(`^[a-fA-F0-9]{32}$`)
	var totalImported int64

	for i, filePath := range files {
		s.mu.Lock()
		s.progress.Current = i + 1
		s.progress.CurrentFile = filepath.Base(filePath)
		s.mu.Unlock()

		imported := s.importFile(filePath, md5Pattern)
		atomic.AddInt64(&totalImported, imported)

		s.mu.Lock()
		s.progress.Imported = atomic.LoadInt64(&totalImported)
		s.mu.Unlock()
	}
}

func (s *CrawlService) importFile(filePath string, md5Pattern *regexp.Regexp) int64 {
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
		if line == "" || strings.HasPrefix(line, "#") || !md5Pattern.MatchString(line) {
			continue
		}

		batch = append(batch, &model.Hash{
			MD5Hash:    strings.ToLower(line),
			SourceFile: sourceFile,
			CreatedAt:  time.Now(),
		})

		if len(batch) >= s.cfg.BatchSize {
			s.repo.CreateBatch(context.Background(), batch)
			imported += int64(len(batch))
			batch = nil
		}
	}

	if len(batch) > 0 {
		s.repo.CreateBatch(context.Background(), batch)
		imported += int64(len(batch))
	}

	return imported
}
