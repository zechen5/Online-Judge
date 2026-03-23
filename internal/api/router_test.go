// Package api tests the HTTP router as an end-to-end integration surface.
// api 包中的测试把 HTTP 路由当作端到端集成面来验证。
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"online-judge-backend/internal/config"
	"online-judge-backend/internal/judger"
	"online-judge-backend/internal/models"
	"online-judge-backend/internal/repositories"
	"online-judge-backend/internal/services"
)

// setupTestServer creates an in-memory server instance for integration tests.
// setupTestServer 为集成测试创建一个内存版服务实例。
func setupTestServer(t *testing.T) http.Handler {
	t.Helper()

	db, err := repositories.NewDatabase(config.DatabaseConfig{
		Driver: "sqlite",
		DSN:    fmt.Sprintf("file:%s?mode=memory&cache=shared", url.QueryEscape(t.Name())),
	})
	if err != nil {
		t.Fatalf("new database: %v", err)
	}

	problemRepo := repositories.NewProblemRepository(db)
	submissionRepo := repositories.NewSubmissionRepository(db)
	judgeService := services.NewJudgeService(judger.NewStub())
	judgeProcessor := services.NewSubmissionJudgeProcessor(submissionRepo, problemRepo, judgeService)
	judgeQueue := judger.NewAsyncQueue(judgeProcessor, 1, 8)
	authService := services.NewAuthService(repositories.NewUserRepository(db), "test-secret")
	return NewServer(db, judgeQueue, "async-test", authService)
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
	}, adminToken, http.StatusCreated)

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
	}, userToken, http.StatusAccepted)

	submissionBody := waitForSubmissionStatus(t, router, submissionID, models.SubmissionStatusAC)
	if submissionBody.Status != models.SubmissionStatusAC {
		t.Fatalf("expected AC, got %s", submissionBody.Status)
	}

	listResp := performRequest(t, router, http.MethodGet, "/api/v1/submissions", nil, userToken)
	assertStatus(t, listResp, http.StatusOK)

	var submissionList struct {
		Items []struct {
			ID           uint   `json:"id"`
			ProblemTitle string `json:"problem_title"`
			Status       string `json:"status"`
		} `json:"items"`
	}
	decodeJSON(t, listResp, &submissionList)
	if len(submissionList.Items) != 1 || submissionList.Items[0].ID != submissionID || submissionList.Items[0].ProblemTitle != "A+B Problem" {
		t.Fatalf("unexpected submission list: %+v", submissionList)
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

func TestAdminProblemReviewUpdateVisibilityAndDeleteFlow(t *testing.T) {
	router := setupTestServer(t)

	adminToken := loginAndReadToken(t, router, "admin", "admin123456")
	userToken := registerAndReadToken(t, router, "student_reviewer", "secret12")

	problemID := postJSONAndReadID(t, router, "/api/v1/problems", map[string]any{
		"title":        "Queue Review Problem",
		"description":  "Pending before review",
		"time_limit":   1200,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "5 6", "output": "11", "is_sample": true},
		},
	}, userToken, http.StatusCreated)

	adminListResp := performRequest(t, router, http.MethodGet, "/api/v1/admin/problems", nil, adminToken)
	assertStatus(t, adminListResp, http.StatusOK)

	var adminList struct {
		Items []struct {
			ID     uint   `json:"id"`
			Title  string `json:"title"`
			Status int    `json:"status"`
		} `json:"items"`
	}
	decodeJSON(t, adminListResp, &adminList)
	if len(adminList.Items) == 0 || adminList.Items[0].ID != problemID || adminList.Items[0].Status != 0 {
		t.Fatalf("unexpected admin problem list: %+v", adminList)
	}

	adminDetailResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/admin/problems/%d", problemID), nil, adminToken)
	assertStatus(t, adminDetailResp, http.StatusOK)

	var adminDetail struct {
		Title     string `json:"title"`
		TestCases []struct {
			Input string `json:"input"`
		} `json:"test_cases"`
	}
	decodeJSON(t, adminDetailResp, &adminDetail)
	if adminDetail.Title != "Queue Review Problem" || len(adminDetail.TestCases) != 1 {
		t.Fatalf("unexpected admin problem detail: %+v", adminDetail)
	}

	publishResp := performRequest(t, router, http.MethodPatch, fmt.Sprintf("/api/v1/admin/problems/%d/status", problemID), map[string]any{
		"status": 1,
	}, adminToken)
	assertStatus(t, publishResp, http.StatusOK)

	publicDetailResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/problems/%d", problemID), nil, "")
	assertStatus(t, publicDetailResp, http.StatusOK)

	updateResp := performRequest(t, router, http.MethodPut, fmt.Sprintf("/api/v1/admin/problems/%d", problemID), map[string]any{
		"title":        "Queue Review Problem Updated",
		"description":  "Reviewed and edited",
		"time_limit":   1500,
		"memory_limit": 384,
		"status":       2,
		"test_cases": []map[string]any{
			{"input": "8 9", "output": "17", "is_sample": true},
		},
	}, adminToken)
	assertStatus(t, updateResp, http.StatusOK)

	publicHiddenResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/problems/%d", problemID), nil, "")
	assertStatus(t, publicHiddenResp, http.StatusNotFound)

	listResp := performRequest(t, router, http.MethodGet, "/api/v1/problems", nil, "")
	assertStatus(t, listResp, http.StatusOK)
	var publicList struct {
		Items []map[string]any `json:"items"`
	}
	decodeJSON(t, listResp, &publicList)
	if len(publicList.Items) != 0 {
		t.Fatalf("expected hidden problem to be absent from public list: %+v", publicList)
	}

	deleteResp := performRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/problems/%d", problemID), nil, adminToken)
	assertStatus(t, deleteResp, http.StatusNoContent)

	missingAdminResp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/admin/problems/%d", problemID), nil, adminToken)
	assertStatus(t, missingAdminResp, http.StatusNotFound)
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

