package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/krushnna/meeting-scheduler/models"
	"github.com/krushnna/meeting-scheduler/services"
	"go.uber.org/zap"
)

// EventController handles HTTP requests for events
type EventController struct {
	service *services.EventService
	logger  *zap.Logger
}

func NewEventController(service *services.EventService, logger *zap.Logger) *EventController {
	return &EventController{
		service: service,
		logger:  logger.With(zap.String("controller", "event")),
	}
}

// CreateEvent validates input and creates a new event. Returns detailed error messages.
func (c *EventController) CreateEvent(ctx *gin.Context) {
	var event models.Event
	if err := ctx.ShouldBindJSON(&event); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	// Validate required fields
	if event.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Event title is required"})
		return
	}
	if event.DurationMinutes <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Event duration must be greater than zero"})
		return
	}

	c.logger.Info("Crreating new event", zap.String("title", event.Title))
	if err := c.service.CreateEvent(&event); err != nil {
		c.logger.Error("Failed to create event", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating event: " + err.Error()})
		return
	}

	c.logger.Info("Event created successfully", zap.Uint("event_id", event.ID))
	ctx.JSON(http.StatusCreated, event)
}

// GetEvent retrieves an event by its ID.
func (c *EventController) GetEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	c.logger.Debug("Fetching event", zap.Uint64("id", id))
	event, err := c.service.GetEvent(uint(id))
	if err != nil {
		c.logger.Error("Event not found", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	ctx.JSON(http.StatusOK, event)
}

// GetAllEvents returns all events with pagination support.
// Query parameters: limit (default 10) and offset (default 0)
func (c *EventController) GetAllEvents(ctx *gin.Context) {
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")
	var limit, offset int
	var err error

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
			return
		}
	} else {
		limit = 10 // default limit
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset value"})
			return
		}
	} else {
		offset = 0 // default offset
	}

	c.logger.Debug("Fetching events with pagination", zap.Int("limit", limit), zap.Int("offset", offset))
	// Call a service method that supports pagination.
	events, err := c.service.GetAllEventsWithPagination(limit, offset)
	if err != nil {
		c.logger.Error("Failed to fetch events", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching events: " + err.Error()})
		return
	}

	c.logger.Info("Retrieved events", zap.Int("count", len(events)))
	ctx.JSON(http.StatusOK, events)
}

// UpdateEvent modifies an existing event.
func (c *EventController) UpdateEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	var event models.Event
	if err := ctx.ShouldBindJSON(&event); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	// Validate fields again if needed.
	if event.Title == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Event title is required"})
		return
	}
	if event.DurationMinutes <= 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Event duration must be greater than zero"})
		return
	}

	c.logger.Info("Updating event", zap.Uint64("id", id))
	if err := c.service.UpdateEvent(uint(id), &event); err != nil {
		c.logger.Error("Failed to update event", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating event: " + err.Error()})
		return
	}

	c.logger.Info("Event updated successfully", zap.Uint64("id", id))
	ctx.JSON(http.StatusOK, gin.H{"message": "Event updated successfully"})
}

