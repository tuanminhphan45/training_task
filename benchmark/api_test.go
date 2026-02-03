package benchmark

import (
	"net/http"
	"testing"
)

func BenchmarkListHashes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/api/v1/hashes?page=1&size=10")
	}
}

func BenchmarkGetStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/api/v1/stats")
	}
}

func BenchmarkGetByMD5(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/api/v1/hashes/de960b8f99230e94858204b711fb8a38")
	}
}

func BenchmarkListBySourceFile(b *testing.B) {
	for i := 0; i < b.N; i++ {
		http.Get("http://localhost:8080/api/v1/hashes?source_file=VirusShare_00000&size=10")
	}
}

// Parallel benchmarks (concurrent requests)
func BenchmarkListHashesParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get("http://localhost:8080/api/v1/hashes?page=1&size=10")
		}
	})
}

func BenchmarkGetStatsParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get("http://localhost:8080/api/v1/stats")
		}
	})
}

func BenchmarkGetByMD5Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get("http://localhost:8080/api/v1/hashes/de960b8f99230e94858204b711fb8a38")
		}
	})
}

func BenchmarkListBySourceFileParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			http.Get("http://localhost:8080/api/v1/hashes?source_file=VirusShare_00000&size=10")
		}
	})
}
