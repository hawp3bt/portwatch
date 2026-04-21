package notify

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_WebhookAndDesktopFlags(t *testing.T) {
	n := New("http://example.com/hook", false)
	if n.webhookURL != "http://example.com/hook" {
		t.Errorf("expected webhookURL to be set")
	}
	if n.desktop {
		t.Errorf("expected desktop to be false")
	}
}

func TestSendWebhook_Success(t *testing.T) {
	var received webhookPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	if err := sendWebhook(ts.URL, "Port Opened", "portwatch: port 8080 opened"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Title != "Port Opened" {
		t.Errorf("expected title 'Port Opened', got %q", received.Title)
	}
	if received.Body != "portwatch: port 8080 opened" {
		t.Errorf("unexpected body: %q", received.Body)
	}
}

func TestSendWebhook_Non2xxReturnsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	if err := sendWebhook(ts.URL, "title", "body"); err == nil {
		t.Error("expected error for non-2xx response")
	}
}

func TestNotifier_Opened_CallsWebhook(t *testing.T) {
	var received webhookPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	n := New(ts.URL, false)
	if err := n.Opened(9090); err != nil {
		t.Fatalf("Opened returned error: %v", err)
	}
	if received.Title != "Port Opened" {
		t.Errorf("unexpected title: %q", received.Title)
	}
}

func TestNotifier_Closed_CallsWebhook(t *testing.T) {
	var received webhookPayload

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := New(ts.URL, false)
	if err := n.Closed(443); err != nil {
		t.Fatalf("Closed returned error: %v", err)
	}
	if received.Title != "Port Closed" {
		t.Errorf("unexpected title: %q", received.Title)
	}
}
