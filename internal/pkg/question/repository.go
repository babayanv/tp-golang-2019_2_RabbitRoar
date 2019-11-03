package question

import "github.com/go-park-mail-ru/2019_2_RabbitRoar/internal/pkg/models"

type Repository interface {
	GetByID(questionID int) (*models.Question, error)
	FetchByTags(tags string, pageSize, page int) (*[]models.Question, error)
	FetchOrderedByRating(desc bool, pageSize, page int) (*[]models.Question, error)
	Create(question models.Question) (*models.Question, error)
	Update(question models.Question) error
	Delete(questionID int) error
}
