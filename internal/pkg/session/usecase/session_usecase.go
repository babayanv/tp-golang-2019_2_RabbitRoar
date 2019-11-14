package usecase

import (
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/models"
	"github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/session"
)

type sessionUseCase struct {
	repository session.Repository
}

func NewSessionUseCase(repository session.Repository) session.UseCase {
	return &sessionUseCase{
		repository: repository,
	}
}

func (uc sessionUseCase) Create(u models.User) (*string, error) {
	return uc.repository.Create(u)
}

func (uc sessionUseCase) Destroy(sessionId string) {
	_ = uc.repository.Destroy(sessionId)
}
