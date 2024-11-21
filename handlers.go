package main

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var tradeAnalysisEngine = NewTradeAnalysisEngine()

func AddBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string    `json:"symbol"`
		Values []float64 `json:"values"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if len(req.Values) > 10000 {
		http.Error(w, "Batch exceeds maximum size of 10,000", http.StatusBadRequest)
		return
	}

	tradeAnalysisEngine.AddBatch(req.Symbol, req.Values)
	w.WriteHeader(http.StatusOK)
}

func StatsHandler(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	k, err := strconv.Atoi(r.URL.Query().Get("k"))
	if err != nil || k < 1 || k > 8 {
		http.Error(w, "Invalid k value", http.StatusBadRequest)
		return
	}

	stats := tradeAnalysisEngine.GetStats(symbol, k)
	json.NewEncoder(w).Encode(stats)
}