package config

import (
	"log/slog"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type Config struct {
	BaseURL string
	Client  *gitlab.Client
	Logger  *slog.Logger
	Token   string
}
