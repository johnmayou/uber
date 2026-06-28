package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		path string
		want []string
	}{
		{"/", []string{"trip:/", "add(1, 1):2"}},
		{"/foo", []string{"trip:/foo", "add(1, 1):2"}},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, tt.path, nil)
		w := httptest.NewRecorder()
		handler(w, r)

		body := w.Body.String()
		for _, s := range tt.want {
			if !strings.Contains(body, s) {
				t.Errorf("path %s: want %q in body, got %q", tt.path, s, body)
			}
		}
	}
}
