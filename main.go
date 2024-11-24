package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/EugeneTseitlin/neptune-ai-trading-analysis/tradinganalysis"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
)


var tradeAnalysisEngine = tradinganalysis.NewTradeAnalysisEngine()

func AddBatchHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Symbol string    `json:"symbol"`
		Values []decimal.Decimal `json:"values"`
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

	stats, err := tradeAnalysisEngine.GetStats(symbol, k)
	if err != nil {
		if errors.Is(err, tradinganalysis.ErrSymbolNotFound) {
			http.NotFound(w, r)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		json.NewEncoder(w).Encode(stats)
	}
}

func main() {
	r := mux.NewRouter()
	
	r.HandleFunc("/add_batch", AddBatchHandler).Methods("POST")
	r.HandleFunc("/stats", StatsHandler).Methods("GET")

	log.Println("Starting server")
	if err := http.ListenAndServe(":8484", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}