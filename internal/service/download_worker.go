package service

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
)

type DownloadJob struct {
	FileName string
	FilePath string
}

type DownloadWorker struct {
	baseURL      string
	downloadedCh chan<- string
	downloaded   *int32
	service      *CrawlService
}

func NewDownloadWorker(baseURL string, downloadedCh chan<- string, downloaded *int32, service *CrawlService) *DownloadWorker {
	return &DownloadWorker{
		baseURL:      baseURL,
		downloadedCh: downloadedCh,
		downloaded:   downloaded,
		service:      service,
	}
}

func (w *DownloadWorker) Process(job DownloadJob) error {
	w.service.mu.Lock()
	w.service.progress.CurrentFile = job.FileName
	w.service.mu.Unlock()

	url := fmt.Sprintf("%s%s", w.baseURL, job.FileName)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp != nil {
			resp.Body.Close()
		}
		return err
	}

	out, err := os.Create(job.FilePath)
	if err != nil {
		resp.Body.Close()
		return err
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	resp.Body.Close()

	if err != nil {
		os.Remove(job.FilePath)
		return err
	}

	atomic.AddInt32(w.downloaded, 1)
	w.service.mu.Lock()
	w.service.progress.Current = int(atomic.LoadInt32(w.downloaded))
	w.service.mu.Unlock()

	w.downloadedCh <- job.FilePath
	return nil
}
