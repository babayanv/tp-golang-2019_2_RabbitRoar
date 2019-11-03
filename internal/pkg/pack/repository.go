package pack

import "github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/models"

type Repository interface {
	GetByID(packID int) (*models.Pack, error)
	GetQuestions(pack models.Pack) (*[]models.Question, error)
	FetchOrderedByRating(desc bool, page, pageSize int) (*[]models.Pack, error)
	FetchByTags(tags string, page, pageSize int) (*[]models.Pack, error)
	Create(pack models.Pack) (*models.Pack, error)
	Update(pack models.Pack) error
	Delete(packID int) error
}
