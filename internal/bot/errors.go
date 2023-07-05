package bot

import "errors"

var (
	ErrNotFound    = errors.New("not found")
	ErrBotNotFound = errors.New("bot not found")
)
