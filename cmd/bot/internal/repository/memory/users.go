package memory

import "github.com/yourname/crn-bot/internal/repository"

type InMemoryUserRepo struct {
	data map[int64]repository.UserBinding
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
	return &InMemoryUserRepo{
		data: make(map[int64]repository.UserBinding),
	}
}

// TODO: реализовать методы интерфейса
