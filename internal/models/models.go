package models

import (
	"time"
)

type UserSession struct {
	UserID       string  `json:"userId" firestore:"userId"`
	Email        string  `json:"email" firestore:"email"`
	Name         string  `json:"name" firestore:"name"`
	AccessToken  string  `json:"accessToken" firestore:"accessToken"`
	RefreshToken *string `json:"refreshToken,omitempty" firestore:"refreshToken,omitempty"`
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
	LastLogin    time.Time `json:"lastLogin" firestore:"lastLogin"`
}

type Task struct {
	ID             string     `json:"id,omitempty" firestore:"-"`
	UserID         string     `json:"userId" firestore:"userId"`
	Title          string     `json:"title" firestore:"title"`
	Description    *string    `json:"description,omitempty" firestore:"description,omitempty"`
	Completed      bool       `json:"completed" firestore:"completed"`
	Status         string     `json:"status" firestore:"status"` // todo, in-progress, completed
	Priority       string     `json:"priority" firestore:"priority"` // low, medium, high
	StartDate      *time.Time `json:"startDate,omitempty" firestore:"startDate,omitempty"`
	DueDate        *time.Time `json:"dueDate,omitempty" firestore:"dueDate,omitempty"`
	EstimatedHours *int       `json:"estimatedHours,omitempty" firestore:"estimatedHours,omitempty"`
	ActualHours    *int       `json:"actualHours,omitempty" firestore:"actualHours,omitempty"`
	GoogleEventID  *string    `json:"googleEventId,omitempty" firestore:"googleEventId,omitempty"`
	CreatedAt      time.Time  `json:"createdAt" firestore:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt" firestore:"updatedAt"`
}

type Meeting struct {
	ID            string     `json:"id,omitempty" firestore:"-"`
	UserID        string     `json:"userId" firestore:"userId"`
	Title         string     `json:"title" firestore:"title"`
	Description   *string    `json:"description,omitempty" firestore:"description,omitempty"`
	StartTime     time.Time  `json:"startTime" firestore:"startTime"`
	EndTime       time.Time  `json:"endTime" firestore:"endTime"`
	Attendees     []string   `json:"attendees,omitempty" firestore:"attendees,omitempty"`
	Location      *string    `json:"location,omitempty" firestore:"location,omitempty"`
	MeetingType   string     `json:"meetingType" firestore:"meetingType"` // call, in-person, video
	Status        string     `json:"status" firestore:"status"` // scheduled, ongoing, completed, cancelled
	GoogleEventID *string    `json:"googleEventId,omitempty" firestore:"googleEventId,omitempty"`
	CreatedAt     time.Time  `json:"createdAt" firestore:"createdAt"`
}

type Reminder struct {
	ID           string    `json:"id,omitempty" firestore:"-"`
	UserID       string    `json:"userId" firestore:"userId"`
	Title        string    `json:"title" firestore:"title"`
	Description  *string   `json:"description,omitempty" firestore:"description,omitempty"`
	ReminderTime time.Time `json:"reminderTime" firestore:"reminderTime"`
	ReminderType string    `json:"reminderType" firestore:"reminderType"` // task, meeting, personal
	IsCompleted  bool      `json:"isCompleted" firestore:"isCompleted"`
	Priority     string    `json:"priority" firestore:"priority"` // low, medium, high
	GoogleEventID *string  `json:"googleEventId,omitempty" firestore:"googleEventId,omitempty"`
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
}

type CalendarEvent struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Start       string  `json:"start"`
	End         string  `json:"end"`
	Type        string  `json:"type"` // task, meeting, reminder
	Status      string  `json:"status"`
	Color       *string `json:"color,omitempty"`
	Description *string `json:"description,omitempty"`
}

type GanttItem struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Start        string   `json:"start"`
	End          string   `json:"end"`
	Progress     int      `json:"progress"`
	Type         string   `json:"type"` // task, meeting
	Status       string   `json:"status"`
	Dependencies []string `json:"dependencies,omitempty"`
	Priority     string   `json:"priority"`
}

type Overview struct {
	Tasks     TaskOverview     `json:"tasks"`
	Meetings  MeetingOverview  `json:"meetings"`
	Reminders ReminderOverview `json:"reminders"`
}

type TaskOverview struct {
	Total        int `json:"total"`
	Completed    int `json:"completed"`
	InProgress   int `json:"inProgress"`
	Todo         int `json:"todo"`
	HighPriority int `json:"highPriority"`
	Overdue      int `json:"overdue"`
}

type MeetingOverview struct {
	Total     int `json:"total"`
	Today     int `json:"today"`
	Upcoming  int `json:"upcoming"`
	Completed int `json:"completed"`
}

type ReminderOverview struct {
	Total     int `json:"total"`
	Pending   int `json:"pending"`
	Completed int `json:"completed"`
	Overdue   int `json:"overdue"`
}

type GoogleUserInfo struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Request/Response DTOs
type CreateTaskRequest struct {
	Title          string     `json:"title" binding:"required"`
	Description    *string    `json:"description"`
	Priority       string     `json:"priority" binding:"required,oneof=low medium high"`
	StartDate      *time.Time `json:"startDate"`
	DueDate        *time.Time `json:"dueDate"`
	EstimatedHours *int       `json:"estimatedHours"`
}

type UpdateTaskRequest struct {
	Title          *string    `json:"title"`
	Description    *string    `json:"description"`
	Priority       *string    `json:"priority" binding:"omitempty,oneof=low medium high"`
	Status         *string    `json:"status" binding:"omitempty,oneof=todo in-progress completed"`
	StartDate      *time.Time `json:"startDate"`
	DueDate        *time.Time `json:"dueDate"`
	EstimatedHours *int       `json:"estimatedHours"`
	ActualHours    *int       `json:"actualHours"`
}

type CreateMeetingRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	StartTime   time.Time  `json:"startTime" binding:"required"`
	EndTime     time.Time  `json:"endTime" binding:"required"`
	Attendees   []string   `json:"attendees"`
	Location    *string    `json:"location"`
	MeetingType string     `json:"meetingType" binding:"required,oneof=call in-person video"`
}

type CreateReminderRequest struct {
	Title        string    `json:"title" binding:"required"`
	Description  *string   `json:"description"`
	ReminderTime time.Time `json:"reminderTime" binding:"required"`
	ReminderType string    `json:"reminderType" binding:"required,oneof=task meeting personal"`
	Priority     string    `json:"priority" binding:"required,oneof=low medium high"`
}

type UpdateMeetingStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=scheduled ongoing completed cancelled"`
}