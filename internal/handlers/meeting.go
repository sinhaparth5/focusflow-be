package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"focusflow-be/internal/models"
	"focusflow-be/internal/services"
)

type MeetingHandler struct {
	firebaseService *services.FirebaseService
	authService     *services.AuthService
}

func NewMeetingHandler(firebaseService *services.FirebaseService, authService *services.AuthService) *MeetingHandler {
	return &MeetingHandler{
		firebaseService: firebaseService,
		authService:     authService,
	}
}

func (h *MeetingHandler) GetMeetings(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)
	meetings, err := h.firebaseService.GetMeetings(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch meetings", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, meetings)
}

func (h *MeetingHandler) CreateMeeting(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	var req models.CreateMeetingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validate that end time is after start time
	if req.EndTime.Before(req.StartTime) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End time must be after start time"})
		return
	}

	meeting := &models.Meeting{
		UserID:      userSession.UserID,
		Title:       req.Title,
		Description: req.Description,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Attendees:   req.Attendees,
		Location:    req.Location,
		MeetingType: req.MeetingType,
		Status:      "scheduled",
	}

	meetingID, err := h.firebaseService.CreateMeeting(meeting)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create meeting", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      meetingID,
		"message": "Meeting created successfully",
	})
}

func (h *MeetingHandler) UpdateMeetingStatus(c *gin.Context) {
	meetingID := c.Param("id")
	if meetingID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Meeting ID is required"})
		return
	}

	var req models.UpdateMeetingStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	updates := map[string]interface{}{
		"status": req.Status,
	}

	if err := h.firebaseService.UpdateMeeting(meetingID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update meeting status", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Meeting status updated successfully"})
}