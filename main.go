package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// Bikebattery represents data about a bike battery.
type BikeBattery struct {
	ID            string  `json:"id"`
	Level         float32 `json:"level"`
	IsCharging    bool    `json:"is_charging"`
	ChargingSpeed float32 `json:"charging_speed"`
}

// ChargingStation represents data about a charging station.
type ChargingStation struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	BatteryLevel int     `json:"battery_level"`
}

var db *sql.DB

func chargeBike(id string) {
	// Get charging speed from database
	for {
		var level float32
		err := db.QueryRow("SELECT level FROM batteries WHERE id = ?", id).Scan(&level)
		if err != nil {
			return
		}

		if level >= 100 {
			db.Exec("UPDATE batteries SET is_charging = false WHERE id = ?", id)
			return
		}

		_, err = db.Exec("UPDATE batteries SET level = level + charging_speed WHERE id = ? AND is_charging = true", id)
		if err != nil {
			return
		}

		time.Sleep(time.Second)
	}
}

func postCharge(c *gin.Context) {
	// Put in charging mode a battery
	id := c.Param("id")

	var isCharging bool
	err := db.QueryRow("SELECT is_charging FROM batteries WHERE id = ?", id).Scan(&isCharging)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "battery not found"})
		return
	}

	if isCharging {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Charging already in progress."})
		return
	}

	_, err = db.Exec("UPDATE batteries SET is_charging = true WHERE id = ?", id)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error starting charging."})
		return
	}

	go chargeBike(id)
	c.IndentedJSON(http.StatusCreated, gin.H{"message": "Charging started."})
}

func deleteCharge(c *gin.Context) {
	// Stop charging a battery
	id := c.Param("id")

	var isCharging bool
	err := db.QueryRow("SELECT is_charging FROM batteries WHERE id = ?", id).Scan(&isCharging)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "battery not found"})
		return
	}

	if !isCharging {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Bike is not currently charging."})
		return
	}
	_, err = db.Exec("UPDATE batteries SET is_charging = false WHERE id = ?", id)
	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Charging stopped."})
		return
	}
}

func getBatteries(c *gin.Context) {
	// getBatteries responds with the list of all batteries as JSON.
	rows, err := db.Query("SELECT id, level, is_charging, charging_speed FROM batteries")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var batteries []BikeBattery
	for rows.Next() {
		var battery BikeBattery
		err := rows.Scan(&battery.ID, &battery.Level, &battery.IsCharging, &battery.ChargingSpeed)
		if err != nil {
			log.Fatal(err)
		}
		batteries = append(batteries, battery)
	}

	c.IndentedJSON(http.StatusOK, batteries)
}

func getBatteryByID(c *gin.Context) {
	// getBatteryByID locates the battery whose ID value matches the id
	// parameter sent by the client, then returns that battery as a response.

	id := c.Param("id")
	// looking for an battery whose ID value matches the parameter.
	var battery BikeBattery
	err := db.QueryRow("SELECT * FROM batteries WHERE id = ?", id).Scan(&battery.ID, &battery.Level, &battery.IsCharging, &battery.ChargingSpeed)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "battery not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, battery)
}

func updateBattery(c *gin.Context) {
	// updateBattery locates the battery whose ID value matches the id and updates the battery
	var battery BikeBattery

	// Call BindJSON to bind the received JSON to BikeBattery.
	if err := c.BindJSON(&battery); err != nil {
		return
	}

	// Add the new BikeBattery to the Database
	_, err := db.Exec("UPDATE batteries SET level = ?, is_charging = ?, charging_speed = ? WHERE id = ? ", battery.Level, battery.IsCharging, battery.ChargingSpeed, battery.ID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting battery."})
		return
	}
	c.IndentedJSON(http.StatusCreated, battery)
}

func postBattery(c *gin.Context) {
	// postAlbums adds an album from JSON received in the request body.
	var battery BikeBattery

	// Call BindJSON to bind the received JSON to BikeBattery.
	if err := c.BindJSON(&battery); err != nil {
		return
	}

	// Add the new BikeBattery to the Database
	_, err := db.Exec("INSERT INTO batteries (id, level, is_charging, charging_speed) VALUES (?, ?, ?, ?)", battery.ID, battery.Level, battery.IsCharging, battery.ChargingSpeed)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Error inserting battery."})
		return
	}
	c.IndentedJSON(http.StatusCreated, battery)
}

func removeBattery(c *gin.Context) {
	// remove battery whose ID value matches the id parameter sent by the client.

	id := c.Param("id")
	// looking for an battery whose ID value matches the parameter.
	var battery BikeBattery
	_, err := db.Exec("DELETE FROM batteries WHERE id = ?", id)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "battery not found"})
		return
	}
	c.IndentedJSON(http.StatusOK, battery)
}

func getChargingStations(c *gin.Context) {
	// Get all charging stations
	rows, err := db.Query("SELECT * FROM stations")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	stations := []ChargingStation{}
	for rows.Next() {
		var s ChargingStation
		err := rows.Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude, &s.BatteryLevel)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		stations = append(stations, s)
	}

	c.JSON(http.StatusOK, stations)
}

func getChargingStationByID(c *gin.Context) {
	// Get a charging station by ID
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	row := db.QueryRow("SELECT * FROM stations WHERE id=?", id)
	var s ChargingStation
	err = row.Scan(&s.ID, &s.Name, &s.Address, &s.Latitude, &s.Longitude, &s.BatteryLevel)
	if err == sql.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "Charging station not found"})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, s)
}

func main() {
	// Open the database connection
	// createDbConnexion()
	var err error
	db, err = sql.Open("mysql", "root:Hfosbj2554++@tcp(localhost:3306)/charging_stations")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	fmt.Println("Success to connect with DB!")

	// Start the router
	router := gin.Default()
	router.GET("/batteries", getBatteries)
	router.GET("/battery/:id", getBatteryByID)
	router.POST("/battery", postBattery)
	router.PUT("/battery/:id", updateBattery)
	router.DELETE("/battery/:id", removeBattery)
	router.POST("/charge/:id", postCharge)
	router.DELETE("/charge/:id", deleteCharge)
	router.GET("/charging-stations", getChargingStations)
	router.GET("/charging-stations/:id", getChargingStationByID)

	fmt.Println("Starting server on port 8080...")
	router.Run(":8080")
}
