package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"focusflow-be/internal/config"
	"focusflow-be/internal/models"
)

type FirebaseService struct {
	projectID string
	apiKey    string
	baseURL   string
}

func NewFirebaseService(cfg *config.Config) (*FirebaseService, error) {
	if cfg.FirebaseProjectID == "" || cfg.FirebaseAPIKey == "" {
		return nil, fmt.Errorf("Firebase project ID and API key are required")
	}

	return &FirebaseService{
		projectID: cfg.FirebaseProjectID,
		apiKey:    cfg.FirebaseAPIKey,
		baseURL:   fmt.Sprintf("https://firestore.googleapis.com/v1/projects/%s/databases/(default)/documents", cfg.FirebaseProjectID),
	}, nil
}

func (s *FirebaseService) Close() error {
	// No connection to close with REST API
	return nil
}

// Helper function to make HTTP requests to Firestore REST API
func (s *FirebaseService) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s?key=%s", s.baseURL, path, s.apiKey)
	
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
	
	client := &http.Client{Timeout: 30 * time.Second}
	return client.Do(req)
}

// Convert Firestore document format to our models
func (s *FirebaseService) parseFirestoreDoc(doc map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	if fields, ok := doc["fields"].(map[string]interface{}); ok {
		for key, value := range fields {
			if valueMap, ok := value.(map[string]interface{}); ok {
				// Extract the actual value based on Firestore type
				if stringValue, ok := valueMap["stringValue"]; ok {
					result[key] = stringValue
				} else if boolValue, ok := valueMap["booleanValue"]; ok {
					result[key] = boolValue
				} else if timestampValue, ok := valueMap["timestampValue"]; ok {
					if t, err := time.Parse(time.RFC3339, timestampValue.(string)); err == nil {
						result[key] = t
					}
				}
				// Add more type conversions as needed
			}
		}
	}
	
	return result
}

// User operations
func (s *FirebaseService) CreateUser(user *models.UserSession) error {
	// Convert user to Firestore format
	firestoreDoc := map[string]interface{}{
		"fields": map[string]interface{}{
			"userId":       map[string]interface{}{"stringValue": user.UserID},
			"email":        map[string]interface{}{"stringValue": user.Email},
			"name":         map[string]interface{}{"stringValue": user.Name},
			"accessToken":  map[string]interface{}{"stringValue": user.AccessToken},
			"createdAt":    map[string]interface{}{"timestampValue": user.CreatedAt.Format(time.RFC3339)},
			"lastLogin":    map[string]interface{}{"timestampValue": user.LastLogin.Format(time.RFC3339)},
		},
	}

	resp, err := s.makeRequest("POST", "/users", firestoreDoc)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create user: %s", body)
	}

	return nil
}

func (s *FirebaseService) GetUser(userID string) (*models.UserSession, error) {
	resp, err := s.makeRequest("GET", fmt.Sprintf("/users/%s", userID), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found")
	}

	var doc map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, err
	}

	fields := s.parseFirestoreDoc(doc)
	
	user := &models.UserSession{
		UserID:      fields["userId"].(string),
		Email:       fields["email"].(string),
		Name:        fields["name"].(string),
		AccessToken: fields["accessToken"].(string),
	}

	if createdAt, ok := fields["createdAt"].(time.Time); ok {
		user.CreatedAt = createdAt
	}
	if lastLogin, ok := fields["lastLogin"].(time.Time); ok {
		user.LastLogin = lastLogin
	}

	return user, nil
}

func (s *FirebaseService) UpdateUser(userID string, updates map[string]interface{}) error {
	// Convert updates to Firestore format
	firestoreUpdates := map[string]interface{}{
		"fields": make(map[string]interface{}),
	}
	
	fields := firestoreUpdates["fields"].(map[string]interface{})
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

	resp, err := s.makeRequest("PATCH", fmt.Sprintf("/users/%s", userID), firestoreUpdates)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update user: %s", body)
	}

	return nil
}

// Simplified task operations (you would implement similar patterns for all CRUD operations)
func (s *FirebaseService) CreateTask(task *models.Task) (string, error) {
	// Implementation similar to CreateUser but for tasks
	// This is a simplified version - you'd need to implement full Firestore format conversion
	return "", fmt.Errorf("not implemented - use the simplified version above as a template")
}

func (s *FirebaseService) GetTasks(userID string) ([]*models.Task, error) {
	// Implementation for getting tasks using REST API
	return nil, fmt.Errorf("not implemented - use the simplified version above as a template")
}

// Add other required methods with similar implementations...
func (s *FirebaseService) UpdateTask(taskID string, updates map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *FirebaseService) DeleteTask(taskID string) error {
	return fmt.Errorf("not implemented")
}

func (s *FirebaseService) CreateMeeting(meeting *models.Meeting) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *FirebaseService) GetMeetings(userID string) ([]*models.Meeting, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *FirebaseService) UpdateMeeting(meetingID string, updates map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *FirebaseService) CreateReminder(reminder *models.Reminder) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (s *FirebaseService) GetReminders(userID string) ([]*models.Reminder, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *FirebaseService) UpdateReminder(reminderID string, updates map[string]interface{}) error {
	return fmt.Errorf("not implemented")
}

func (s *FirebaseService) GetAllTasks() ([]*models.Task, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *FirebaseService) GetAllMeetings() ([]*models.Meeting, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *FirebaseService) GetAllReminders() ([]*models.Reminder, error) {
	return nil, fmt.Errorf("not implemented")
}