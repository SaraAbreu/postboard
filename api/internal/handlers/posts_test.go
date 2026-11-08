package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SaraAbreu/postboard/api/internal/models"
	"github.com/SaraAbreu/postboard/api/internal/store"
)

func newTestServer(t *testing.T) (*http.ServeMux, *store.MemoryStore) {
	t.Helper()
	s := store.NewMemoryStore()
	mux := http.NewServeMux()
	NewPostHandler(s, slog.New(slog.NewTextHandler(io.Discard, nil))).Routes(mux)
	return mux, s
}

// Table-driven test: el estilo habitual en Go.
func TestCreatePost(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{"válido", `{"title":"Hola","content":"x","platform":"instagram","status":"draft"}`, http.StatusCreated},
		{"sin título", `{"title":"","platform":"instagram","status":"draft"}`, http.StatusUnprocessableEntity},
		{"plataforma inválida", `{"title":"Hola","platform":"tiktok","status":"draft"}`, http.StatusUnprocessableEntity},
		{"scheduled sin fecha", `{"title":"Hola","platform":"linkedin","status":"scheduled"}`, http.StatusUnprocessableEntity},
		{"JSON roto", `{"title":`, http.StatusBadRequest},
		{"campo desconocido", `{"title":"Hola","platform":"instagram","status":"draft","foo":1}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux, _ := newTestServer(t)
			req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, req)
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, quería %d (body: %s)", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	mux, s := newTestServer(t)
	_, _ = s.Create(t.Context(), models.PostInput{Title: "Uno", Platform: "facebook", Status: "draft"})

	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/posts/1", http.StatusOK},
		{"/posts/99", http.StatusNotFound},
		{"/posts/abc", http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))
			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, quería %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestDeletePost(t *testing.T) {
	mux, s := newTestServer(t)
	_, _ = s.Create(t.Context(), models.PostInput{Title: "Uno", Platform: "facebook", Status: "draft"})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/posts/1", nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, quería 204", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/posts/1", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("segundo delete: status = %d, quería 404", rec.Code)
	}
}
