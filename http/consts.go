package nethttp

import (
	"time"
)

const (
	// DefaultRequestTimeout - таймаут запроса по умолчанию
	DefaultRequestTimeout = time.Second * 45

	HeaderAuth = "Authorization"
	// HeaderContentType - имя заголовка типа контента
	HeaderContentType = "Content-Type"
	// HeaderTraceID - имя заголовка идентификатора трассировки
	HeaderTraceID = "Trace-Request-ID"
)
