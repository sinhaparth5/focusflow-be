package services

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"

	"focusflow-be/internal/config"
	"focusflow-be/internal/models"
)

type GoogleService struct {
	config      *config.Config
	oauthConfig *oauth2.Config
}

func NewGoogleService(cfg *config.Config) *GoogleService {
	oauthConfig := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURI,
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/calendar",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleService{
		config:      cfg,
		oauthConfig: oauthConfig,
	}
}

func (s *GoogleService) GetAuthURL() string {
	return s.oauthConfig.AuthCodeURL("state", oauth2.AccessTypeOffline)
}

func (s *GoogleService) ExchangeCodeForToken(code string) (*oauth2.Token, error) {
	return s.oauthConfig.Exchange(context.Background(), code)
}

func (s *GoogleService) GetUserInfo(token *oauth2.Token) (*models.GoogleUserInfo, error) {
	client := s.oauthConfig.Client(context.Background(), token)
	
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var userInfo models.GoogleUserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *GoogleService) CreateCalendarEvent(token *oauth2.Token, task *models.Task) (string, error) {
	if task.DueDate == nil {
		return "", nil
	}

	ctx := context.Background()
	client := s.oauthConfig.Client(ctx, token)
	
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	startTime := time.Now()
	if task.StartDate != nil {
		startTime = *task.StartDate
	}

	event := &calendar.Event{
		Summary:     task.Title,
		Description: func() string {
			if task.Description != nil {
				return *task.Description
			}
			return ""
		}(),
		Start: &calendar.EventDateTime{
			DateTime: startTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: task.DueDate.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		ColorId: func() string {
			switch task.Priority {
			case "high":
				return "11"
			case "medium":
				return "5"
			default:
				return "2"
			}
		}(),
	}

	createdEvent, err := calendarService.Events.Insert("primary", event).Do()
	if err != nil {
		return "", err
	}

	return createdEvent.Id, nil
}

func (s *GoogleService) CreateCalendarMeeting(token *oauth2.Token, meeting *models.Meeting) (string, error) {
	ctx := context.Background()
	client := s.oauthConfig.Client(ctx, token)
	
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	event := &calendar.Event{
		Summary:     meeting.Title,
		Description: func() string {
			if meeting.Description != nil {
				return *meeting.Description
			}
			return ""
		}(),
		Start: &calendar.EventDateTime{
			DateTime: meeting.StartTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: meeting.EndTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Location: func() string {
			if meeting.Location != nil {
				return *meeting.Location
			}
			return ""
		}(),
		Attendees: func() []*calendar.EventAttendee {
			var attendees []*calendar.EventAttendee
			for _, email := range meeting.Attendees {
				attendees = append(attendees, &calendar.EventAttendee{
					Email: email,
				})
			}
			return attendees
		}(),
		ColorId: "9",
	}

	createdEvent, err := calendarService.Events.Insert("primary", event).Do()
	if err != nil {
		return "", err
	}

	return createdEvent.Id, nil
}

func (s *GoogleService) CreateCalendarReminder(token *oauth2.Token, reminder *models.Reminder) (string, error) {
	ctx := context.Background()
	client := s.oauthConfig.Client(ctx, token)
	
	calendarService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return "", err
	}

	endTime := reminder.ReminderTime.Add(15 * time.Minute)

	event := &calendar.Event{
		Summary:     reminder.Title,
		Description: func() string {
			if reminder.Description != nil {
				return *reminder.Description
			}
			return ""
		}(),
		Start: &calendar.EventDateTime{
			DateTime: reminder.ReminderTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		End: &calendar.EventDateTime{
			DateTime: endTime.Format(time.RFC3339),
			TimeZone: "UTC",
		},
		Reminders: &calendar.EventReminders{
			UseDefault: false,
			Overrides: []*calendar.EventReminder{
				{
					Method:  "popup",
					Minutes: 10,
				},
				{
					Method:  "email",
					Minutes: 30,
				},
			},
		},
		ColorId: "8",
	}

	createdEvent, err := calendarService.Events.Insert("primary", event).Do()
	if err != nil {
		return "", err
	}

	return createdEvent.Id, nil
}