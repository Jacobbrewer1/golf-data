package domain

import (
	"github.com/jacobbrewer1/golf-data/pkg/models"
	repo "github.com/jacobbrewer1/golf-data/pkg/repositories/processor"
)

type Domain interface {
	SaveClub(club *models.Club) error
	SaveCourse(course *models.Course) error
	SaveDetails(details *models.CourseDetails) error
	SaveHole(holes *models.Hole) error
}

type domain struct {
	r repo.Repository
}

func NewDomain(r repo.Repository) Domain {
	return &domain{
		r: r,
	}
}
