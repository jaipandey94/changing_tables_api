package main

import (
	"encoding/json"
	"log"
	"net/http"
	// "strconv"
	// "strings"
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

func createLocation(w http.ResponseWriter, r *http.Request) {
	log.Println("POST request received")

	var newLocation Location
	err := json.NewDecoder(r.Body).Decode(&newLocation)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded location: %+v\n", newLocation)

	// Generate a simple ID (in real app, database would do this)
	newLocation.ID = len(locations) + 1
	locations = append(locations, newLocation)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newLocation)
}

func handleLocations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getLocations(w, r)
	case http.MethodPost:
		createLocation(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/locations", handleLocations)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
