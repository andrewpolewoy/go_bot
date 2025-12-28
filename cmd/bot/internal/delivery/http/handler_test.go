package http

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
)

type notifierMock struct {
	calls []struct {
		login string
		msg   string
	}
	err error
}

func (n *notifierMock) NotifyAssignee(login, msg string) error {
	n.calls = append(n.calls, struct {
		login string
		msg   string
	}{login: login, msg: msg})
	return n.err
}

func sign(t *testing.T, secret string, body []byte) string {
	t.Helper()
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestGitHubWebhook_Assigned_SendsNotification(t *testing.T) {
	secret := "secret"
	n := &notifierMock{}
	h := NewHandler(n, secret, nil)

	body := []byte(`{
		"action":"assigned",
		"pull_request":{"title":"PR title","html_url":"https://example.com/pr/1"},
		"assignee":{"login":"andrewpolewoy"}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/github/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", sign(t, secret, body))

	rr := httptest.NewRecorder()
	h.GitHubWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(n.calls) != 1 {
		t.Fatalf("expected 1 notify call, got %d", len(n.calls))
	}
	if n.calls[0].login != "andrewpolewoy" {
		t.Fatalf("expected login andrewpolewoy, got %q", n.calls[0].login)
	}
}

func TestGitHubWebhook_NotAssigned_NoNotification(t *testing.T) {
	secret := "secret"
	n := &notifierMock{}
	h := NewHandler(n, secret, nil)

	body := []byte(`{
		"action":"opened",
		"pull_request":{"title":"PR title","html_url":"https://example.com/pr/1"},
		"assignee":{"login":"andrewpolewoy"}
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/github/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "pull_request")
	req.Header.Set("X-Hub-Signature-256", sign(t, secret, body))

	rr := httptest.NewRecorder()
	h.GitHubWebhook(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if len(n.calls) != 0 {
		t.Fatalf("expected 0 notify calls, got %d", len(n.calls))
	}
}
