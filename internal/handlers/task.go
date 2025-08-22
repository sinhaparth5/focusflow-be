package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"focusflow-be/internal/models"
	"focusflow-be/internal/services"
)

type TaskHandler struct {
	firebaseService *services.FirebaseService
	authService     *services.AuthService
}

func NewTaskHandler(firebaseService *services.FirebaseService, authService *services.AuthService) *TaskHandler {
	return &TaskHandler{
		firebaseService: firebaseService,
		authService:     authService,
	}
}

func (h *TaskHandler) GetTasks(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)
	tasks, err := h.firebaseService.GetTasks(userSession.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (h *TaskHandler) CreateTask(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	var req models.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	task := &models.Task{
		UserID:         userSession.UserID,
		Title:          req.Title,
		Description:    req.Description,
		Completed:      false,
		Status:         "todo",
		Priority:       req.Priority,
		StartDate:      req.StartDate,
		DueDate:        req.DueDate,
		EstimatedHours: req.EstimatedHours,
	}

	taskID, err := h.firebaseService.CreateTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":      taskID,
		"message": "Task created successfully",
	})
}

func (h *TaskHandler) UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	var req models.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Status != nil {
		updates["status"] = *req.Status
		if *req.Status == "completed" {
			updates["completed"] = true
		}
	}
	if req.StartDate != nil {
		updates["startDate"] = *req.StartDate
	}
	if req.DueDate != nil {
		updates["dueDate"] = *req.DueDate
	}
	if req.EstimatedHours != nil {
		updates["estimatedHours"] = *req.EstimatedHours
	}
	if req.ActualHours != nil {
		updates["actualHours"] = *req.ActualHours
	}

	if err := h.firebaseService.UpdateTask(taskID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task updated successfully"})
}

func (h *TaskHandler) DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	if err := h.firebaseService.DeleteTask(taskID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted successfully"})
}

func (h *TaskHandler) StartTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	updates := map[string]interface{}{
		"status":    "in-progress",
		"startedAt": time.Now(),
	}

	if err := h.firebaseService.UpdateTask(taskID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task started successfully"})
}

func (h *TaskHandler) CompleteTask(c *gin.Context) {
	taskID := c.Param("id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	updates := map[string]interface{}{
		"status":      "completed",
		"completed":   true,
		"completedAt": time.Now(),
	}

	if err := h.firebaseService.UpdateTask(taskID, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete task", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Task completed successfully"})
}