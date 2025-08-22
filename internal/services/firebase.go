package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"focusflow-be/internal/config"
	"focusflow-be/internal/models"
)

type FirebaseService struct {
	projectID string
	apiKey    string
	baseURL   string
	client    *http.Client
}

func NewFirebaseService(cfg *config.Config) (*FirebaseService, error) {
	if cfg.FirebaseProjectID == "" {
		return nil, fmt.Errorf("Firebase project ID is required")
	}

	log.Printf("🔥 Initializing Firebase REST API for project: %s", cfg.FirebaseProjectID)

	return &FirebaseService{
		projectID: cfg.FirebaseProjectID,
		apiKey:    cfg.FirebaseAPIKey,
		baseURL:   fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents", cfg.FirebaseProjectID),
		client:    &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (s *FirebaseService) Close() error {
	log.Printf("🔥 Firebase service closed")
	return nil
}

// Helper function to make HTTP requests to Firestore REST API
func (s *FirebaseService) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	url := s.baseURL + path
	if s.apiKey != "" {
		url += "?key=" + s.apiKey
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	return s.client.Do(req)
}

// Convert our models to Firestore document format
func (s *FirebaseService) toFirestoreDoc(data interface{}) map[string]interface{} {
	doc := map[string]interface{}{
		"fields": make(map[string]interface{}),
	}
	fields := doc["fields"].(map[string]interface{})

	// Convert based on type
	switch v := data.(type) {
	case *models.UserSession:
		fields["userId"] = map[string]interface{}{"stringValue": v.UserID}
		fields["email"] = map[string]interface{}{"stringValue": v.Email}
		fields["name"] = map[string]interface{}{"stringValue": v.Name}
		fields["accessToken"] = map[string]interface{}{"stringValue": v.AccessToken}
		if v.RefreshToken != nil {
			fields["refreshToken"] = map[string]interface{}{"stringValue": *v.RefreshToken}
		}
		fields["createdAt"] = map[string]interface{}{"timestampValue": v.CreatedAt.Format(time.RFC3339)}
		fields["lastLogin"] = map[string]interface{}{"timestampValue": v.LastLogin.Format(time.RFC3339)}

	case *models.Task:
		fields["userId"] = map[string]interface{}{"stringValue": v.UserID}
		fields["title"] = map[string]interface{}{"stringValue": v.Title}
		if v.Description != nil {
			fields["description"] = map[string]interface{}{"stringValue": *v.Description}
		}
		fields["completed"] = map[string]interface{}{"booleanValue": v.Completed}
		fields["status"] = map[string]interface{}{"stringValue": v.Status}
		fields["priority"] = map[string]interface{}{"stringValue": v.Priority}
		if v.StartDate != nil {
			fields["startDate"] = map[string]interface{}{"timestampValue": v.StartDate.Format(time.RFC3339)}
		}
		if v.DueDate != nil {
			fields["dueDate"] = map[string]interface{}{"timestampValue": v.DueDate.Format(time.RFC3339)}
		}
		if v.EstimatedHours != nil {
			fields["estimatedHours"] = map[string]interface{}{"integerValue": fmt.Sprintf("%d", *v.EstimatedHours)}
		}
		if v.ActualHours != nil {
			fields["actualHours"] = map[string]interface{}{"integerValue": fmt.Sprintf("%d", *v.ActualHours)}
		}
		fields["createdAt"] = map[string]interface{}{"timestampValue": v.CreatedAt.Format(time.RFC3339)}
		fields["updatedAt"] = map[string]interface{}{"timestampValue": v.UpdatedAt.Format(time.RFC3339)}
	}

	return doc
}

// Convert Firestore document back to our models
func (s *FirebaseService) fromFirestoreDoc(doc map[string]interface{}, result interface{}) error {
	fields, ok := doc["fields"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid firestore document format")
	}

	switch v := result.(type) {
	case *models.UserSession:
		if userId, ok := s.getStringValue(fields, "userId"); ok {
			v.UserID = userId
		}
		if email, ok := s.getStringValue(fields, "email"); ok {
			v.Email = email
		}
		if name, ok := s.getStringValue(fields, "name"); ok {
			v.Name = name
		}
		if accessToken, ok := s.getStringValue(fields, "accessToken"); ok {
			v.AccessToken = accessToken
		}
		if refreshToken, ok := s.getStringValue(fields, "refreshToken"); ok {
			v.RefreshToken = &refreshToken
		}
		if createdAt, ok := s.getTimestampValue(fields, "createdAt"); ok {
			v.CreatedAt = createdAt
		}
		if lastLogin, ok := s.getTimestampValue(fields, "lastLogin"); ok {
			v.LastLogin = lastLogin
		}

	case *models.Task:
		if userId, ok := s.getStringValue(fields, "userId"); ok {
			v.UserID = userId
		}
		if title, ok := s.getStringValue(fields, "title"); ok {
			v.Title = title
		}
		if description, ok := s.getStringValue(fields, "description"); ok {
			v.Description = &description
		}
		if completed, ok := s.getBooleanValue(fields, "completed"); ok {
			v.Completed = completed
		}
		if status, ok := s.getStringValue(fields, "status"); ok {
			v.Status = status
		}
		if priority, ok := s.getStringValue(fields, "priority"); ok {
			v.Priority = priority
		}
		if startDate, ok := s.getTimestampValue(fields, "startDate"); ok {
			v.StartDate = &startDate
		}
		if dueDate, ok := s.getTimestampValue(fields, "dueDate"); ok {
			v.DueDate = &dueDate
		}
		if createdAt, ok := s.getTimestampValue(fields, "createdAt"); ok {
			v.CreatedAt = createdAt
		}
		if updatedAt, ok := s.getTimestampValue(fields, "updatedAt"); ok {
			v.UpdatedAt = updatedAt
		}
	}

	return nil
}

// Helper functions to extract values from Firestore fields
func (s *FirebaseService) getStringValue(fields map[string]interface{}, key string) (string, bool) {
	if field, ok := fields[key].(map[string]interface{}); ok {
		if value, ok := field["stringValue"].(string); ok {
			return value, true
		}
	}
	return "", false
}

func (s *FirebaseService) getBooleanValue(fields map[string]interface{}, key string) (bool, bool) {
	if field, ok := fields[key].(map[string]interface{}); ok {
		if value, ok := field["booleanValue"].(bool); ok {
			return value, true
		}
	}
	return false, false
}

func (s *FirebaseService) getTimestampValue(fields map[string]interface{}, key string) (time.Time, bool) {
	if field, ok := fields[key].(map[string]interface{}); ok {
		if value, ok := field["timestampValue"].(string); ok {
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}

// User operations
func (s *FirebaseService) CreateUser(user *models.UserSession) error {
	doc := s.toFirestoreDoc(user)
	resp, err := s.makeRequest("POST", "/users", doc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create user: %s", body)
	}

	log.Printf("✅ User created: %s", user.Email)
	return nil
}

func (s *FirebaseService) GetUser(userID string) (*models.UserSession, error) {
	resp, err := s.makeRequest("GET", "/users/"+userID, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user not found")
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to get user")
	}

	var doc map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	var user models.UserSession
	if err := s.fromFirestoreDoc(doc, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *FirebaseService) UpdateUser(userID string, updates map[string]interface{}) error {
	// Create update document
	doc := map[string]interface{}{
		"fields": make(map[string]interface{}),
	}
	fields := doc["fields"].(map[string]interface{})

	// Add lastLogin timestamp
	updates["lastLogin"] = time.Now()

	for key, value := range updates {
		switch v := value.(type) {
		case string:
			fields[key] = map[string]interface{}{"stringValue": v}
		case time.Time:
			fields[key] = map[string]interface{}{"timestampValue": v.Format(time.RFC3339)}
		case bool:
			fields[key] = map[string]interface{}{"booleanValue": v}
		}
	}

	resp, err := s.makeRequest("PATCH", "/users/"+userID, doc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update user: %s", body)
	}

	return nil
}

// Task operations
func (s *FirebaseService) CreateTask(task *models.Task) (string, error) {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	doc := s.toFirestoreDoc(task)
	resp, err := s.makeRequest("POST", "/tasks", doc)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to create task: %s", body)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	// Extract document ID from the response
	if name, ok := result["name"].(string); ok {
		parts := strings.Split(name, "/")
		if len(parts) > 0 {
			docID := parts[len(parts)-1]
			log.Printf("✅ Task created: %s (ID: %s)", task.Title, docID)
			return docID, nil
		}
	}

	return "", fmt.Errorf("failed to extract document ID")
}

func (s *FirebaseService) GetTasks(userID string) ([]*models.Task, error) {
	log.Printf("🔍 Fetching tasks for user: %s", userID)

	// For simplicity, we'll get all tasks and filter client-side
	// In a real implementation, you'd use Firestore queries
	resp, err := s.makeRequest("GET", "/tasks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return []*models.Task{}, nil // Return empty array instead of error
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return []*models.Task{}, nil
	}

	var tasks []*models.Task
	if documents, ok := result["documents"].([]interface{}); ok {
		for _, docInterface := range documents {
			if doc, ok := docInterface.(map[string]interface{}); ok {
				var task models.Task
				if err := s.fromFirestoreDoc(doc, &task); err == nil {
					// Filter by user ID
					if task.UserID == userID {
						// Extract document ID
						if name, ok := doc["name"].(string); ok {
							parts := strings.Split(name, "/")
							if len(parts) > 0 {
								task.ID = parts[len(parts)-1]
							}
						}
						tasks = append(tasks, &task)
					}
				}
			}
		}
	}

	log.Printf("✅ Found %d tasks for user %s", len(tasks), userID)
	return tasks, nil
}

func (s *FirebaseService) UpdateTask(taskID string, updates map[string]interface{}) error {
	// Create update document
	doc := map[string]interface{}{
		"fields": make(map[string]interface{}),
	}
	fields := doc["fields"].(map[string]interface{})

	updates["updatedAt"] = time.Now()

	for key, value := range updates {
		switch v := value.(type) {
		case string:
			fields[key] = map[string]interface{}{"stringValue": v}
		case time.Time:
			fields[key] = map[string]interface{}{"timestampValue": v.Format(time.RFC3339)}
		case bool:
			fields[key] = map[string]interface{}{"booleanValue": v}
		}
	}

	resp, err := s.makeRequest("PATCH", "/tasks/"+taskID, doc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update task: %s", body)
	}

	return nil
}

func (s *FirebaseService) DeleteTask(taskID string) error {
	resp, err := s.makeRequest("DELETE", "/tasks/"+taskID, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete task: %s", body)
	}

	return nil
}

// Simplified implementations for meetings and reminders
func (s *FirebaseService) CreateMeeting(meeting *models.Meeting) (string, error) {
	log.Printf("📅 Meeting creation not fully implemented yet")
	return "meeting-id", nil
}

func (s *FirebaseService) GetMeetings(userID string) ([]*models.Meeting, error) {
	log.Printf("📅 Getting meetings for user: %s", userID)
	return []*models.Meeting{}, nil
}

func (s *FirebaseService) UpdateMeeting(meetingID string, updates map[string]interface{}) error {
	log.Printf("📅 Meeting update not fully implemented yet")
	return nil
}

func (s *FirebaseService) CreateReminder(reminder *models.Reminder) (string, error) {
	log.Printf("⏰ Reminder creation not fully implemented yet")
	return "reminder-id", nil
}

func (s *FirebaseService) GetReminders(userID string) ([]*models.Reminder, error) {
	log.Printf("⏰ Getting reminders for user: %s", userID)
	return []*models.Reminder{}, nil
}

func (s *FirebaseService) UpdateReminder(reminderID string, updates map[string]interface{}) error {
	log.Printf("⏰ Reminder update not fully implemented yet")
	return nil
}

func (s *FirebaseService) GetAllTasks() ([]*models.Task, error) {
	return []*models.Task{}, nil
}

func (s *FirebaseService) GetAllMeetings() ([]*models.Meeting, error) {
	return []*models.Meeting{}, nil
}

func (s *FirebaseService) GetAllReminders() ([]*models.Reminder, error) {
	return []*models.Reminder{}, nil
}