func TestDeleteProblemShiftsFollowingIDsForward(t *testing.T) {
	router := setupTestServer(t)
	adminToken := loginAndReadToken(t, router, "admin", "admin123456")
	userToken := registerAndReadToken(t, router, "student_shift", "secret12")

	firstID := postJSONAndReadID(t, router, "/api/v1/admin/problems", map[string]any{
		"title":        "First Problem",
		"description":  "first",
		"time_limit":   1000,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "1", "output": "1", "is_sample": true},
		},
	}, adminToken, http.StatusCreated)

	secondID := postJSONAndReadID(t, router, "/api/v1/admin/problems", map[string]any{
		"title":        "Second Problem",
		"description":  "second",
		"time_limit":   1000,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "2", "output": "2", "is_sample": true},
		},
	}, adminToken, http.StatusCreated)

	thirdID := postJSONAndReadID(t, router, "/api/v1/admin/problems", map[string]any{
		"title":        "Third Problem",
		"description":  "third",
		"time_limit":   1000,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "3", "output": "3", "is_sample": true},
		},
	}, adminToken, http.StatusCreated)

	if firstID != 1 || secondID != 2 || thirdID != 3 {
		t.Fatalf("unexpected initial ids: %d %d %d", firstID, secondID, thirdID)
	}

	submissionID := postJSONAndReadID(t, router, "/api/v1/submissions", map[string]any{
		"problem_id": thirdID,
		"language":   "python",
		"code":       "print(input().strip())",
	}, userToken, http.StatusAccepted)
	waitForSubmissionStatus(t, router, submissionID, models.SubmissionStatusAC)

	deleteResp := performRequest(t, router, http.MethodDelete, fmt.Sprintf("/api/v1/admin/problems/%d", secondID), nil, adminToken)
	assertStatus(t, deleteResp, http.StatusNoContent)

	shiftedResp := performRequest(t, router, http.MethodGet, "/api/v1/problems/2", nil, "")
	assertStatus(t, shiftedResp, http.StatusOK)

	var shiftedProblem struct {
		ID    uint   `json:"id"`
		Title string `json:"title"`
	}
	decodeJSON(t, shiftedResp, &shiftedProblem)
	if shiftedProblem.ID != 2 || shiftedProblem.Title != "Third Problem" {
		t.Fatalf("unexpected shifted problem: %+v", shiftedProblem)
	}

	oldResp := performRequest(t, router, http.MethodGet, "/api/v1/problems/3", nil, "")
	assertStatus(t, oldResp, http.StatusNotFound)

	listResp := performRequest(t, router, http.MethodGet, "/api/v1/submissions", nil, userToken)
	assertStatus(t, listResp, http.StatusOK)

	var submissionList struct {
		Items []struct {
			ProblemID    uint   `json:"problem_id"`
			ProblemTitle string `json:"problem_title"`
		} `json:"items"`
	}
	decodeJSON(t, listResp, &submissionList)
	if len(submissionList.Items) != 1 || submissionList.Items[0].ProblemID != 2 || submissionList.Items[0].ProblemTitle != "Third Problem" {
		t.Fatalf("unexpected shifted submission relation: %+v", submissionList)
	}

	newID := postJSONAndReadID(t, router, "/api/v1/admin/problems", map[string]any{
		"title":        "Fourth Problem",
		"description":  "fourth",
		"time_limit":   1000,
		"memory_limit": 256,
		"test_cases": []map[string]any{
			{"input": "4", "output": "4", "is_sample": true},
		},
	}, adminToken, http.StatusCreated)
	if newID != 3 {
		t.Fatalf("expected reseeded id 3, got %d", newID)
	}
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
func postJSONAndReadID(t *testing.T, router http.Handler, path string, body any, token string, expectedStatus int) uint {
	t.Helper()

	resp := performRequest(t, router, http.MethodPost, path, body, token)
	assertStatus(t, resp, expectedStatus)

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

func waitForSubmissionStatus(t *testing.T, router http.Handler, submissionID uint, expected string) struct {
	Status string `json:"status"`
} {
	t.Helper()

	var payload struct {
		Status string `json:"status"`
	}
	for attempt := 0; attempt < 20; attempt++ {
		resp := performRequest(t, router, http.MethodGet, fmt.Sprintf("/api/v1/submissions/%d", submissionID), nil, "")
		assertStatus(t, resp, http.StatusOK)
		decodeJSON(t, resp, &payload)
		if payload.Status == expected {
			return payload
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("submission %d did not reach status %s", submissionID, expected)
	return payload
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
