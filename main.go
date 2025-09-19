package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type Location struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lat     float64 `json:"latitude"`
	Lng     float64 `json:"longitude"`
}

// Mock data for now
var locations = []Location{
	{1, "Popo Downtown", "123 Main St", 50.6829, -26.9890},
	{2, "Popito Suburbs", "234 Bitty St", 87.6987, 69.989},
}

func getLocations(w http.ResponseWriter, r *http.Request) {
	log.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func main() {
	http.HandleFunc("/locations", getLocations)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
