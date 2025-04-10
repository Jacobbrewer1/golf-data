package api

import (
	"github.com/jacobbrewer1/golf-data/pkg/apis/specs/api"
	"github.com/jacobbrewer1/golf-data/pkg/services/api/domain"
)

type service struct {
	dom domain.Domain
}

func NewService(dom domain.Domain) api.ServerInterface {
	return &service{
		dom: dom,
	}
}
