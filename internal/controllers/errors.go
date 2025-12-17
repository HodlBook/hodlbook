package controllers

import "errors"

var (
	ErrNilRepository = errors.New("repository cannot be nil")
)
