package services

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"focusflow-be/internal/config"
	"focusflow-be/internal/models"
)

type FirebaseService struct {
	client *firestore.Client
	ctx    context.Context
}

func NewFirebaseService(cfg *config.Config) (*FirebaseService, error) {
	ctx := context.Background()

	var app *firebase.App
	var err error

	// Try different authentication methods
	serviceAccountPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	
	if serviceAccountPath != "" && fileExists(serviceAccountPath) {
		// Use service account key if available
		log.Printf("Using service account key: %s", serviceAccountPath)
		opt := option.WithCredentialsFile(serviceAccountPath)
		app, err = firebase.NewApp(ctx, &firebase.Config{
			ProjectID: cfg.FirebaseProjectID,
		}, opt)
	} else {
		// Use Application Default Credentials (works with gcloud auth)
		log.Printf("Using Application Default Credentials for project: %s", cfg.FirebaseProjectID)
		app, err = firebase.NewApp(ctx, &firebase.Config{
			ProjectID: cfg.FirebaseProjectID,
		})
	}

	if err != nil {
		return nil, err
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, err
	}

	log.Printf("✅ Firebase Firestore initialized successfully")
	return &FirebaseService{
		client: client,
		ctx:    ctx,
	}, nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func (s *FirebaseService) Close() error {
	return s.client.Close()
}

// User operations
func (s *FirebaseService) CreateUser(user *models.UserSession) error {
	_, err := s.client.Collection("users").Doc(user.UserID).Set(s.ctx, user)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		return err
	}
	log.Printf("✅ User created: %s", user.Email)
	return nil
}

func (s *FirebaseService) GetUser(userID string) (*models.UserSession, error) {
	doc, err := s.client.Collection("users").Doc(userID).Get(s.ctx)
	if err != nil {
		return nil, err
	}

	var user models.UserSession
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *FirebaseService) UpdateUser(userID string, updates map[string]interface{}) error {
	updates["lastLogin"] = time.Now()
	
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("users").Doc(userID).Update(s.ctx, firebaseUpdates)
	return err
}

// Task operations
func (s *FirebaseService) CreateTask(task *models.Task) (string, error) {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("tasks").Add(s.ctx, task)
	if err != nil {
		log.Printf("Error creating task: %v", err)
		return "", err
	}
	
	log.Printf("✅ Task created: %s (ID: %s)", task.Title, docRef.ID)
	return docRef.ID, nil
}

func (s *FirebaseService) GetTasks(userID string) ([]*models.Task, error) {
	log.Printf("🔍 Fetching tasks for user: %s", userID)
	
	iter := s.client.Collection("tasks").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var tasks []*models.Task
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var task models.Task
		if err := doc.DataTo(&task); err != nil {
			log.Printf("Error parsing task document: %v", err)
			continue
		}
		task.ID = doc.Ref.ID
		tasks = append(tasks, &task)
	}

	log.Printf("✅ Found %d tasks for user %s", len(tasks), userID)
	return tasks, nil
}

func (s *FirebaseService) UpdateTask(taskID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()
	
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("tasks").Doc(taskID).Update(s.ctx, firebaseUpdates)
	if err != nil {
		log.Printf("Error updating task %s: %v", taskID, err)
	}
	return err
}

func (s *FirebaseService) DeleteTask(taskID string) error {
	_, err := s.client.Collection("tasks").Doc(taskID).Delete(s.ctx)
	if err != nil {
		log.Printf("Error deleting task %s: %v", taskID, err)
	}
	return err
}

// Meeting operations
func (s *FirebaseService) CreateMeeting(meeting *models.Meeting) (string, error) {
	meeting.CreatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("meetings").Add(s.ctx, meeting)
	if err != nil {
		log.Printf("Error creating meeting: %v", err)
		return "", err
	}
	
	log.Printf("✅ Meeting created: %s (ID: %s)", meeting.Title, docRef.ID)
	return docRef.ID, nil
}

