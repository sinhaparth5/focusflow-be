package main

import (
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"focusflow-be/internal/config"
	"focusflow-be/internal/handlers"
	"focusflow-be/internal/middleware"
	"focusflow-be/internal/services"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration from environment variables
	cfg := config.New()

	// Validate required configuration
	if cfg.FirebaseProjectID == "" || cfg.GoogleClientID == "" || cfg.JWTSecret == "" {
		log.Fatal("Missing required environment variables. Please check FIREBASE_PROJECT_ID, GOOGLE_CLIENT_ID, and JWT_SECRET")
	}

	// Initialize Firebase service
	firebaseService, err := services.NewFirebaseService(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase service: %v", err)
	}
	defer firebaseService.Close()

	// Initialize other services
	googleService := services.NewGoogleService(cfg)
	authService := services.NewAuthService(cfg)

	// Initialize all handlers with their dependencies
	authHandler := handlers.NewAuthHandler(authService, googleService, firebaseService)
	taskHandler := handlers.NewTaskHandler(firebaseService, authService)
	meetingHandler := handlers.NewMeetingHandler(firebaseService, authService)
	reminderHandler := handlers.NewReminderHandler(firebaseService, authService)
	dashboardHandler := handlers.NewDashboardHandler(firebaseService, authService)

	// Setup Gin router with middleware
	r := gin.Default()

	// Configure CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Root health check endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "FocusFlow Task Management API",
			"version": "1.0.0",
			"status":  "online",
			"docs":    "https://github.com/sinhaparth5/focusflow-be",
			"endpoints": gin.H{
				"authentication": gin.H{
					"google_auth": "GET /auth/google",
					"callback":    "GET /auth/callback",
					"me":          "GET /auth/me",
					"debug":       "GET /auth/debug",
				},
				"tasks": gin.H{
					"list":     "GET /tasks",
					"create":   "POST /tasks",
					"update":   "PUT /tasks/:id",
					"delete":   "DELETE /tasks/:id",
					"start":    "PATCH /tasks/:id/start",
					"complete": "PATCH /tasks/:id/complete",
				},
				"meetings": gin.H{
					"list":         "GET /meetings",
					"create":       "POST /meetings",
					"updateStatus": "PATCH /meetings/:id/status",
				},
				"reminders": gin.H{
					"list":     "GET /reminders",
					"create":   "POST /reminders",
					"complete": "PATCH /reminders/:id/complete",
				},
				"dashboard": gin.H{
					"calendar": "GET /dashboard/calendar",
					"gantt":    "GET /dashboard/gantt",
					"overview": "GET /dashboard/overview",
				},
			},
		})
	})

	// Authentication routes (public)
	authGroup := r.Group("/auth")
	{
		authGroup.GET("/google", authHandler.GoogleAuth)
		authGroup.GET("/callback", authHandler.GoogleCallback)
		authGroup.GET("/debug", authHandler.Debug)
		
		// Protected auth routes
		authGroup.GET("/me", middleware.AuthMiddleware(authService), authHandler.GetMe)
	}

	// Protected API routes (require authentication)
	api := r.Group("/")
	api.Use(middleware.AuthMiddleware(authService))
	{
		// Task management endpoints
		taskGroup := api.Group("/tasks")
		{
			taskGroup.GET("/", taskHandler.GetTasks)
			taskGroup.POST("/", taskHandler.CreateTask)
			taskGroup.PUT("/:id", taskHandler.UpdateTask)
			taskGroup.DELETE("/:id", taskHandler.DeleteTask)
			taskGroup.PATCH("/:id/start", taskHandler.StartTask)
			taskGroup.PATCH("/:id/complete", taskHandler.CompleteTask)
		}

		// Meeting management endpoints
		meetingGroup := api.Group("/meetings")
		{
			meetingGroup.GET("/", meetingHandler.GetMeetings)
			meetingGroup.POST("/", meetingHandler.CreateMeeting)
			meetingGroup.PATCH("/:id/status", meetingHandler.UpdateMeetingStatus)
		}

		// Reminder management endpoints
		reminderGroup := api.Group("/reminders")
		{
			reminderGroup.GET("/", reminderHandler.GetReminders)
			reminderGroup.POST("/", reminderHandler.CreateReminder)
			reminderGroup.PATCH("/:id/complete", reminderHandler.CompleteReminder)
		}

		// Dashboard analytics endpoints
		dashboardGroup := api.Group("/dashboard")
		{
			dashboardGroup.GET("/calendar", dashboardHandler.GetCalendarEvents)
			dashboardGroup.GET("/gantt", dashboardHandler.GetGanttData)
			dashboardGroup.GET("/overview", dashboardHandler.GetOverview)
		}
	}

	// Get port from environment or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start the server
	log.Printf("🚀 FocusFlow API server starting on port %s", port)
	log.Printf("📍 Health check: http://localhost:%s/", port)
	log.Printf("🔐 Authentication: http://localhost:%s/auth/google", port)
	log.Printf("📚 API Documentation: Check README.md for endpoints")
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}