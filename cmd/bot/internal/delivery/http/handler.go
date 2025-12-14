package httpdelivery

import "net/http"

type GitHubHandler struct {
	// service *service.NotifierService
	// secret  []byte
}

func NewGitHubHandler( /* deps */ ) *GitHubHandler {
	return &GitHubHandler{}
}

func (h *GitHubHandler) Webhook(w http.ResponseWriter, r *http.Request) {
	// TODO: верификация HMAC, парс JSON, вызвать service
	w.WriteHeader(http.StatusOK)
}
