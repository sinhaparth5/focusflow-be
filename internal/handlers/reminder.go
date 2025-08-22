package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"focusflow-be/internal/models"
	"focusflow-be/internal/services"
)

type ReminderHandler struct {
	firebaseService *services.FirebaseService
	authService     *services.AuthService
}

func NewReminderHandler(firebaseService *services.FirebaseService, authService *services.AuthService) *ReminderHandler {
	return &ReminderHandler{
		firebaseService: firebaseService,
		authService:     authService,
	}
}

func (h *ReminderHandler) GetReminders(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)
	reminders, err := h.firebaseService.GetReminders(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reminders", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, reminders)
}

func (h *ReminderHandler) CreateReminder(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	var req models.CreateReminderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	reminder := &models.Reminder{
		UserID:       userSession.UserID,
		Title:        req.Title,
		Description:  req.Description,
		ReminderTime: req.ReminderTime,
		ReminderType: req.ReminderType,
		IsCompleted:  false,
		Priority:     req.Priority,
	}

	reminderID, err := h.firebaseService.CreateReminder(reminder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reminder", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      reminderID,
		"message": "Reminder created successfully",
	})
}

func (h *ReminderHandler) CompleteReminder(c *gin.Context) {
	reminderID := c.Param("id")
	if reminderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Reminder ID is required"})
		return
	}

	updates := map[string]interface{}{
		"isCompleted": true,
		"completedAt": time.Now(),
	}

	if err := h.firebaseService.UpdateReminder(reminderID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete reminder", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reminder marked as completed"})
}