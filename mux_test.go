package chi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareRxCloneURLParams(t *testing.T) {
	r := NewRouter()
	r.Route("/users/{userID}", func(r Router) {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
				ctx := context.WithValue(req.Context(), "customKey", "customValue")
				next.ServeHTTP(w, req.WithContext(ctx))
			})
		})
		r.Get("/profile", func(w http.ResponseWriter, req *http.Request) {
			userID := URLParam(req, "userID")
			if userID != "123" {
				t.Errorf("expected userID to be '123', got '%s'", userID)
			}
			if RouteContext(req.Context()) == nil {
				t.Error("expected active route context after request cloning")
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		})
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	res, err := http.Get(ts.URL + "/users/123/profile")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected status OK, got %d", res.StatusCode)
	}
}

func TestNestedMiddlewareClonePreservesURLParams(t *testing.T) {
	r := NewRouter()
	r.Route("/orgs/{orgID}", func(r Router) {
		r.Route("/users/{userID}", func(r Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					ctx := context.WithValue(req.Context(), "traceID", "abc")
					next.ServeHTTP(w, req.WithContext(ctx))
				})
			})
			r.Get("/profile", func(w http.ResponseWriter, req *http.Request) {
				if got := URLParam(req, "orgID"); got != "acme" {
					t.Fatalf("expected orgID acme, got %q", got)
				}
				if got := URLParam(req, "userID"); got != "123" {
					t.Fatalf("expected userID 123, got %q", got)
				}
				w.WriteHeader(http.StatusOK)
			})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/orgs/acme/users/123/profile", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
