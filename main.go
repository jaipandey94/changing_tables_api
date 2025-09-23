package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Location struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Address  string  `json:"address"`
	Lat      float64 `json:"latitude"`
	Lng      float64 `json:"longitude"`
	Distance float64 `json:"distance,omitempty"` // Only included when searching by distance
}

var db *sql.DB

func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 3958.756 //Earth's radius in miles

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

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

	query := r.URL.Query()
	search := query.Get("search")
	nearParam := query.Get("near")     // Format: "lat, lng"
	radiusParam := query.Get("radius") // In miles
	cityParam := query.Get("city")

	//Build SQL Query
	var sqlQuery string
	var args []interface{}
	var argCount int
	var hasWhere bool

	//Base query
	sqlQuery = "SELECT id, name, address, latitude, longitude FROM locations"

	// Add search filter
	if search != "" {
		if !hasWhere {
			sqlQuery += " WHERE"
			hasWhere = true
		} else {
			sqlQuery += " AND"
		}
		argCount++
		searchPattern := "%" + search + "%"
		sqlQuery += fmt.Sprintf(" (name ILIKE $%d OR address ILIKE $%d)", argCount, argCount+1)
		args = append(args, searchPattern, searchPattern)
		argCount++ // Increment again since we used two parameters
	}

	// Add City Filter
	if cityParam != "" {
		if !hasWhere {
			sqlQuery += " WHERE"
			hasWhere = true
		} else {
			sqlQuery += " AND"
		}
		argCount++
		sqlQuery += fmt.Sprintf(" address ILIKE $%d", argCount)
		args = append(args, "%"+cityParam+"%")
	}

	// Parse location-based search
	var searchLat, searchLng, radiusMiles float64
	var hasLocationSearch bool
	if nearParam != "" {
		coords := strings.Split(nearParam, ",")
		if len(coords) == 2 {
			if lat, err := strconv.ParseFloat(strings.TrimSpace(coords[0]), 64); err == nil {
				if lng, err := strconv.ParseFloat(strings.TrimSpace(coords[1]), 64); err == nil {
					searchLat = lat
					searchLng = lng
					hasLocationSearch = true

					// Parse radius (default to 10 miles)
					radiusMiles = 10
					if radiusParam != "" {
						if r, err := strconv.ParseFloat(radiusParam, 64); err == nil {
							radiusMiles = r
						}
					}

				}
			}
		}
	}

	sqlQuery += " ORDER BY id"

	// Execute the query
	log.Printf("SQL Query: %s", sqlQuery)
	log.Printf("Parameters: %v", args)
	rows, err := db.Query(sqlQuery, args...)
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

		// If doing location-based search, calculate distance and filter
		if hasLocationSearch {
			distance := calculateDistance(searchLat, searchLng, loc.Lat, loc.Lng)
			if distance <= radiusMiles {
				loc.Distance = math.Round(distance*100) / 100 // Round to 2 decimal places
				locations = append(locations, loc)
			}
		} else {
			locations = append(locations, loc)
		}
	}

	// Sort by distance if doing location-based search
	if hasLocationSearch {
		// Simple bubble sort by distance (for small datasets)
		for i := 0; i < len(locations)-1; i++ {
			for j := 0; j < len(locations)-i-1; j++ {
				if locations[j].Distance > locations[j+1].Distance {
					locations[j], locations[j+1] = locations[j+1], locations[j]
				}
			}
		}
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
		return
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
	http.HandleFunc("/locations/", handleLocations)
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
