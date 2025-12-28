package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Notifier interface {
	NotifyAssignee(assigneeLogin, msg string) error
}

type Handler struct {
	notifier Notifier
	secret   []byte
	logger   *log.Logger
}

func NewHandler(n Notifier, githubSecret string, logger *log.Logger) *Handler {
	if logger == nil {
		logger = log.Default()
	}

	return &Handler{
		notifier: n,
		secret:   []byte(strings.TrimSpace(githubSecret)),
		logger:   logger,
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
		h.logger.Printf("[github] read body error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validateGitHubSignature(body, r.Header.Get("X-Hub-Signature-256"), h.secret); err != nil {
		if errors.Is(err, ErrMissingSignature) {
			h.logger.Printf("[github] missing signature")
		} else {
			h.logger.Printf("[github] bad signature: %v", err)
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
		h.logger.Printf("[github] pull_request unmarshal error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if payload.Action != "assigned" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if payload.Assignee == nil || payload.Assignee.Login == "" {
		h.logger.Printf("[github] assigned action without assignee")
		w.WriteHeader(http.StatusOK)
		return
	}

	msg := fmt.Sprintf("На вас назначен pull request: %s — %s", payload.PullRequest.Title, payload.PullRequest.HTMLURL)
	if err := h.notifier.NotifyAssignee(payload.Assignee.Login, msg); err != nil {

		h.logger.Printf("[github] notify assignee error: %v", err)
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
		h.logger.Printf("[github] pull_request_review unmarshal error: %v", err)
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

	type prWithAuthor struct {
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	var prData prWithAuthor
	_ = json.Unmarshal(body, &struct {
		PullRequest *prWithAuthor `json:"pull_request"`
	}{PullRequest: &prData})

	assigneeLogin := strings.ToLower(payload.PullRequest.User.Login)
	if assigneeLogin == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.notifier.NotifyAssignee(assigneeLogin, textBuilder.String()); err != nil {
		h.logger.Printf("[github] notify assignee (review) error: %v", err)
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
		h.logger.Printf("[github] pull_request_review_comment unmarshal error: %v", err)
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

	type prWithAuthor struct {
		User struct {
			Login string `json:"login"`
		} `json:"user"`
	}
	var prData prWithAuthor
	_ = json.Unmarshal(body, &struct {
		PullRequest *prWithAuthor `json:"pull_request"`
	}{PullRequest: &prData})

	assigneeLogin := strings.ToLower(payload.PullRequest.User.Login)
	if assigneeLogin == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := h.notifier.NotifyAssignee(assigneeLogin, sb.String()); err != nil {
		h.logger.Printf("[github] notify assignee (review_comment) error: %v", err)
	}

	w.WriteHeader(http.StatusOK)
}
