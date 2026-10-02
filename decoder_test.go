package render

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	t.Run("valid json object", func(t *testing.T) {
		var v struct {
			Title string `json:"title"`
		}
		err := DecodeJSON(strings.NewReader(`{"title": "hello"}`), &v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v.Title != "hello" {
			t.Fatalf("expected hello, got %s", v.Title)
		}
	})

	t.Run("valid json with trailing whitespace", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader("{\"title\": \"hello\"} \n\t\r\n "), &v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v["title"] != "hello" {
			t.Fatalf("expected hello, got %v", v["title"])
		}
	})

	t.Run("rejects trailing garbage data", func(t *testing.T) {
		var v map[string]interface{}
		// Regression test for issue #42
		err := DecodeJSON(strings.NewReader("{}\n some garbage data"), &v)
		if err == nil {
			t.Fatal("expected error when trailing garbage data follows valid JSON, got nil")
		}
	})

	t.Run("rejects trailing second json object", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader(`{"a": 1} {"b": 2}`), &v)
		if err == nil {
			t.Fatal("expected error when second JSON object follows, got nil")
		}
	})

	t.Run("rejects trailing primitive", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader(`{"a": 1} 123`), &v)
		if err == nil {
			t.Fatal("expected error when trailing number follows JSON, got nil")
		}
	})

	t.Run("rejects trailing delimiter", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader(`{"a": 1} ]`), &v)
		if err == nil {
			t.Fatal("expected error when trailing delimiter follows JSON, got nil")
		}

		err = DecodeJSON(strings.NewReader(`{"a": 1} }`), &v)
		if err == nil {
			t.Fatal("expected error when trailing delimiter follows JSON, got nil")
		}
	})

	t.Run("rejects trailing null", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader(`{"a": 1} null`), &v)
		if err == nil {
			t.Fatal("expected error when trailing null follows JSON, got nil")
		}
	})

	t.Run("invalid json returns error", func(t *testing.T) {
		var v map[string]interface{}
		err := DecodeJSON(strings.NewReader(`{invalid`), &v)
		if err == nil {
			t.Fatal("expected error on invalid JSON, got nil")
		}
	})
}

func TestDefaultDecoder(t *testing.T) {
	t.Run("json with trailing garbage returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{}\n some garbage data"))
		req.Header.Set("Content-Type", "application/json")

		var v map[string]interface{}
		err := DefaultDecoder(req, &v)
		if err == nil {
			t.Fatal("expected error for request with trailing garbage, got nil")
		}
	})

	t.Run("valid json succeeds", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"status": "ok"}`))
		req.Header.Set("Content-Type", "application/json")

		var v map[string]string
		err := DefaultDecoder(req, &v)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v["status"] != "ok" {
			t.Fatalf("expected ok, got %v", v["status"])
		}
	})

	t.Run("unsupported content type returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("plain text"))
		req.Header.Set("Content-Type", "text/plain")

		var v map[string]interface{}
		err := DefaultDecoder(req, &v)
		if err == nil {
			t.Fatal("expected error for unsupported content type, got nil")
		}
	})
}
