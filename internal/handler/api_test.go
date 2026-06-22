package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"room-api/internal/auth"
	"room-api/internal/middleware"
	"room-api/internal/realtime"
)

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestAuthAPIInvalidRequests(t *testing.T) {
	router := testRouter()
	userHandler := NewUserHandler(nil, nil)

	authRoutes := router.Group("/api/v1/auth")
	{
		authRoutes.POST("/register", userHandler.Register)
		authRoutes.POST("/login", userHandler.Login)
		authRoutes.POST("/reset-password", userHandler.ResetPassword)
	}

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		wantCode    int
		wantMessage string
	}{
		{
			name:        "register rejects invalid json",
			method:      http.MethodPost,
			path:        "/api/v1/auth/register",
			body:        "{",
			wantCode:    500,
			wantMessage: "参数错误",
		},
		{
			name:        "register requires email code",
			method:      http.MethodPost,
			path:        "/api/v1/auth/register",
			body:        `{"account":"alex_001","email":"alex@example.com","password":"123456","nickname":"Alex","avatar_url":"https://example.com/avatar.png"}`,
			wantCode:    500,
			wantMessage: "验证码错误",
		},
		{
			name:        "login rejects invalid json",
			method:      http.MethodPost,
			path:        "/api/v1/auth/login",
			body:        "{",
			wantCode:    500,
			wantMessage: "参数错误",
		},
		{
			name:        "reset password rejects invalid json",
			method:      http.MethodPost,
			path:        "/api/v1/auth/reset-password",
			body:        "{",
			wantCode:    500,
			wantMessage: "参数错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performJSONRequest(router, tt.method, tt.path, tt.body, nil)
			assertAPIResponse(t, response, tt.wantCode, tt.wantMessage)
		})
	}
}

func TestProtectedAPIsRequireAuthorization(t *testing.T) {
	router := testRouter()
	jwtSvc := auth.NewService("test-secret")
	userHandler := NewUserHandler(nil, nil)

	users := router.Group("/api/v1/users", middleware.Auth(jwtSvc))
	{
		users.GET("/me", userHandler.Me)
	}

	tests := []struct {
		name   string
		header string
	}{
		{name: "missing header"},
		{name: "malformed header", header: "Token abc"},
		{name: "invalid token", header: "Bearer invalid-token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{}
			if tt.header != "" {
				headers["Authorization"] = tt.header
			}

			response := performJSONRequest(router, http.MethodGet, "/api/v1/users/me", "", headers)
			assertAPIResponse(t, response, 401, "未登录")
		})
	}
}

func TestRoomAndMessageAPIInvalidRequests(t *testing.T) {
	router := testRouter()
	hub := realtime.NewHub()
	roomHandler := NewRoomHandler(nil, hub)
	messageHandler := NewMessageHandler(nil, hub)

	authed := router.Group("/api/v1", withTestUser(1))
	{
		authed.POST("/rooms", roomHandler.Create)
		authed.POST("/rooms/:room_id/messages", messageHandler.Create)
	}

	tests := []struct {
		name        string
		method      string
		path        string
		body        string
		wantCode    int
		wantMessage string
	}{
		{
			name:        "create room rejects invalid json",
			method:      http.MethodPost,
			path:        "/api/v1/rooms",
			body:        "{",
			wantCode:    500,
			wantMessage: "参数错误",
		},
		{
			name:        "message rejects invalid room id",
			method:      http.MethodPost,
			path:        "/api/v1/rooms/not-number/messages",
			body:        `{"content":"hello"}`,
			wantCode:    500,
			wantMessage: "参数错误",
		},
		{
			name:        "message rejects invalid json",
			method:      http.MethodPost,
			path:        "/api/v1/rooms/1/messages",
			body:        "{",
			wantCode:    500,
			wantMessage: "参数错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := performJSONRequest(router, tt.method, tt.path, tt.body, nil)
			assertAPIResponse(t, response, tt.wantCode, tt.wantMessage)
		})
	}
}

func testRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func withTestUser(userID uint64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func performJSONRequest(router http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertAPIResponse(t *testing.T, response *httptest.ResponseRecorder, wantCode int, wantMessage string) {
	t.Helper()

	if response.Code != http.StatusOK {
		t.Fatalf("HTTP status = %d, want %d, body = %s", response.Code, http.StatusOK, response.Body.String())
	}

	var body apiResponse
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("response body is not valid JSON: %v, body = %s", err, response.Body.String())
	}

	if body.Code != wantCode || body.Message != wantMessage {
		t.Fatalf("response = {code:%d message:%q}, want {code:%d message:%q}", body.Code, body.Message, wantCode, wantMessage)
	}

	if body.Data == nil {
		t.Fatal("response data field is missing")
	}
}
