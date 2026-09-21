package order

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerHealth(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(testService()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
	if recorder.Code != http.StatusOK {
		t.Fatalf("code = %d", recorder.Code)
	}
}

func TestHandlerGetOrder(t *testing.T) {
	recorder := httptest.NewRecorder()
	NewHandler(testService()).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders/1", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), "shipped") {
		t.Fatalf("code = %d, body = %s", recorder.Code, recorder.Body.String())
	}
}

func TestHandlerRejectsInvalidRequests(t *testing.T) {
	cases := []struct {
		name string
		url  string
		body string
	}{
		{name: "invalid id", url: "/orders/x", body: ""},
		{name: "malformed json", url: "/orders/1/status", body: "{"},
		{name: "missing status", url: "/orders/1/status", body: "{}"},
		{name: "unknown status", url: "/orders/1/status", body: `{"status":"unknown"}`},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			method := http.MethodGet
			if tt.body != "" {
				method = http.MethodPut
			}
			recorder := httptest.NewRecorder()
			NewHandler(testService()).ServeHTTP(recorder, httptest.NewRequest(method, tt.url, strings.NewReader(tt.body)))
			if recorder.Code != http.StatusBadRequest {
				t.Fatalf("code = %d", recorder.Code)
			}
		})
	}
}

func TestHandlerRejectsTrailingJSONWithoutUpdatingOrder(t *testing.T) {
	repository := &fakeRepository{}
	service := NewService(repository, ReadModeFallback, WriteModeDual)
	request := httptest.NewRequest(http.MethodPut, "/orders/1/status", strings.NewReader(`{"status":"shipped"}{"status":"canceled"}`))
	recorder := httptest.NewRecorder()

	NewHandler(service).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("code = %d", recorder.Code)
	}
	if repository.updated {
		t.Fatal("repository must not be updated")
	}
}

func testService() *Service {
	legacy := 2
	return NewService(&fakeRepository{order: Order{ID: 1, StatusCode: &legacy}}, ReadModeFallback, WriteModeDual)
}
