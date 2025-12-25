package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	stdhttp "net/http"
	"strings"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/service"
)

type Handler struct {
	svc    *service.Notifier
	secret []byte
}

func NewHandler(svc *service.Notifier, githubSecret string) *Handler {
	return &Handler{svc: svc, secret: []byte(githubSecret)}
}

func (h *Handler) GitHubWebhook(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPost {
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if len(h.secret) > 0 {
		if err := validateGitHubSignature(body, r.Header.Get("X-Hub-Signature-256"), h.secret); err != nil {
			w.WriteHeader(stdhttp.StatusUnauthorized)
			return
		}
	}

	event := r.Header.Get("X-GitHub-Event")
	if event != "pull_request" {
		w.WriteHeader(stdhttp.StatusOK)
		return
	}

	var payload struct {
		Action      string `json:"action"`
		PullRequest struct {
			Title string `json:"title"`
			HTML  string `json:"html_url"`
		} `json:"pull_request"`
		Assignee *struct {
			Login string `json:"login"`
		} `json:"assignee"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(stdhttp.StatusBadRequest)
		return
	}

	if payload.Action != "assigned" || payload.Assignee == nil || payload.Assignee.Login == "" {
		w.WriteHeader(stdhttp.StatusOK)
		return
	}

	_ = h.svc.NotifyAssignee(context.Background(), payload.Assignee.Login, payload.PullRequest.Title, payload.PullRequest.HTML)
	w.WriteHeader(stdhttp.StatusOK)
}

func validateGitHubSignature(body []byte, sigHeader string, secret []byte) error {
	// expected: "sha256=" + hex
	const prefix = "sha256="
	if !strings.HasPrefix(sigHeader, prefix) {
		return errors.New("missing sha256 signature")
	}
	gotHex := strings.TrimPrefix(sigHeader, prefix)
	got, err := hex.DecodeString(gotHex)
	if err != nil {
		return errors.New("bad signature hex")
	}

	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(body)
	expected := mac.Sum(nil)

	if !hmac.Equal(got, expected) {
		return errors.New("signature mismatch")
	}
	return nil
}
