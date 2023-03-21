package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// prepare the test database
func TestMain(m *testing.M) {
	// Open a connection to the database.
	var err error
	db, err = sql.Open("mysql", "root:Hfosbj2554++@tcp(localhost:3306)/charging_stations_test")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	// Check if the "batteries" table exists
	rows, err := db.Query("SELECT * FROM batteries LIMIT 1")
	if err != nil {
		// If the table doesn't exist, create it
		_, err := db.Exec(`CREATE TABLE batteries (id VARCHAR(50) NOT NULL PRIMARY KEY,
				level FLOAT(5, 2) NOT NULL,is_charging BOOL NOT NULL,charging_speed FLOAT(5, 2) NOT NULL)`)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// If the table exists, remove all data from it
		_, err := db.Exec("DELETE FROM batteries")
		if err != nil {
			panic(err.Error())
		}
		rows.Close()
	}

	// Insert some test data into the database.
	_, err = db.Exec("INSERT INTO batteries (id, level, is_charging, charging_speed) VALUES ('BBP_test', 50, false, 0.5)")
	if err != nil {
		panic(err.Error())
	}

	// Run the tests.
	m.Run()
}

func TestGetBatteryByID(t *testing.T) {

	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/battery/:id", getBatteryByID)

	// Create a mock request to get a battery with ID BBP_test
	req, err := http.NewRequest(http.MethodGet, "/battery/BBP_test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock response recorder
	res := httptest.NewRecorder()

	// Call the getBatteryByID function
	router.ServeHTTP(res, req)

	// Check the response status code is OK (200)
	assert.Equal(t, http.StatusOK, res.Code)

	// Check that the response body contains the expected battery with ID 1
	expectedBody := "{\n    \"id\": \"BBP_test\",\n    \"level\": 50,\n    \"is_charging\": false,\n    \"charging_speed\": 0.5\n}"
	assert.Equal(t, expectedBody, res.Body.String())
}

func TestGetBatteries(t *testing.T) {
	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/batteries", getBatteries)

	// Create a mock request to get all batteries
	req, err := http.NewRequest(http.MethodGet, "/batteries", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock response recorder
	res := httptest.NewRecorder()

	// Call the getBatteries function
	router.ServeHTTP(res, req)

	// Check the response status code is OK (200)
	assert.Equal(t, http.StatusOK, res.Code)

	// Check that the response body contains the expected batteries
	expectedBody := "[\n    {\n        \"id\": \"BBP_test\",\n        \"level\": 50,\n        \"is_charging\": false,\n        \"charging_speed\": 0.5\n    }\n]"
	assert.Equal(t, expectedBody, res.Body.String())
}

func TestRemoveBattery(t *testing.T) {

	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/battery/:id", removeBattery)

	// Create a mock request to remove a battery with ID BBP_test
	req, err := http.NewRequest(http.MethodDelete, "/battery/BBP_test", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a mock response recorder
	res := httptest.NewRecorder()

	// Call the removeBattery function
	router.ServeHTTP(res, req)

	// Check the response status code is OK (200)
	assert.Equal(t, http.StatusOK, res.Code)

	// Check the response body
	expectedBody := "{\n    \"id\": \"\",\n    \"level\": 0,\n    \"is_charging\": false,\n    \"charging_speed\": 0\n}"
	assert.Equal(t, expectedBody, res.Body.String())

	// Check that the battery was removed from the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM batteries WHERE id = 'BBP_test'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestPostBattery(t *testing.T) {

	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/battery", postBattery)

	// Create a new battery to add to the database
	battery := BikeBattery{
		ID:            "BBP_new_test",
		Level:         10,
		IsCharging:    true,
		ChargingSpeed: 5,
	}

	// Create a request body from the battery struct
	jsonBody, err := json.Marshal(battery)
	if err != nil {
		t.Fatal(err)
	}

	// Set up the request with the JSON request body
	req, err := http.NewRequest(http.MethodPost, "/battery", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Create a mock response recorder
	res := httptest.NewRecorder()

	// Call the postBattery function
	router.ServeHTTP(res, req)

	// Check the response status code
	assert.Equal(t, http.StatusCreated, res.Code)

	// Check the response body
	var responseBattery BikeBattery
	err = json.Unmarshal(res.Body.Bytes(), &responseBattery)
	require.NoError(t, err)
	assert.Equal(t, battery, responseBattery)

	// Check that the battery was created in the database
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM batteries WHERE id = 'BBP_new_test'").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestIntegrate(t *testing.T) {

	// Set up a test router
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/battery", postBattery)
	router.GET("/battery/:id", getBatteryByID)
	router.PUT("/battery/:id", updateBattery)
	router.DELETE("/battery/:id", removeBattery)
	router.POST("/charge/:id", postCharge)
	router.DELETE("/charge/:id", deleteCharge)

	// create new battery in the DB
	battery := BikeBattery{
		ID:            "BBP_integrate_test",
		Level:         10,
		IsCharging:    false,
		ChargingSpeed: 5,
	}
	jsonBody, err := json.Marshal(battery)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest(http.MethodPost, "/battery", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusCreated, res.Code)

	// get battery from DB
	req, err = http.NewRequest(http.MethodGet, "/battery/BBP_integrate_test", nil)
	if err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	// modify a value of the battery
	battery = BikeBattery{
		ID:            "BBP_integrate_test",
		Level:         20,
		IsCharging:    false,
		ChargingSpeed: 5,
	}
	jsonBody, err = json.Marshal(battery)
	if err != nil {
		t.Fatal(err)
	}
	req, err = http.NewRequest(http.MethodPut, "/battery/BBP_integrate_test", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusCreated, res.Code)

	// request to charge batterry
	req, err = http.NewRequest(http.MethodPost, "/charge/BBP_integrate_test", nil)
	if err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusCreated, res.Code)

	// wait 5 seconds and verfify if charge increased
	time.Sleep(5 * time.Second)
	req, err = http.NewRequest(http.MethodGet, "/battery/BBP_integrate_test", nil)
	if err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
	var responseBattery BikeBattery
	err = json.Unmarshal(res.Body.Bytes(), &responseBattery)
	require.NoError(t, err)
	assert.Equal(t, float32(45), responseBattery.Level)

	// stop charge and verifiy
	req, err = http.NewRequest(http.MethodDelete, "/charge/BBP_integrate_test", nil)
	if err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)

	// remove battery from DB and verify
	req, err = http.NewRequest(http.MethodDelete, "/battery/BBP_integrate_test", nil)
	if err != nil {
		t.Fatal(err)
	}
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	assert.Equal(t, http.StatusOK, res.Code)
}
