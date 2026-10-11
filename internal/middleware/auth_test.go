package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/arezoohaghighi/todo-api/internal/auth"
	"github.com/arezoohaghighi/todo-api/internal/model"
)

type fakeSessionRepository struct {
	receivedHash string
	session      *model.Session
}

func (r *fakeSessionRepository) Create(context.Context, *model.Session) error { return nil }
func (r *fakeSessionRepository) FindActiveByTokenHash(_ context.Context, tokenHash string) (*model.Session, error) {
	r.receivedHash = tokenHash
	return r.session, nil
}
func (r *fakeSessionRepository) Revoke(context.Context, string) error { return nil }

func TestAuthHashesBearerTokenAndSetsUser(t *testing.T) {
	token := "plain-token"
	userID := uuid.New()
	repo := &fakeSessionRepository{session: &model.Session{UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := UserIDFromContext(r.Context()); got != userID.String() {
			t.Fatalf("user id = %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	Auth(repo, next).ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	if repo.receivedHash != auth.HashToken(token) {
		t.Fatal("middleware passed an unhashed token")
	}
}

func TestAuthRejectsMissingHeaderAsJSON(t *testing.T) {
	repo := &fakeSessionRepository{}
	response := httptest.NewRecorder()
	Auth(repo, http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("next handler called") })).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", response.Code)
	}
	if got := response.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
}
