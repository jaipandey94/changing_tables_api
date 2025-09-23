package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Location struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lat     float64 `json:"latitude"`
	Lng     float64 `json:"longitude"`
}

var db *sql.DB

// handleError provides unified error handling and logging
func handleError(w http.ResponseWriter, err error, message string, statusCode int) {
	log.Printf("%s: %v", message, err)
	http.Error(w, message, statusCode)
}

// handleDatabaseError handles database-specific errors
func handleDatabaseError(w http.ResponseWriter, err error, operation string) {
	if err == sql.ErrNoRows {
		handleError(w, err, "Location not found", http.StatusNotFound)
		return
	}
	handleError(w, err, "Database error", http.StatusInternalServerError)
}

// Initialize database connection
func initDB() {
	var err error
	//Connection string
	connStr := "user=jaipandey dbname=changing_tables sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatal(("Failed to ping database"))
	}

	log.Println("Successfully connected to the database!")
}

// // Mock data for now
// var locations = []Location{
// 	{1, "Popo Downtown", "123 Main St", 50.6829, -26.9890},
// 	{2, "Popito Suburbs", "234 Bitty St", 87.6987, 69.989},
// }

func getLocations(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("Select id, name, address, latitude, longitude FROM locations ORDER by id")
	if err != nil {
		handleDatabaseError(w, err, "query locations")
		return
	}
	defer rows.Close()

	var locations []Location
	for rows.Next() {
		var loc Location
		err := rows.Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Lat, &loc.Lng)
		if err != nil {
			log.Printf("Row scan error: %v", err)
			continue
		}
		locations = append(locations, loc)
	}

	log.Printf("Method: %s, URL: %s", r.Method, r.URL.Path)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(locations)
}

func getLocationByID(w http.ResponseWriter, r *http.Request, id int) {
	var loc Location
	err := db.QueryRow(
		"SELECT id, name, address, latitude, longitude FROM locations WHERE id = $1", id).Scan(&loc.ID, &loc.Name, &loc.Address, &loc.Lat, &loc.Lng)

	if err != nil {
		handleDatabaseError(w, err, "get location by ID")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loc)
}

func updateLocation(w http.ResponseWriter, r *http.Request, id int) {
	var updatedLocation Location
	err := json.NewDecoder(r.Body).Decode(&updatedLocation)
	if err != nil {
		handleError(w, err, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Update in Database
	result, err := db.Exec(
		"UPDATE locations SET name = $1, address = $2, latitude = $3, longitude = $4 WHERE id = $5", updatedLocation.Name, updatedLocation.Address, updatedLocation.Lat, updatedLocation.Lng, id)

	if err != nil {
		handleDatabaseError(w, err, "update location")
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		handleError(w, err, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Location not found", http.StatusNotFound)
	}

	// Return the updated location with the correct ID
	updatedLocation.ID = id
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedLocation)
}

func deleteLocation(w http.ResponseWriter, r *http.Request, id int) {
	result, err := db.Exec("DELETE FROM locations WHERE id = $1", id)

	if err != nil {
		handleDatabaseError(w, err, "delete location")
		return
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		handleError(w, err, "Database error", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Location not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func createLocation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Println("POST request received")

	var newLocation Location
	err := json.NewDecoder(r.Body).Decode(&newLocation)
	if err != nil {
		handleError(w, err, "Invalid JSON", http.StatusBadRequest)
		return
	}

	log.Printf("Decoded location: %+v\n", newLocation)

	// // Generate a simple ID (in real app, database would do this)
	// newLocation.ID = len(locations) + 1

	err = db.QueryRow("INSERT INTO locations (name, address, latitude, longitude) VALUES ($1, $2, $3, $4) returning id", newLocation.Name, newLocation.Address, newLocation.Lat, newLocation.Lng).Scan(&newLocation.ID)

	if err != nil {
		handleDatabaseError(w, err, "create location")
		return
	}

	log.Printf("Created location with ID: %d", newLocation.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newLocation)
}

func handleLocations(w http.ResponseWriter, r *http.Request) {

	//Parse URL path to extract ID if present
	path := strings.TrimPrefix(r.URL.Path, "/locations")


	if path == "" || path == "/" {
		switch r.Method {
		case http.MethodGet:
			getLocations(w, r)
		case http.MethodPost:
			createLocation(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}

	// Handle /locations/{id}
	idStr := strings.Trim(path, "/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid location ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodGet:
		getLocationByID(w, r, id)
	case http.MethodPut:
		updateLocation(w, r, id)
	case http.MethodDelete:
		deleteLocation(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func main() {
	initDB()
	defer db.Close()
	http.HandleFunc("/locations", handleLocations)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
