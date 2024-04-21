package storage

import "errors"

var (
	ErrExpressionNotFound = errors.New("expression not found")
	ErrAppExists          = errors.New("app already exists") // lol, it's unrealistic error on first view imho
)
