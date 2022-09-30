package ssh

import (
	"context"
)

type (
	Option func(o clientOptions)

	Session interface {
		Command(ctx context.Context, cmd string, args ...any) error
		Shell(ctx context.Context) error
		Close() error
	}

	Client interface {
		Connect(endpoint, key string, o ...Option) error
		Session(endpoint string, o ...Option) (Session, error)
	}
)
