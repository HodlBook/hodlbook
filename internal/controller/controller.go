package controller

import "hodlbook/internal/repo"

type Controller struct {
	repo *repo.Repository
}

func New(r *repo.Repository) (*Controller, error) {
	if r == nil {
		return nil, ErrNilRepository
	}
	return &Controller{repo: r}, nil
}
