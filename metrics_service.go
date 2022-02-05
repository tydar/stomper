package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type MetricsService struct {
	SentCount       uint64
	ReceivedCount   uint64
	ErrorCount      uint64
	serverStartTime time.Time
}

func NewMetricsService() *MetricsService {
	now := time.Now()
	return &MetricsService{
		SentCount:       0,
		ReceivedCount:   0,
		ErrorCount:      0,
		serverStartTime: now,
	}
}

func (ms *MetricsService) IncSent() {
	atomic.AddUint64(&ms.SentCount, uint64(1))
}

func (ms *MetricsService) AddSent(n uint64) {
	atomic.AddUint64(&ms.SentCount, n)
}

func (ms *MetricsService) IncReceived() {
	atomic.AddUint64(&ms.ReceivedCount, uint64(1))
}

func (ms *MetricsService) IncError() {
	atomic.AddUint64(&ms.ErrorCount, uint64(1))
}

func (ms *MetricsService) GetSentCount() uint64 {
	return atomic.LoadUint64(&ms.SentCount)
}

func (ms *MetricsService) GetReceivedCount() uint64 {
	return atomic.LoadUint64(&ms.ReceivedCount)
}

func (ms *MetricsService) GetErrorCount() uint64 {
	return atomic.LoadUint64(&ms.ErrorCount)
}

func (ms *MetricsService) GetServerStartTime() time.Time {
	// no need for atomic here bc it will not be manipulated after initialization
	return ms.serverStartTime
}

// currently nearly identical to MetricsService but its own type for clarity
// and to allow easier decoupling if needed in futrue
type metricsResponse struct {
	SentCount       uint64
	ReceivedCount   uint64
	ErrorCount      uint64
	ServerStartTime time.Time
	Timestamp       time.Time
}

func (ms *MetricsService) ListenAndServeJSON(addr string) {
	log.Printf("starting metrics server on %s\n", addr)
	metricsHandler := func(w http.ResponseWriter, r *http.Request) {
		respStruct := metricsResponse{
			SentCount:       ms.GetSentCount(),
			ReceivedCount:   ms.GetReceivedCount(),
			ErrorCount:      ms.GetErrorCount(),
			ServerStartTime: ms.GetServerStartTime(),
			Timestamp:       time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(respStruct)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/stomper", metricsHandler)
	http.ListenAndServe(addr, mux)
}
