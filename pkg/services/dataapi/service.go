package dataapi

import (
	api "github.com/jacobbrewer1/golf-data/pkg/apis/specs/dataapi"
)

type service struct{}

func NewService() api.ServerInterface {
	return &service{}
}
