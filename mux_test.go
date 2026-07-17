package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareRxCloneURLParams(t *testing.T) {
	r := NewRouter()

	r.Get("/users/{userID}/profile", func(w http.ResponseWriter, req *http.Request) {
		userID := URLParam(req, "userID")
		if userID != "123" {
			t.Errorf("expected userID to be '123', got '%s'", userID)
		}
		w.Write([]byte("ok"))
	})

	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ctx := context.WithValue(req.Context(), "customKey", "customValue")
			next.ServeHTTP(w, req.WithContext(ctx))
		})
	}
	handler := mw(r)

	ts := httptest.NewServer(handler)
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
