# FocusFlow Backend API

A Go-based task management REST API with Google OAuth authentication, Firebase Firestore storage, and Google Calendar integration.

## 🚀 Live API

**Base URL**: `https://focusflow-be-production.up.railway.app`

## 📋 Quick Start

### Authentication
1. Visit: `https://focusflow-be-production.up.railway.app/auth/google`
2. Complete Google OAuth login
3. Copy JWT token from success page
4. Use in API requests: `Authorization: Bearer <your-token>`

## 📚 API Endpoints

### Authentication
- `GET /auth/google` - Start OAuth flow
- `GET /auth/me` - Get current user

### Tasks
- `GET /tasks` - Get all tasks
- `POST /tasks` - Create task
- `PUT /tasks/:id` - Update task
- `PATCH /tasks/:id/start` - Start task
- `PATCH /tasks/:id/complete` - Complete task
- `DELETE /tasks/:id` - Delete task

### Meetings
- `GET /meetings` - Get all meetings
- `POST /meetings` - Create meeting
- `PATCH /meetings/:id/status` - Update meeting status

### Reminders
- `GET /reminders` - Get all reminders
- `POST /reminders` - Create reminder
- `PATCH /reminders/:id/complete` - Complete reminder

### Dashboard
- `GET /dashboard/calendar` - Calendar events
- `GET /dashboard/gantt` - Gantt chart data
- `GET /dashboard/overview` - Statistics overview

## 📝 Example Requests

### Create Task
```bash
curl -X POST https://focusflow-be-production.up.railway.app/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Complete project",
    "description": "Finish the API",
    "priority": "high",
    "dueDate": "2025-01-15T17:00:00Z"
  }'
```

### Get Tasks
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     https://focusflow-be-production.up.railway.app/tasks
```

### Create Meeting
```bash
curl -X POST https://focusflow-be-production.up.railway.app/meetings \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Team Standup",
    "startTime": "2025-01-01T10:00:00Z",
    "endTime": "2025-01-01T10:30:00Z",
    "meetingType": "video"
  }'
```

## 🏗️ Local Development

### Prerequisites
- Go 1.21+
- Firebase project
- Google OAuth credentials

### Setup
```bash
# Clone repository
git clone <your-repo>
cd focusflow-backend

# Install dependencies
go mod download

# Set environment variables
cp .env.example .env
# Edit .env with your credentials

# Run application
go run main.go
```

### Environment Variables
```env
PORT=8080
FIREBASE_PROJECT_ID=your-firebase-project-id
FIREBASE_API_KEY=your-firebase-api-key
FIREBASE_AUTH_DOMAIN=your-project.firebaseapp.com
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8080/auth/callback
JWT_SECRET=your-super-secure-jwt-secret-32-chars-min
```

## 🚀 Deployment

### Railway (Current)
1. Push to GitHub
2. Connect repository to Railway
3. Set environment variables in Railway dashboard
4. Deploy automatically

### Docker
```bash
# Build image
docker build -t focusflow-backend .

# Run container
docker run -p 8080:8080 --env-file .env focusflow-backend
```

### Other Platforms
- **Heroku**: `git push heroku main`
- **Google Cloud Run**: `gcloud run deploy`
- **AWS ECS**: Use provided Dockerfile
- **DigitalOcean**: Use App Platform

## 🔧 Tech Stack

- **Language**: Go 1.21
- **Framework**: Gin
- **Database**: Firebase Firestore
- **Authentication**: Google OAuth 2.0 + JWT
- **Calendar**: Google Calendar API
- **Deployment**: Railway

## 📊 Data Models

### Task
```json
{
  "title": "string (required)",
  "description": "string (optional)",
  "priority": "low|medium|high (required)",
  "status": "todo|in-progress|completed",
  "startDate": "ISO 8601 date",
  "dueDate": "ISO 8601 date",
  "estimatedHours": "number"
}
```

### Meeting
```json
{
  "title": "string (required)",
  "startTime": "ISO 8601 date (required)",
  "endTime": "ISO 8601 date (required)",
  "meetingType": "call|in-person|video (required)",
  "attendees": ["email1", "email2"],
  "location": "string"
}
```

### Reminder
```json
{
  "title": "string (required)",
  "reminderTime": "ISO 8601 date (required)",
  "reminderType": "task|meeting|personal (required)",
  "priority": "low|medium|high (required)",
  "description": "string"
}
```

## 🔒 Security

- Google OAuth 2.0 authentication
- JWT tokens (24-hour expiration)
- HTTPS enforcement
- CORS enabled
- User data isolation

## 📈 Status

✅ **Production Ready**  
✅ **Live on Railway**  
✅ **Google OAuth Working**  
✅ **Tasks CRUD Complete**  
✅ **Meetings & Reminders**  
✅ **Dashboard Analytics**

## 🆘 Support

For issues or questions:
1. Check Railway logs for errors
2. Verify environment variables
3. Test authentication flow
4. Validate request format

**Health Check**: `GET https://focusflow-be-production.up.railway.app/`