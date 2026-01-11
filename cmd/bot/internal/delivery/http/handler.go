package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/logger"
)

type Notifier interface {
	NotifyAssignee(ctx context.Context, assigneeLogin, msg string) error
}

type Handler struct {
	notifier Notifier
	secret   []byte
	logger   logger.Logger
}

func NewHandler(n Notifier, githubSecret string, log logger.Logger) *Handler {
	if log == nil {
		log = logger.NewDefault()
	}

	return &Handler{
		notifier: n,
		secret:   []byte(strings.TrimSpace(githubSecret)),
		logger:   log,
	}
}

func (h *Handler) GitHubWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	defer func() { _ = r.Body.Close() }()
	body, err := io.ReadAll(r.Body)

	if err != nil {
		h.logger.Error("github read body error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validateGitHubSignature(body, r.Header.Get("X-Hub-Signature-256"), h.secret); err != nil {
		if errors.Is(err, ErrMissingSignature) {
			h.logger.Error("github missing signature")
		} else {
			h.logger.Error("github bad signature", "err", err)
		}
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	event := r.Header.Get("X-GitHub-Event")

	switch event {
	case "pull_request":
		h.handlePullRequest(w, body)
	case "pull_request_review":
		h.handlePullRequestReview(w, body)
	case "pull_request_review_comment":
		h.handlePullRequestReviewComment(w, body)
	default:
		w.WriteHeader(http.StatusOK)
	}
}

type pullRequestPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
	Assignee *struct {
		Login string `json:"login"`
	} `json:"assignee"`
}

func (h *Handler) handlePullRequest(w http.ResponseWriter, body []byte) {
	var payload pullRequestPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("github pull_request unmarshal error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "assigned" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Assignee == nil || payload.Assignee.Login == "" {
		h.logger.Error("github assigned action without assignee")
		w.WriteHeader(http.StatusOK)
		return
	}

	msg := fmt.Sprintf("На вас назначен pull request: %s — %s", payload.PullRequest.Title, payload.PullRequest.HTMLURL)
	ctx := context.Background()
	if err := h.notifier.NotifyAssignee(ctx, payload.Assignee.Login, msg); err != nil {
		h.logger.Error("github notify assignee error", "err", err, "login", payload.Assignee.Login)
	}

	w.WriteHeader(http.StatusOK)
}

type pullRequestReviewPayload struct {
	Action string `json:"action"`
	Review struct {
		State string `json:"state"`
		Body  string `json:"body"`
	} `json:"review"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`

	Reviewer *struct {
		Login string `json:"login"`
	} `json:"sender"`
}

func (h *Handler) handlePullRequestReview(w http.ResponseWriter, body []byte) {
	var payload pullRequestReviewPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("github pull_request_review unmarshal error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "submitted" {
		w.WriteHeader(http.StatusOK)
		return
	}

	state := strings.ToLower(payload.Review.State)
	if state == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	reviewText := trimText(payload.Review.Body, 400)

	var msgPrefix string
	switch state {
	case "approved":
		msgPrefix = "Ваш PR одобрен"
	case "changes_requested":
		msgPrefix = "По вашему PR запрошены изменения"
	case "commented":
		msgPrefix = "Новый review по вашему PR"
	default:
		w.WriteHeader(http.StatusOK)
		return
	}

	title := payload.PullRequest.Title
	url := payload.PullRequest.HTMLURL

	var textBuilder strings.Builder
	textBuilder.WriteString(msgPrefix)
	textBuilder.WriteString(": ")
	textBuilder.WriteString(title)
	textBuilder.WriteString(" — ")
	textBuilder.WriteString(url)

	if reviewText != "" {
		textBuilder.WriteString("\n\nReview: ")
		textBuilder.WriteString(reviewText)
	}

	// Важно: нам нужен логин assignee, а не автора review.
	// GitHub в этом event не несёт assignee явно, поэтому:
	// 1) либо считаем, что PR-ревью всегда автору PR (и смотрим author.login),
	// 2) либо оставляем на потом расширение модели.
	// Для простоты задания можно считать, что "assignee == author" и добавить поле.
	// Если в твоей текущей модели уже есть assignee в PullRequest — используй его.
	//
	// Ниже — вариант с author.login, если ты решишь маппить на него.

	assigneeLogin := strings.ToLower(payload.PullRequest.User.Login)
	if assigneeLogin == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()
	if err := h.notifier.NotifyAssignee(ctx, assigneeLogin, textBuilder.String()); err != nil {
		h.logger.Error("github notify assignee (review) error", "err", err, "login", assigneeLogin)
	}

	w.WriteHeader(http.StatusOK)
}

type pullRequestReviewCommentPayload struct {
	Action  string `json:"action"`
	Comment struct {
		Body string `json:"body"`
	} `json:"comment"`
	PullRequest struct {
		Title   string `json:"title"`
		HTMLURL string `json:"html_url"`
		User    struct {
			Login string `json:"login"`
		} `json:"user"`
	} `json:"pull_request"`
}

func (h *Handler) handlePullRequestReviewComment(w http.ResponseWriter, body []byte) {
	var payload pullRequestReviewCommentPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		h.logger.Error("github pull_request_review_comment unmarshal error", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "created" {
		w.WriteHeader(http.StatusOK)
		return
	}

	commentText := trimText(payload.Comment.Body, 400)
	title := payload.PullRequest.Title
	url := payload.PullRequest.HTMLURL

	var sb strings.Builder
	sb.WriteString("Новый комментарий к вашему PR: ")
	sb.WriteString(title)
	sb.WriteString(" — ")
	sb.WriteString(url)

	if commentText != "" {
		sb.WriteString("\n\nКомментарий: ")
		sb.WriteString(commentText)
	}

	assigneeLogin := strings.ToLower(payload.PullRequest.User.Login)
	if assigneeLogin == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	ctx := context.Background()
	if err := h.notifier.NotifyAssignee(ctx, assigneeLogin, sb.String()); err != nil {
		h.logger.Error("github notify assignee (review_comment) error", "err", err, "login", assigneeLogin)
	}

	w.WriteHeader(http.StatusOK)
}