// DeleteEvent removes an event.
func (c *EventController) DeleteEvent(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	c.logger.Info("Deleting event", zap.Uint64("id", id))
	if err := c.service.DeleteEvent(uint(id)); err != nil {
		c.logger.Error("Failed to delete event", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting event: " + err.Error()})
		return
	}

	c.logger.Info("Event deleted successfully", zap.Uint64("id", id))
	ctx.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

// TimeSlotController handles HTTP requests for time slots.
type TimeSlotController struct {
	service *services.TimeSlotService
	logger  *zap.Logger
}

func NewTimeSlotController(service *services.TimeSlotService, logger *zap.Logger) *TimeSlotController {
	return &TimeSlotController{
		service: service,
		logger:  logger.With(zap.String("controller", "timeslot")),
	}
}

// CreateTimeSlot creates a new timeslot associated with an event.
func (c *TimeSlotController) CreateTimeSlot(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid event ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	var timeSlot models.TimeSlot
	if err := ctx.ShouldBindJSON(&timeSlot); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	timeSlot.EventID = uint(eventID)
	c.logger.Info("Creating time slot", zap.Uint64("event_id", eventID))
	if err := c.service.CreateTimeSlot(&timeSlot); err != nil {
		c.logger.Error("Failed to create time slot", zap.Uint64("event_id", eventID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating time slot: " + err.Error()})
		return
	}

	c.logger.Info("Time slot created successfully", zap.Uint("slot_id", timeSlot.ID), zap.Uint64("event_id", eventID))
	ctx.JSON(http.StatusCreated, timeSlot)
}

// GetTimeSlotsByEvent retrieves all timeslots for a given event.
func (c *TimeSlotController) GetTimeSlotsByEvent(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid event ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	c.logger.Debug("Fetching time slots for event", zap.Uint64("event_id", eventID))
	timeSlots, err := c.service.GetTimeSlotsByEvent(uint(eventID))
	if err != nil {
		c.logger.Error("Failed to fetch time slots", zap.Uint64("event_id", eventID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching time slots: " + err.Error()})
		return
	}

	c.logger.Info("Retrieved time slots", zap.Uint64("event_id", eventID), zap.Int("count", len(timeSlots)))
	ctx.JSON(http.StatusOK, timeSlots)
}

// UpdateTimeSlot updates an existing timeslot.
func (c *TimeSlotController) UpdateTimeSlot(ctx *gin.Context) {
	slotID, err := strconv.ParseUint(ctx.Param("slotId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid time slot ID format", zap.String("slot_id", ctx.Param("slotId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time slot ID format"})
		return
	}

	var timeSlot models.TimeSlot
	if err := ctx.ShouldBindJSON(&timeSlot); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	c.logger.Info("Updating time slot", zap.Uint64("slot_id", slotID))
	if err := c.service.UpdateTimeSlot(uint(slotID), &timeSlot); err != nil {
		c.logger.Error("Failed to update time slot", zap.Uint64("slot_id", slotID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating time slot: " + err.Error()})
		return
	}

	c.logger.Info("Time slot updated successfully", zap.Uint64("slot_id", slotID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Time slot updated successfully"})
}

// DeleteTimeSlot deletes a timeslot.
func (c *TimeSlotController) DeleteTimeSlot(ctx *gin.Context) {
	slotID, err := strconv.ParseUint(ctx.Param("slotId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid time slot ID format", zap.String("slot_id", ctx.Param("slotId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time slot ID format"})
		return
	}

	c.logger.Info("Deleting time slot", zap.Uint64("slot_id", slotID))
	if err := c.service.DeleteTimeSlot(uint(slotID)); err != nil {
		c.logger.Error("Failed to delete time slot", zap.Uint64("slot_id", slotID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting time slot: " + err.Error()})
		return
	}

	c.logger.Info("Time slot deleted successfully", zap.Uint64("slot_id", slotID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Time slot deleted successfully"})
}

// UserController handles HTTP requests for usersa
type UserController struct {
	service *services.UserService
	logger  *zap.Logger
}

func NewUserController(service *services.UserService, logger *zap.Logger) *UserController {
	return &UserController{
		service: service,
		logger:  logger.With(zap.String("controller", "user")),
	}
}

// CreateUser creates a new user.
func (c *UserController) CreateUser(ctx *gin.Context) {
	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	c.logger.Info("Creating user", zap.String("email", user.Email))
	if err := c.service.CreateUser(&user); err != nil {
		c.logger.Error("Failed to create user", zap.String("email", user.Email), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user: " + err.Error()})
		return
	}

	c.logger.Info("User created successfully", zap.Uint("user_id", user.ID), zap.String("email", user.Email))
	ctx.JSON(http.StatusCreated, user)
}

// GetUser retrieves a user by ID.
func (c *UserController) GetUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	c.logger.Debug("Fetching user", zap.Uint64("id", id))
	user, err := c.service.GetUser(uint(id))
	if err != nil {
		c.logger.Error("User not found", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}

// GetAllUsers retrieves all users.
func (c *UserController) GetAllUsers(ctx *gin.Context) {
	c.logger.Debug("Fetching all users")
	users, err := c.service.GetAllUsers()
	if err != nil {
		c.logger.Error("Failed to fetch users", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching users: " + err.Error()})
		return
	}

	c.logger.Info("Retrieved all users", zap.Int("count", len(users)))
	ctx.JSON(http.StatusOK, users)
}

// UpdateUser updates an existing user.
func (c *UserController) UpdateUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user models.User
	if err := ctx.ShouldBindJSON(&user); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	c.logger.Info("Updating user", zap.Uint64("id", id))
	if err := c.service.UpdateUser(uint(id), &user); err != nil {
		c.logger.Error("Failed to update user", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user: " + err.Error()})
		return
	}

	c.logger.Info("User updated successfully", zap.Uint64("id", id))
	ctx.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
}

// DeleteUser deletes a user.
func (c *UserController) DeleteUser(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid ID format", zap.String("id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	c.logger.Info("Deleting user", zap.Uint64("id", id))
	if err := c.service.DeleteUser(uint(id)); err != nil {
		c.logger.Error("Failed to delete user", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting user: " + err.Error()})
		return
	}

	c.logger.Info("User deleted successfully", zap.Uint64("id", id))
	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// AvailabilityController handles HTTP requests for user availability.
type AvailabilityController struct {
	service *services.AvailabilityService
	logger  *zap.Logger
}

func NewAvailabilityController(service *services.AvailabilityService, logger *zap.Logger) *AvailabilityController {
	return &AvailabilityController{
		service: service,
		logger:  logger.With(zap.String("controller", "availability")),
	}
}

// CreateAvailability creates a new availability record.
func (c *AvailabilityController) CreateAvailability(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid user ID format", zap.String("user_id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	eventID, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid event ID format", zap.String("event_id", ctx.Param("eventId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	var availability models.UserAvailability
	if err := ctx.ShouldBindJSON(&availability); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	availability.UserID = uint(userID)
	availability.EventID = uint(eventID)

	c.logger.Info("Creating availability", zap.Uint64("user_id", userID), zap.Uint64("event_id", eventID))
	if err := c.service.CreateAvailability(&availability); err != nil {
		c.logger.Error("Failed to create availability", zap.Uint64("user_id", userID), zap.Uint64("event_id", eventID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating availability: " + err.Error()})
		return
	}

	c.logger.Info("Availability created successfully", zap.Uint("avail_id", availability.ID))
	ctx.JSON(http.StatusCreated, availability)
}

// GetUserAvailability retrieves availability records for a user in an event.
func (c *AvailabilityController) GetUserAvailability(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid user ID format", zap.String("user_id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	eventID, err := strconv.ParseUint(ctx.Param("eventId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid event ID format", zap.String("event_id", ctx.Param("eventId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	c.logger.Debug("Fetching user availability", zap.Uint64("user_id", userID), zap.Uint64("event_id", eventID))
	availabilities, err := c.service.GetUserAvailability(uint(userID), uint(eventID))
	if err != nil {
		c.logger.Error("Failed to fetch availability", zap.Uint64("user_id", userID), zap.Uint64("event_id", eventID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching availability: " + err.Error()})
		return
	}

	c.logger.Info("Retrieved user availability", zap.Uint64("user_id", userID), zap.Uint64("event_id", eventID), zap.Int("count", len(availabilities)))
	ctx.JSON(http.StatusOK, availabilities)
}

// UpdateAvailability updates an existing availability record.
func (c *AvailabilityController) UpdateAvailability(ctx *gin.Context) {
	availID, err := strconv.ParseUint(ctx.Param("availId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid availability ID format", zap.String("avail_id", ctx.Param("availId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid availability ID format"})
		return
	}

	var availability models.UserAvailability
	if err := ctx.ShouldBindJSON(&availability); err != nil {
		c.logger.Error("Failed to bind JSON", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload: " + err.Error()})
		return
	}

	c.logger.Info("Updating availability", zap.Uint64("avail_id", availID))
	if err := c.service.UpdateAvailability(uint(availID), &availability); err != nil {
		c.logger.Error("Failed to update availability", zap.Uint64("avail_id", availID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating availability: " + err.Error()})
		return
	}

	c.logger.Info("Availability updated successfully", zap.Uint64("avail_id", availID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Availability updated successfully"})
}

// DeleteAvailability deletes an availability record.
func (c *AvailabilityController) DeleteAvailability(ctx *gin.Context) {
	availID, err := strconv.ParseUint(ctx.Param("availId"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid availability ID format", zap.String("avail_id", ctx.Param("availId")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid availability ID format"})
		return
	}

	c.logger.Info("Deleting availability", zap.Uint64("avail_id", availID))
	if err := c.service.DeleteAvailability(uint(availID)); err != nil {
		c.logger.Error("Failed to delete availability", zap.Uint64("avail_id", availID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting availability: " + err.Error()})
		return
	}

	c.logger.Info("Availability deleted successfully", zap.Uint64("avail_id", availID))
	ctx.JSON(http.StatusOK, gin.H{"message": "Availability deleted successfully"})
}

// RecommendationController handles HTTP requests for time slot recommendations.
type RecommendationController struct {
	service *services.RecommendationService
	logger  *zap.Logger
}

func NewRecommendationController(service *services.RecommendationService, logger *zap.Logger) *RecommendationController {
	return &RecommendationController{
		service: service,
		logger:  logger.With(zap.String("controller", "recommendation")),
	}
}

// GetRecommendations generates and returns time slot recommendations.
// It relies on proper JSON struct tags (with omitempty) in the models to omit null values.
func (c *RecommendationController) GetRecommendations(ctx *gin.Context) {
	eventID, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		c.logger.Error("Invalid event ID format", zap.String("event_id", ctx.Param("id")), zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID format"})
		return
	}

	c.logger.Info("Generating recommendations", zap.Uint64("event_id", eventID))
	recommendations, err := c.service.GetRecommendations(uint(eventID))
	if err != nil {
		c.logger.Error("Failed to generate recommendations", zap.Uint64("event_id", eventID), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error generating recommendations: " + err.Error()})
		return
	}

	c.logger.Info("Recommendations generated successfully", zap.Uint64("event_id", eventID), zap.Int("count", len(recommendations)))
	ctx.JSON(http.StatusOK, recommendations)
}
