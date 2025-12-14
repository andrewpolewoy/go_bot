package repository

type UserBinding struct {
	TelegramID  int64
	GitHubLogin string
}

type UserRepository interface {
	SaveBinding(binding UserBinding) error
	GetByGitHubLogin(login string) ([]UserBinding, error)
	GetByTelegramID(tgID int64) (*UserBinding, error)
}
