package ssh

import (
	"io"
	"time"
)

type clientOptions struct {
	stderr, stdout io.Writer
	sessionTimeOut time.Duration
	prefix         string
}

func WithSessionTimeout(t time.Duration) Option {
	return func(o clientOptions) {
		o.sessionTimeOut = t
	}
}

func WithPrefix(prefix string) Option {
	return func(o clientOptions) {
		o.prefix = prefix
	}
}

func WithStdout(w io.Writer) Option {
	return func(o clientOptions) {
		o.stdout = w
	}
}

func WithStderr(w io.Writer) Option {
	return func(o clientOptions) {
		o.stderr = w
	}
}

func processOptions(opts ...Option) clientOptions {
	options := clientOptions{}

	for o := 0; o < len(opts); o++ {
		opts[o](options)
	}

	if options.sessionTimeOut == 0 {
		options.sessionTimeOut = defaultSessionTimeout
	}

	return options
}