func (s *FirebaseService) GetMeetings(userID string) ([]*models.Meeting, error) {
	log.Printf("🔍 Fetching meetings for user: %s", userID)
	
	iter := s.client.Collection("meetings").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var meetings []*models.Meeting
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var meeting models.Meeting
		if err := doc.DataTo(&meeting); err != nil {
			log.Printf("Error parsing meeting document: %v", err)
			continue
		}
		meeting.ID = doc.Ref.ID
		meetings = append(meetings, &meeting)
	}

	log.Printf("✅ Found %d meetings for user %s", len(meetings), userID)
	return meetings, nil
}

func (s *FirebaseService) UpdateMeeting(meetingID string, updates map[string]interface{}) error {
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("meetings").Doc(meetingID).Update(s.ctx, firebaseUpdates)
	if err != nil {
		log.Printf("Error updating meeting %s: %v", meetingID, err)
	}
	return err
}

// Reminder operations
func (s *FirebaseService) CreateReminder(reminder *models.Reminder) (string, error) {
	reminder.CreatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("reminders").Add(s.ctx, reminder)
	if err != nil {
		log.Printf("Error creating reminder: %v", err)
		return "", err
	}
	
	log.Printf("✅ Reminder created: %s (ID: %s)", reminder.Title, docRef.ID)
	return docRef.ID, nil
}

func (s *FirebaseService) GetReminders(userID string) ([]*models.Reminder, error) {
	log.Printf("🔍 Fetching reminders for user: %s", userID)
	
	iter := s.client.Collection("reminders").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var reminders []*models.Reminder
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var reminder models.Reminder
		if err := doc.DataTo(&reminder); err != nil {
			log.Printf("Error parsing reminder document: %v", err)
			continue
		}
		reminder.ID = doc.Ref.ID
		reminders = append(reminders, &reminder)
	}

	log.Printf("✅ Found %d reminders for user %s", len(reminders), userID)
	return reminders, nil
}

func (s *FirebaseService) UpdateReminder(reminderID string, updates map[string]interface{}) error {
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("reminders").Doc(reminderID).Update(s.ctx, firebaseUpdates)
	if err != nil {
		log.Printf("Error updating reminder %s: %v", reminderID, err)
	}
	return err
}

// Dashboard operations
func (s *FirebaseService) GetAllTasks() ([]*models.Task, error) {
	iter := s.client.Collection("tasks").Documents(s.ctx)
	defer iter.Stop()

	var tasks []*models.Task
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var task models.Task
		if err := doc.DataTo(&task); err != nil {
			continue
		}
		task.ID = doc.Ref.ID
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *FirebaseService) GetAllMeetings() ([]*models.Meeting, error) {
	iter := s.client.Collection("meetings").Documents(s.ctx)
	defer iter.Stop()

	var meetings []*models.Meeting
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var meeting models.Meeting
		if err := doc.DataTo(&meeting); err != nil {
			continue
		}
		meeting.ID = doc.Ref.ID
		meetings = append(meetings, &meeting)
	}

	return meetings, nil
}

func (s *FirebaseService) GetAllReminders() ([]*models.Reminder, error) {
	iter := s.client.Collection("reminders").Documents(s.ctx)
	defer iter.Stop()

	var reminders []*models.Reminder
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var reminder models.Reminder
		if err := doc.DataTo(&reminder); err != nil {
			continue
		}
		reminder.ID = doc.Ref.ID
		reminders = append(reminders, &reminder)
	}

	return reminders, nil
}

func (s *FirebaseService) Close() error {
	return s.client.Close()
}

// User operations
func (s *FirebaseService) CreateUser(user *models.UserSession) error {
	_, err := s.client.Collection("users").Doc(user.UserID).Set(s.ctx, user)
	return err
}

func (s *FirebaseService) GetUser(userID string) (*models.UserSession, error) {
	doc, err := s.client.Collection("users").Doc(userID).Get(s.ctx)
	if err != nil {
		return nil, err
	}

	var user models.UserSession
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *FirebaseService) UpdateUser(userID string, updates map[string]interface{}) error {
	updates["lastLogin"] = time.Now()
	_, err := s.client.Collection("users").Doc(userID).Update(s.ctx, []firestore.Update{
		{Path: "accessToken", Value: updates["accessToken"]},
		{Path: "refreshToken", Value: updates["refreshToken"]},
		{Path: "lastLogin", Value: updates["lastLogin"]},
	})
	return err
}

