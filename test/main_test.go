package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/krushnna/meeting-scheduler/models"
	"github.com/krushnna/meeting-scheduler/routers"
	"github.com/krushnna/meeting-scheduler/utils"
)

// init is called before tests run.
func init() {
	gin.SetMode(gin.TestMode)
	// Initialize Zap logger for tests (if not already initialized)
	utils.InitLogger()
}

// setupTestRouter creates an in-memory DB, auto-migrates models,
// and returns a test router.
func setupTestRouter() (*gin.Engine, *gorm.DB) {
	// Use in-memory SQLite for testing.
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect test database")
	}

	// Auto-migrate all models.
	err = db.AutoMigrate(
		&models.Event{},
		&models.TimeSlot{},
		&models.User{},
		&models.UserAvailability{},
	)
	if err != nil {
		panic("failed to migrate test database")
	}

	logger := utils.GetLogger()
	router := routers.SetupRouter(db, logger)
	return router, db
}

// TestHealthEndpoint verifies the /health endpoint.
func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	req, _ := http.NewRequest("GET", "/health", nil)
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", resp.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("Error unmarshalling response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("Expected health status 'ok', got %v", body["status"])
	}
}

// TestEventEndpoints tests create, get, update, and delete event endpoints.
func TestEventEndpoints(t *testing.T) {
	router, _ := setupTestRouter()

	// Create Event
	eventPayload := map[string]interface{}{
		"title":            "Test Event",
		"description":      "Test Description",
		"organizer_id":     1,
		"duration_minutes": 60,
	}
	jsonPayload, _ := json.Marshal(eventPayload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(jsonPayload))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Expected 201 on event creation, got %d", resp.Code)
	}

	var createdEvent models.Event
	if err := json.Unmarshal(resp.Body.Bytes(), &createdEvent); err != nil {
		t.Fatalf("Error unmarshalling created event: %v", err)
	}

	// Get Event
	req, _ = http.NewRequest("GET", "/api/v1/events/"+strconv.Itoa(int(createdEvent.ID)), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on get event, got %d", resp.Code)
	}

	// Update Event
	updatePayload := map[string]interface{}{
		"title":            "Updated Test Event",
		"description":      "Updated Description",
		"organizer_id":     1,
		"duration_minutes": 90,
	}
	jsonUpdate, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest("PUT", "/api/v1/events/"+strconv.Itoa(int(createdEvent.ID)), bytes.NewBuffer(jsonUpdate))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on update event, got %d", resp.Code)
	}

	// Delete Event
	req, _ = http.NewRequest("DELETE", "/api/v1/events/"+strconv.Itoa(int(createdEvent.ID)), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on delete event, got %d", resp.Code)
	}
}

// TestTimeSlotEndpoints tests creating and retrieving timeslots for an event.
func TestTimeSlotEndpoints(t *testing.T) {
	router, _ := setupTestRouter()

	// First create an event to attach timeslots
	eventPayload := map[string]interface{}{
		"title":            "TimeSlot Test Event",
		"description":      "Test timeslot event",
		"organizer_id":     1,
		"duration_minutes": 60,
	}
	eventJSON, _ := json.Marshal(eventPayload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	var event models.Event
	if err := json.Unmarshal(resp.Body.Bytes(), &event); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}

	// Create a TimeSlot for the event
	startTime := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(26 * time.Hour).Format(time.RFC3339)
	timeslotPayload := map[string]interface{}{
		"start_time": startTime,
		"end_time":   endTime,
	}
	timeslotJSON, _ := json.Marshal(timeslotPayload)
	req, _ = http.NewRequest("POST", "/api/v1/events/"+strconv.Itoa(int(event.ID))+"/timeslots", bytes.NewBuffer(timeslotJSON))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Expected 201 on timeslot creation, got %d", resp.Code)
	}

	// Retrieve all TimeSlots for the event
	req, _ = http.NewRequest("GET", "/api/v1/events/"+strconv.Itoa(int(event.ID))+"/timeslots", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on get timeslots, got %d", resp.Code)
	}
}

// TestUserEndpoints tests creating, retrieving, updating, and deleting a user.
func TestUserEndpoints(t *testing.T) {
	router, _ := setupTestRouter()

	// Create User
	userPayload := map[string]interface{}{
		"name":     "Test User",
		"email":    "testuser@example.com",
		"timezone": "UTC",
	}
	userJSON, _ := json.Marshal(userPayload)
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Expected 201 on user creation, got %d", resp.Code)
	}
	var user models.User
	if err := json.Unmarshal(resp.Body.Bytes(), &user); err != nil {
		t.Fatalf("Error unmarshalling user: %v", err)
	}

	// Get User
	req, _ = http.NewRequest("GET", "/api/v1/users/"+strconv.Itoa(int(user.ID)), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on get user, got %d", resp.Code)
	}

	// Update User
	updatePayload := map[string]interface{}{
		"name":     "Updated Test User",
		"email":    "updateduser@example.com",
		"timezone": "UTC",
	}
	updateJSON, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest("PUT", "/api/v1/users/"+strconv.Itoa(int(user.ID)), bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on update user, got %d", resp.Code)
	}

	// Delete User
	req, _ = http.NewRequest("DELETE", "/api/v1/users/"+strconv.Itoa(int(user.ID)), nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on delete user, got %d", resp.Code)
	}
}

