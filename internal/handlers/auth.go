package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"focusflow-be/internal/models"
	"focusflow-be/internal/services"
)

type AuthHandler struct {
	authService     *services.AuthService
	googleService   *services.GoogleService
	firebaseService *services.FirebaseService
}

func NewAuthHandler(authService *services.AuthService, googleService *services.GoogleService, firebaseService *services.FirebaseService) *AuthHandler {
	return &AuthHandler{
		authService:     authService,
		googleService:   googleService,
		firebaseService: firebaseService,
	}
}

func (h *AuthHandler) GoogleAuth(c *gin.Context) {
	url := h.googleService.GetAuthURL()
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error": "No authorization code provided",
		})
		return
	}

	// Handle OAuth errors
	if errorParam := c.Query("error"); errorParam != "" {
		errorDesc := c.Query("error_description")
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       errorParam,
			"description": errorDesc,
		})
		return
	}

	token, err := h.googleService.ExchangeCodeForToken(code)
	if err != nil {
		log.Printf("Token exchange error: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "Token exchange failed",
			"description": err.Error(),
		})
		return
	}

	userInfo, err := h.googleService.GetUserInfo(token)
	if err != nil {
		log.Printf("Get user info error: %v", err)
		c.HTML(http.StatusBadRequest, "error.html", gin.H{
			"error":       "Failed to get user info",
			"description": err.Error(),
		})
		return
	}

	userSession := &models.UserSession{
		UserID:       userInfo.ID,
		Email:        userInfo.Email,
		Name:         userInfo.Name,
		AccessToken:  token.AccessToken,
		RefreshToken: &token.RefreshToken,
		CreatedAt:    time.Now(),
		LastLogin:    time.Now(),
	}

	// Check if user exists
	existingUser, err := h.firebaseService.GetUser(userInfo.ID)
	if err != nil {
		// User doesn't exist, create new one
		if err := h.firebaseService.CreateUser(userSession); err != nil {
			log.Printf("Create user error: %v", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{
				"error":       "Failed to create user",
				"description": err.Error(),
			})
			return
		}
	} else {
		// Update existing user
		updates := map[string]interface{}{
			"accessToken":  token.AccessToken,
			"refreshToken": &token.RefreshToken,
		}
		if err := h.firebaseService.UpdateUser(existingUser.UserID, updates); err != nil {
			log.Printf("Update user error: %v", err)
		}
	}

	jwtToken, err := h.authService.CreateJWT(userSession)
	if err != nil {
		log.Printf("JWT creation error: %v", err)
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":       "Failed to create JWT",
			"description": err.Error(),
		})
		return
	}

	// Return success page with token
	successHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Authentication Successful</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 50px auto; padding: 20px; }
        .success { background: #d4edda; border: 1px solid #c3e6cb; padding: 15px; border-radius: 5px; }
        .token { background: #f8f9fa; border: 1px solid #dee2e6; padding: 10px; margin: 10px 0; word-break: break-all; font-family: monospace; font-size: 12px; }
        button { background: #007bff; color: white; border: none; padding: 10px 15px; border-radius: 5px; cursor: pointer; margin: 5px; }
        .test-section { background: #e7f3ff; border: 1px solid #b8daff; padding: 10px; margin: 10px 0; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="success">
        <h2>✅ Authentication Successful!</h2>
        <p><strong>Welcome:</strong> %s (%s)</p>
        <p><strong>User ID:</strong> %s</p>
        
        <h3>Your JWT Token:</h3>
        <div class="token" id="token">%s</div>
        <button onclick="copyToken()">Copy Token</button>
        
        <div class="test-section">
            <h4>Quick API Test:</h4>
            <p>You can test the API immediately:</p>
            <button onclick="testMe()">Test /auth/me</button>
            <button onclick="testTasks()">Test /tasks</button>
            <div id="test-results"></div>
        </div>
        
        <h3>Manual Testing:</h3>
        <p>Use this token in your API requests:</p>
        <pre>Authorization: Bearer %s</pre>
        
        <p>Example curl command:</p>
        <pre>curl -H "Authorization: Bearer %s" \\
     %s/auth/me</pre>
    </div>
    
    <script>
        const token = '%s';
        const apiBase = '%s';
        
        function copyToken() {
            navigator.clipboard.writeText(token).then(() => {
                alert('Token copied to clipboard!');
            });
        }
        
        async function testMe() {
            try {
                const response = await fetch(apiBase + '/auth/me', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                const data = await response.json();
                document.getElementById('test-results').innerHTML = 
                    '<h5>/auth/me Result:</h5><pre>' + JSON.stringify(data, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('test-results').innerHTML = 
                    '<h5>Error:</h5><pre>' + error.message + '</pre>';
            }
        }
        
        async function testTasks() {
            try {
                const response = await fetch(apiBase + '/tasks', {
                    headers: { 'Authorization': 'Bearer ' + token }
                });
                const data = await response.json();
                document.getElementById('test-results').innerHTML = 
                    '<h5>/tasks Result:</h5><pre>' + JSON.stringify(data, null, 2) + '</pre>';
            } catch (error) {
                document.getElementById('test-results').innerHTML = 
                    '<h5>Error:</h5><pre>' + error.message + '</pre>';
            }
        }
    </script>
</body>
</html>
    `, userSession.Name, userSession.Email, userSession.UserID, jwtToken, jwtToken, jwtToken, c.Request.Host, jwtToken, c.Request.Host)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(successHTML))
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
		return
	}

	userSession := user.(*models.UserSession)
	c.JSON(http.StatusOK, gin.H{
		"id":    userSession.UserID,
		"email": userSession.Email,
		"name":  userSession.Name,
	})
}

func (h *AuthHandler) Debug(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":            "ok",
		"hasClientId":       h.googleService != nil,
		"hasFirebaseConfig": h.firebaseService != nil,
		"redirectUri":       h.googleService.GetAuthURL(),
	})
}