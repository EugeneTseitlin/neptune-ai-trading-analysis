package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	
	r.HandleFunc("/add_batch", AddBatchHandler).Methods("POST")
	r.HandleFunc("/stats", StatsHandler).Methods("GET")

	log.Println("Starting server")
	if err := http.ListenAndServe(":8484", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}