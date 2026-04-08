package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"online-judge-backend/login-flow-subset/internal/config"
	"online-judge-backend/login-flow-subset/internal/repositories"
	"online-judge-backend/login-flow-subset/internal/services"
)

func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{
		Server: config.ServerConfig{Port: "0"},
		Database: config.DatabaseConfig{
			DSN: fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name()),
		},
		Auth: config.AuthConfig{
			JWTSecret:            "test-secret",
			DefaultAdminUsername: "admin",
			DefaultAdminPassword: "admin123456",
		},
	}

	db, err := repositories.NewDatabase(cfg)
	if err != nil {
		t.Fatalf("new database: %v", err)
	}

	authService := services.NewAuthService(repositories.NewUserRepository(db), cfg.Auth.JWTSecret)
	return NewServer(authService)
}

func TestRegisterAndMeFlow(t *testing.T) {
	router := setupTestServer(t)

	resp := performRequest(t, router, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": "student_a",
		"password": "secret12",
	}, "")
	assertStatus(t, resp, http.StatusCreated)

	var payload struct {
		Token string `json:"token"`
		User  struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	decodeJSON(t, resp, &payload)
	if payload.User.Username != "student_a" || payload.Token == "" {
		t.Fatalf("unexpected register payload: %+v", payload)
	}

	meResp := performRequest(t, router, http.MethodGet, "/api/v1/auth/me", nil, payload.Token)
	assertStatus(t, meResp, http.StatusOK)

	var meBody struct {
		User struct {
			Username string `json:"username"`
		} `json:"user"`
	}
	decodeJSON(t, meResp, &meBody)
	if meBody.User.Username != "student_a" {
		t.Fatalf("unexpected me payload: %+v", meBody)
	}
}

func TestDefaultAdminCanLogin(t *testing.T) {
	router := setupTestServer(t)

	resp := performRequest(t, router, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": "admin",
		"password": "admin123456",
	}, "")
	assertStatus(t, resp, http.StatusOK)

	var payload struct {
		User struct {
			Role int `json:"role"`
		} `json:"user"`
	}
	decodeJSON(t, resp, &payload)
	if payload.User.Role != 1 {
		t.Fatalf("expected admin role, got %+v", payload)
	}
}

func TestStaticAssetsServe(t *testing.T) {
	router := setupTestServer(t)

	indexResp := performRequest(t, router, http.MethodGet, "/", nil, "")
	assertStatus(t, indexResp, http.StatusOK)

	jsResp := performRequest(t, router, http.MethodGet, "/assets/app.js", nil, "")
	assertStatus(t, jsResp, http.StatusOK)
}

func performRequest(t *testing.T, router http.Handler, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()

	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
	}

	req, err := http.NewRequest(method, path, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Fatalf("unexpected status %d, body=%s", resp.Code, resp.Body.String())
	}
}

func decodeJSON(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(resp.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, resp.Body.String())
	}
}
