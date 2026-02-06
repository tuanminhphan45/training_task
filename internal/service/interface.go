package service

type Crawler interface {
	Start() (CrawlProgress, error)
	Status() CrawlProgress
}

type CrawlProgress struct {
	IsRunning   bool   `json:"is_running"`
	Phase       string `json:"phase"`
	Total       int    `json:"total"`
	Current     int    `json:"current"`
	Imported    int64  `json:"imported"`
	CurrentFile string `json:"current_file"`
}