// Task operations
func (s *FirebaseService) CreateTask(task *models.Task) (string, error) {
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("tasks").Add(s.ctx, task)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (s *FirebaseService) GetTasks(userID string) ([]*models.Task, error) {
	iter := s.client.Collection("tasks").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var tasks []*models.Task
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var task models.Task
		if err := doc.DataTo(&task); err != nil {
			continue
		}
		task.ID = doc.Ref.ID
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *FirebaseService) UpdateTask(taskID string, updates map[string]interface{}) error {
	updates["updatedAt"] = time.Now()
	
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("tasks").Doc(taskID).Update(s.ctx, firebaseUpdates)
	return err
}

func (s *FirebaseService) DeleteTask(taskID string) error {
	_, err := s.client.Collection("tasks").Doc(taskID).Delete(s.ctx)
	return err
}

// Meeting operations
func (s *FirebaseService) CreateMeeting(meeting *models.Meeting) (string, error) {
	meeting.CreatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("meetings").Add(s.ctx, meeting)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (s *FirebaseService) GetMeetings(userID string) ([]*models.Meeting, error) {
	iter := s.client.Collection("meetings").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var meetings []*models.Meeting
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var meeting models.Meeting
		if err := doc.DataTo(&meeting); err != nil {
			continue
		}
		meeting.ID = doc.Ref.ID
		meetings = append(meetings, &meeting)
	}

	return meetings, nil
}

func (s *FirebaseService) UpdateMeeting(meetingID string, updates map[string]interface{}) error {
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("meetings").Doc(meetingID).Update(s.ctx, firebaseUpdates)
	return err
}

// Reminder operations
func (s *FirebaseService) CreateReminder(reminder *models.Reminder) (string, error) {
	reminder.CreatedAt = time.Now()
	
	docRef, _, err := s.client.Collection("reminders").Add(s.ctx, reminder)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (s *FirebaseService) GetReminders(userID string) ([]*models.Reminder, error) {
	iter := s.client.Collection("reminders").Where("userId", "==", userID).Documents(s.ctx)
	defer iter.Stop()

	var reminders []*models.Reminder
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var reminder models.Reminder
		if err := doc.DataTo(&reminder); err != nil {
			continue
		}
		reminder.ID = doc.Ref.ID
		reminders = append(reminders, &reminder)
	}

	return reminders, nil
}

func (s *FirebaseService) UpdateReminder(reminderID string, updates map[string]interface{}) error {
	var firebaseUpdates []firestore.Update
	for key, value := range updates {
		firebaseUpdates = append(firebaseUpdates, firestore.Update{
			Path:  key,
			Value: value,
		})
	}

	_, err := s.client.Collection("reminders").Doc(reminderID).Update(s.ctx, firebaseUpdates)
	return err
}

// Dashboard operations
func (s *FirebaseService) GetAllTasks() ([]*models.Task, error) {
	iter := s.client.Collection("tasks").Documents(s.ctx)
	defer iter.Stop()

	var tasks []*models.Task
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var task models.Task
		if err := doc.DataTo(&task); err != nil {
			continue
		}
		task.ID = doc.Ref.ID
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (s *FirebaseService) GetAllMeetings() ([]*models.Meeting, error) {
	iter := s.client.Collection("meetings").Documents(s.ctx)
	defer iter.Stop()

	var meetings []*models.Meeting
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var meeting models.Meeting
		if err := doc.DataTo(&meeting); err != nil {
			continue
		}
		meeting.ID = doc.Ref.ID
		meetings = append(meetings, &meeting)
	}

	return meetings, nil
}

func (s *FirebaseService) GetAllReminders() ([]*models.Reminder, error) {
	iter := s.client.Collection("reminders").Documents(s.ctx)
	defer iter.Stop()

	var reminders []*models.Reminder
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var reminder models.Reminder
		if err := doc.DataTo(&reminder); err != nil {
			continue
		}
		reminder.ID = doc.Ref.ID
		reminders = append(reminders, &reminder)
	}

	return reminders, nil
}