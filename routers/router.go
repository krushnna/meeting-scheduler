package routers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/krushnna/meeting-scheduler/controllers"
	"github.com/krushnna/meeting-scheduler/repository"
	"github.com/krushnna/meeting-scheduler/services"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// SetupRouter initializes the Gin router, middleware, and routes.
func SetupRouter(db *gorm.DB, logger *zap.Logger) *gin.Engine {
	// Initialize repositories
	eventRepo := repository.NewEventRepository(db)
	timeSlotRepo := repository.NewTimeSlotRepository(db)
	userRepo := repository.NewUserRepository(db)
	userAvailabilityRepo := repository.NewUserAvailabilityRepository(db)

	// Initialize services
	eventService := services.NewEventService(eventRepo)
	timeSlotService := services.NewTimeSlotService(timeSlotRepo)
	userService := services.NewUserService(userRepo)
	availabilityService := services.NewAvailabilityService(userAvailabilityRepo)
	recommendationService := services.NewRecommendationService(eventRepo, timeSlotRepo, userAvailabilityRepo)

	// Initialize controllers
	eventController := controllers.NewEventController(eventService, logger)
	timeSlotController := controllers.NewTimeSlotController(timeSlotService, logger)
	userController := controllers.NewUserController(userService, logger)
	availabilityController := controllers.NewAvailabilityController(availabilityService, logger)
	recommendationController := controllers.NewRecommendationController(recommendationService, logger)

	// Create router and apply middleware
	router := gin.Default()


	// Serve docs folder for static files (if needed)
	router.Static("/docs", "./docs")

	// Swagger route: if you are using swag generated docs, uncomment and update the URL option.
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/openapi.yaml")))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "time": time.Now().Format(time.RFC3339)})
	})

	// API routes grouped under /api/v1
	api := router.Group("/api/v1")
	{
		// Events endpoints
		events := api.Group("/events")
		{
			events.POST("", eventController.CreateEvent)
			events.GET("", eventController.GetAllEvents)
			events.GET("/:id", eventController.GetEvent)
			events.PUT("/:id", eventController.UpdateEvent)
			events.DELETE("/:id", eventController.DeleteEvent)
			events.GET("/:id/recommendations", recommendationController.GetRecommendations)

			// TimeSlots endpoints for an event
			timeslots := events.Group("/:id/timeslots")
			{
				timeslots.POST("", timeSlotController.CreateTimeSlot)
				timeslots.GET("", timeSlotController.GetTimeSlotsByEvent)
				timeslots.PUT("/:slotId", timeSlotController.UpdateTimeSlot)
				timeslots.DELETE("/:slotId", timeSlotController.DeleteTimeSlot)
			}
		}

		// Users endpoints
		users := api.Group("/users")
		{
			users.POST("", userController.CreateUser)
			users.GET("", userController.GetAllUsers)
			users.GET("/:id", userController.GetUser)
			users.PUT("/:id", userController.UpdateUser)
			users.DELETE("/:id", userController.DeleteUser)
			users.POST("/:id/events/:eventId/availability", availabilityController.CreateAvailability)
			users.GET("/:id/events/:eventId/availability", availabilityController.GetUserAvailability)
		}
	}

	return router
}
