// Package api tests the HTTP router as an end-to-end integration surface.
// api 包中的测试把 HTTP 路由当作端到端集成面来验证。
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"online-judge-backend/internal/config"
	"online-judge-backend/internal/judger"
	"online-judge-backend/internal/repositories"
	"online-judge-backend/internal/services"
)

// setupTestServer creates an in-memory server instance for integration tests.
// setupTestServer 为集成测试创建一个内存版服务实例。
func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	db, err := repositories.NewDatabase(config.DatabaseConfig{
		Driver: "sqlite",
		DSN:    "file::memory:?cache=shared",
	})
	if err != nil {
		t.Fatalf("new database: %v", err)
	}

	judgeService := services.NewJudgeService(judger.NewStub())
	authService := services.NewAuthService(repositories.NewUserRepository(db), "test-secret")
	return NewServer(db, judgeService, authService)
}

func TestAuthAndProtectedProblemFlow(t *testing.T) {
	router := setupTestServer(t)

	adminToken := loginAndReadToken(t, router, "admin", "admin123456")

	problemID := postJSONAndReadID(t, router, "/api/v1/admin/problems", map[string]any{
		"title":        "A+B Problem",
		"description":  "Calculate a + b",
		"time_limit":   1000,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "1 2", "output": "3", "is_sample": true},
			{"input": "3 4", "output": "7", "is_sample": false},
		},
	}, adminToken)

	resp := performRequest(t, router, http.MethodGet, "/api/v1/problems", nil, "")
	assertStatus(t, resp, http.StatusOK)

	var listBody struct {
		Items []map[string]any `json:"items"`
		Total float64          `json:"total"`
	}
	decodeJSON(t, resp, &listBody)
	if len(listBody.Items) != 1 || int(listBody.Total) != 1 {
		t.Fatalf("unexpected problem list: %+v", listBody)
	}

	detailResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/problems/%d", problemID), nil, "")
	assertStatus(t, detailResp, http.StatusOK)

	var detailBody struct {
		Title   string `json:"title"`
		Samples []any  `json:"samples"`
	}
	decodeJSON(t, detailResp, &detailBody)
	if detailBody.Title != "A+B Problem" || len(detailBody.Samples) != 1 {
		t.Fatalf("unexpected problem detail: %+v", detailBody)
	}

	userToken := registerAndReadToken(t, router, "student_a", "secret12")
	submissionID := postJSONAndReadID(t, router, "/api/v1/submissions", map[string]any{
		"problem_id": problemID,
		"language":   "cpp",
		"code":       "#include <iostream>\nint main(){return 0;}",
	}, userToken)

	submissionResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/submissions/%d", submissionID), nil, "")
	assertStatus(t, submissionResp, http.StatusOK)

	var submissionBody struct {
		Status string `json:"status"`
	}
	decodeJSON(t, submissionResp, &submissionBody)
	if submissionBody.Status != "AC" {
		t.Fatalf("expected AC, got %s", submissionBody.Status)
	}

	rankResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/problems/%d/rank", problemID), nil, "")
	assertStatus(t, rankResp, http.StatusOK)

	var rankBody struct {
		Ranks []struct {
			Rank     int    `json:"rank"`
			Username string `json:"username"`
		} `json:"ranks"`
	}
	decodeJSON(t, rankResp, &rankBody)
	if len(rankBody.Ranks) != 1 || rankBody.Ranks[0].Username != "student_a" {
		t.Fatalf("unexpected leaderboard: %+v", rankBody)
	}
}

func TestAdminRouteRequiresAdminRole(t *testing.T) {
	router := setupTestServer(t)

	userToken := registerAndReadToken(t, router, "student_b", "secret12")
	resp := performRequest(t, router, http.MethodGet, "/api/v1/admin/status", nil, userToken)
	assertStatus(t, resp, http.StatusForbidden)
}

func TestStaticAssetsAreServedWithExpectedContentType(t *testing.T) {
	router := setupTestServer(t)

	indexResp := performRequest(t, router, http.MethodGet, "/", nil, "")
	assertStatus(t, indexResp, http.StatusOK)

	cssResp := performRequest(t, router, http.MethodGet, "/assets/styles.css", nil, "")
	assertStatus(t, cssResp, http.StatusOK)
	if got := cssResp.Header().Get("Content-Type"); got != "text/css; charset=utf-8" && got != "text/css" {
		t.Fatalf("unexpected css content-type: %s", got)
	}

	jsResp := performRequest(t, router, http.MethodGet, "/assets/app.js", nil, "")
	assertStatus(t, jsResp, http.StatusOK)
}

// performRequest builds an HTTP request and sends it to the in-process router.
// performRequest 构造 HTTP 请求并发送给进程内路由。
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

// postJSONAndReadID is a helper for POST endpoints that return a numeric id.
// postJSONAndReadID 是针对返回数值型 id 的 POST 接口辅助函数。
func postJSONAndReadID(t *testing.T, router http.Handler, path string, body any, token string) uint {
	t.Helper()

	resp := performRequest(t, router, http.MethodPost, path, body, token)
	assertStatus(t, resp, http.StatusCreated)

	var payload map[string]any
	decodeJSON(t, resp, &payload)

	idValue, ok := payload["id"]
	if !ok {
		t.Fatalf("missing id in response: %s", resp.Body.String())
	}

	switch v := idValue.(type) {
	case float64:
		return uint(v)
	case string:
		id, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("parse id: %v", err)
		}
		return uint(id)
	default:
		t.Fatalf("unsupported id type %T", idValue)
		return 0
	}
}

func registerAndReadToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	resp := performRequest(t, router, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"username": username,
		"password": password,
	}, "")
	assertStatus(t, resp, http.StatusCreated)

	var payload struct {
		Token string `json:"token"`
	}
	decodeJSON(t, resp, &payload)
	return payload.Token
}

func loginAndReadToken(t *testing.T, router http.Handler, username, password string) string {
	t.Helper()
	resp := performRequest(t, router, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"username": username,
		"password": password,
	}, "")
	assertStatus(t, resp, http.StatusOK)

	var payload struct {
		Token string `json:"token"`
	}
	decodeJSON(t, resp, &payload)
	return payload.Token
}

// assertStatus keeps failure messages short but still includes the response body.
// assertStatus 让断言失败信息保持简洁，同时附带响应体便于排查。
func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if resp.Code != expected {
		t.Fatalf("unexpected status %d, body=%s", resp.Code, resp.Body.String())
	}
}

// decodeJSON unmarshals a response and fails the test with the raw body on error.
// decodeJSON 负责反序列化响应；若失败，会连同原始响应体一起报错。
func decodeJSON(t *testing.T, resp *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(resp.Body.Bytes(), target); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, resp.Body.String())
	}
}
