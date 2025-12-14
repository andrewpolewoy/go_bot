package service

import "github.com/yourname/crn-bot/internal/repository"

type NotifierService struct {
	users repository.UserRepository
	// сюда потом добавим telegram client
}

func NewNotifierService(users repository.UserRepository) *NotifierService {
	return &NotifierService{users: users}
}

// TODO: методы: RegisterGitHubLogin, HandleGitHubEvent, HandleTelegramCommand...
