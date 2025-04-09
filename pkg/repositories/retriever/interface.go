package retriever

import "github.com/jacobbrewer1/golf-data/pkg/models"

type Repository interface {
	GetClubs() ([]*models.Club, error)
	GetCourses() ([]*models.Course, error)
}
