package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"focusflow-be/internal/models"
	"focusflow-be/internal/services"
)

type DashboardHandler struct {
	firebaseService *services.FirebaseService
	authService     *services.AuthService
}

func NewDashboardHandler(firebaseService *services.FirebaseService, authService *services.AuthService) *DashboardHandler {
	return &DashboardHandler{
		firebaseService: firebaseService,
		authService:     authService,
	}
}

func (h *DashboardHandler) GetCalendarEvents(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	var events []models.CalendarEvent

	// Get tasks
	tasks, err := h.firebaseService.GetTasks(userSession.UserID)
	if err == nil {
		for _, task := range tasks {
			if task.DueDate != nil {
				startTime := time.Now()
				if task.StartDate != nil {
					startTime = *task.StartDate
				}

				color := "#10b981" // green for low
				if task.Priority == "medium" {
					color = "#f59e0b" // yellow
				} else if task.Priority == "high" {
					color = "#ef4444" // red
				}

				events = append(events, models.CalendarEvent{
					ID:          task.ID,
					Title:       task.Title,
					Start:       startTime.Format(time.RFC3339),
					End:         task.DueDate.Format(time.RFC3339),
					Type:        "task",
					Status:      task.Status,
					Color:       &color,
					Description: task.Description,
				})
			}
		}
	}

	// Get meetings
	meetings, err := h.firebaseService.GetMeetings(userSession.UserID)
	if err == nil {
		for _, meeting := range meetings {
			color := "#3b82f6" // blue
			events = append(events, models.CalendarEvent{
				ID:          meeting.ID,
				Title:       meeting.Title,
				Start:       meeting.StartTime.Format(time.RFC3339),
				End:         meeting.EndTime.Format(time.RFC3339),
				Type:        "meeting",
				Status:      meeting.Status,
				Color:       &color,
				Description: meeting.Description,
			})
		}
	}

	// Get reminders
	reminders, err := h.firebaseService.GetReminders(userSession.UserID)
	if err == nil {
		for _, reminder := range reminders {
			color := "#8b5cf6" // purple
			status := "pending"
			if reminder.IsCompleted {
				status = "completed"
			}

			events = append(events, models.CalendarEvent{
				ID:          reminder.ID,
				Title:       reminder.Title,
				Start:       reminder.ReminderTime.Format(time.RFC3339),
				End:         reminder.ReminderTime.Format(time.RFC3339),
				Type:        "reminder",
				Status:      status,
				Color:       &color,
				Description: reminder.Description,
			})
		}
	}

	c.JSON(http.StatusOK, events)
}

func (h *DashboardHandler) GetGanttData(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	var ganttItems []models.GanttItem

	// Get tasks with start and end dates
	tasks, err := h.firebaseService.GetTasks(userSession.UserID)
	if err == nil {
		for _, task := range tasks {
			if task.StartDate != nil && task.DueDate != nil {
				progress := 0
				if task.Status == "completed" {
					progress = 100
				} else if task.Status == "in-progress" {
					progress = 50
				}

				ganttItems = append(ganttItems, models.GanttItem{
					ID:       task.ID,
					Title:    task.Title,
					Start:    task.StartDate.Format(time.RFC3339),
					End:      task.DueDate.Format(time.RFC3339),
					Progress: progress,
					Type:     "task",
					Status:   task.Status,
					Priority: task.Priority,
				})
			}
		}
	}

	// Get meetings
	meetings, err := h.firebaseService.GetMeetings(userSession.UserID)
	if err == nil {
		for _, meeting := range meetings {
			progress := 0
			if meeting.Status == "completed" {
				progress = 100
			} else if meeting.Status == "ongoing" {
				progress = 50
			}

			ganttItems = append(ganttItems, models.GanttItem{
				ID:       meeting.ID,
				Title:    meeting.Title,
				Start:    meeting.StartTime.Format(time.RFC3339),
				End:      meeting.EndTime.Format(time.RFC3339),
				Progress: progress,
				Type:     "meeting",
				Status:   meeting.Status,
				Priority: "medium",
			})
		}
	}

	c.JSON(http.StatusOK, ganttItems)
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)

	overview := models.Overview{}
	today := time.Now().Format("2006-01-02")

	// Get task statistics
	tasks, err := h.firebaseService.GetTasks(userSession.UserID)
	if err == nil {
		overview.Tasks.Total = len(tasks)
		for _, task := range tasks {
			switch task.Status {
			case "completed":
				overview.Tasks.Completed++
			case "in-progress":
				overview.Tasks.InProgress++
			case "todo":
				overview.Tasks.Todo++
			}

			if task.Priority == "high" {
				overview.Tasks.HighPriority++
			}

			// Check if overdue
			if task.DueDate != nil && task.Status != "completed" {
				if task.DueDate.Format("2006-01-02") < today {
					overview.Tasks.Overdue++
				}
			}
		}
	}

	// Get meeting statistics
	meetings, err := h.firebaseService.GetMeetings(userSession.UserID)
	if err == nil {
		overview.Meetings.Total = len(meetings)
		for _, meeting := range meetings {
			if meeting.StartTime.Format("2006-01-02") == today {
				overview.Meetings.Today++
			}

			switch meeting.Status {
			case "scheduled":
				overview.Meetings.Upcoming++
			case "completed":
				overview.Meetings.Completed++
			}
		}
	}

	// Get reminder statistics
	reminders, err := h.firebaseService.GetReminders(userSession.UserID)
	if err == nil {
		overview.Reminders.Total = len(reminders)
		now := time.Now()
		for _, reminder := range reminders {
			if reminder.IsCompleted {
				overview.Reminders.Completed++
			} else {
				overview.Reminders.Pending++
				if reminder.ReminderTime.Before(now) {
					overview.Reminders.Overdue++
				}
			}
		}
	}

	c.JSON(http.StatusOK, overview)
}