// TestAvailabilityEndpoints tests creating and retrieving availability for a user and event.
func TestAvailabilityEndpoints(t *testing.T) {
	router, _ := setupTestRouter()

	// Create an event for availability testing.
	eventPayload := map[string]interface{}{
		"title":            "Availability Test Event",
		"description":      "Event for testing user availability",
		"organizer_id":     1,
		"duration_minutes": 60,
	}
	eventJSON, _ := json.Marshal(eventPayload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	var event models.Event
	if err := json.Unmarshal(resp.Body.Bytes(), &event); err != nil {
		t.Fatalf("Error unmarshalling event: %v", err)
	}

	// Create a user for availability.
	userPayload := map[string]interface{}{
		"name":     "Availability User",
		"email":    "availability@example.com",
		"timezone": "UTC",
	}
	userJSON, _ := json.Marshal(userPayload)
	req, _ = http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	var user models.User
	if err := json.Unmarshal(resp.Body.Bytes(), &user); err != nil {
		t.Fatalf("Error unmarshalling user: %v", err)
	}

	// Create Availability for the user and event.
	availPayload := map[string]interface{}{
		"start_time": time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"end_time":   time.Now().Add(4 * time.Hour).Format(time.RFC3339),
	}
	availJSON, _ := json.Marshal(availPayload)
	req, _ = http.NewRequest("POST", "/api/v1/users/"+strconv.Itoa(int(user.ID))+"/events/"+strconv.Itoa(int(event.ID))+"/availability", bytes.NewBuffer(availJSON))
	req.Header.Set("Content-Type", "application/json")
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusCreated {
		t.Errorf("Expected 201 on creating availability, got %d", resp.Code)
	}

	// Retrieve Availability for the user and event.
	req, _ = http.NewRequest("GET", "/api/v1/users/"+strconv.Itoa(int(user.ID))+"/events/"+strconv.Itoa(int(event.ID))+"/availability", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on retrieving availability, got %d", resp.Code)
	}
}


func TestRecommendationEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	// Create Event
	eventPayload := map[string]interface{}{
		"title":            "Recommendation Test Event",
		"organizer_id":     1,
		"duration_minutes": 60,
	}
	eventJSON, _ := json.Marshal(eventPayload)
	req, _ := http.NewRequest("POST", "/api/v1/events", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	var event models.Event
	json.Unmarshal(resp.Body.Bytes(), &event)

	// Create Time Slot
	start := time.Now().Add(24 * time.Hour)
	end := start.Add(2 * time.Hour)
	timeslotJSON, _ := json.Marshal(map[string]interface{}{
		"start_time": start.Format(time.RFC3339),
		"end_time":   end.Format(time.RFC3339),
	})
	req, _ = http.NewRequest("POST", "/api/v1/events/"+strconv.Itoa(int(event.ID))+"/timeslots", bytes.NewBuffer(timeslotJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), req)

	// Create Users and Availability
	user1 := createTestUser(router, "user1@test.com")
	user2 := createTestUser(router, "user2@test.com")
	createAvailability(router, user1.ID, event.ID, start.Add(30*time.Minute), end.Add(-30*time.Minute))
	createAvailability(router, user2.ID, event.ID, start, end)

	// Get Recommendations
	req, _ = http.NewRequest("GET", "/api/v1/events/"+strconv.Itoa(int(event.ID))+"/recommendations", nil)
	resp = httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected 200 on recommendations, got %d", resp.Code)
	}

	var recommendations []models.TimeSlotRecommendation
	json.Unmarshal(resp.Body.Bytes(), &recommendations)
	if len(recommendations) == 0 {
		t.Error("Expected at least one recommendation")
	}
}

func createTestUser(router *gin.Engine, email string) models.User {
	userJSON, _ := json.Marshal(map[string]interface{}{
		"name":     "Test User",
		"email":    email,
		"timezone": "UTC",
	})
	req, _ := http.NewRequest("POST", "/api/v1/users", bytes.NewBuffer(userJSON))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	var user models.User
	json.Unmarshal(resp.Body.Bytes(), &user)
	return user
}

func createAvailability(router *gin.Engine, userID, eventID uint, start, end time.Time) {
	availJSON, _ := json.Marshal(map[string]interface{}{
		"start_time": start.Format(time.RFC3339),
		"end_time":   end.Format(time.RFC3339),
	})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/users/%d/events/%d/availability", userID, eventID), bytes.NewBuffer(availJSON))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), req)
}
