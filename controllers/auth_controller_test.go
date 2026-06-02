package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"expense-tracker-api/controllers"
	"expense-tracker-api/models"

	beegoContext "github.com/beego/beego/v2/server/web/context"
)

type apiResponse struct {
	Success bool                   `json:"success"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func newTempUserCSV(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "users.csv")
	if err := os.WriteFile(path, []byte("id,name,email,password,created_at\n"), 0644); err != nil {
		t.Fatalf("failed to create temp user CSV: %v", err)
	}
	return path, func() {
		models.SetUsersCSVPath("")
		os.Remove(path)
	}
}

func buildControllerContext(t *testing.T, method, url string, body []byte, headers map[string]string) (*beegoContext.Context, *httptest.ResponseRecorder) {
	t.Helper()
	req := httptest.NewRequest(method, url, bytes.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	ctx := beegoContext.NewContext()
	ctx.Reset(rec, req)
	ctx.Input.RequestBody = body
	return ctx, rec
}

func setPathParam(ctx *beegoContext.Context, key, value string) {
	// For testing without the router, we store path params in request form
	if ctx.Request.Form == nil {
		ctx.Request.Form = make(url.Values)
	}
	ctx.Request.Form.Set(key, value)
}

func executeControllerAction(t *testing.T, action func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && err.Error() == "user stop run" {
				return
			}
			panic(r)
		}
	}()
	action()
}

func TestMain(m *testing.M) {
	cwd, err := os.Getwd()
	if err == nil {
		_ = os.Chdir(filepath.Join(cwd, ".."))
	}
	os.Exit(m.Run())
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		seedData    string
		wantStatus  int
		wantMessage string
	}{
		{
			name:        "valid registration",
			body:        `{"name":"John Doe","email":"john@example.com","password":"secret123"}`,
			wantStatus:  201,
			wantMessage: "User registered successfully",
		},
		{
			name:        "missing email",
			body:        `{"name":"John Doe","password":"secret123"}`,
			wantStatus:  400,
			wantMessage: "Email is required",
		},
		{
			name:        "duplicate email",
			seedData:    "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n",
			body:        `{"name":"Jane Doe","email":"john@example.com","password":"secret123"}`,
			wantStatus:  409,
			wantMessage: "Email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			models.SetUsersCSVPath(path)

			if tt.seedData != "" {
				if err := os.WriteFile(path, []byte("id,name,email,password,created_at\n"+tt.seedData), 0644); err != nil {
					t.Fatalf("failed to seed user CSV: %v", err)
				}
			}

			ctx, rec := buildControllerContext(t, http.MethodPost, "/api/v1/auth/register", []byte(tt.body), map[string]string{
				"Content-Type": "application/json",
			})
			controller := controllers.AuthController{}
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})
			controller.Prepare()
			executeControllerAction(t, controller.Register)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp apiResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, resp.Message)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name        string
		seedData    string
		body        string
		wantStatus  int
		wantMessage string
		wantUserID  float64
	}{
		{
			name:        "valid login",
			seedData:    "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n",
			body:        `{"email":"john@example.com","password":"secret123"}`,
			wantStatus:  200,
			wantMessage: "Login successful",
			wantUserID:  1,
		},
		{
			name:        "wrong password",
			seedData:    "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n",
			body:        `{"email":"john@example.com","password":"wrongpass"}`,
			wantStatus:  401,
			wantMessage: "Invalid email or password",
		},
		{
			name:        "missing password",
			seedData:    "1,John Doe,john@example.com,secret123,2025-06-01T00:00:00Z\n",
			body:        `{"email":"john@example.com"}`,
			wantStatus:  400,
			wantMessage: "Email and password are required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, cleanup := newTempUserCSV(t)
			defer cleanup()
			models.SetUsersCSVPath(path)

			if tt.seedData != "" {
				if err := os.WriteFile(path, []byte("id,name,email,password,created_at\n"+tt.seedData), 0644); err != nil {
					t.Fatalf("failed to seed user CSV: %v", err)
				}
			}

			ctx, rec := buildControllerContext(t, http.MethodPost, "/api/v1/auth/login", []byte(tt.body), map[string]string{
				"Content-Type": "application/json",
			})
			controller := controllers.AuthController{}
			controller.Ctx = ctx
			controller.Data = make(map[interface{}]interface{})
			controller.Prepare()
			executeControllerAction(t, controller.Login)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}

			var resp apiResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse response: %v", err)
			}

			if resp.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, resp.Message)
			}

			if tt.wantStatus == 200 {
				if resp.Data == nil {
					t.Fatal("expected data to be returned")
				}
				userID, ok := resp.Data["user_id"].(float64)
				if !ok || userID != tt.wantUserID {
					t.Fatalf("expected user_id %v, got %v", tt.wantUserID, resp.Data["user_id"])
				}
			}
		})
	}
}
