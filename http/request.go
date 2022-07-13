package nethttp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/otel/trace"

	"git.corout.in/golibs/errors"
	"git.corout.in/golibs/tags"
)

// Настройки
const (
	EnableDumpEnvVar = "HTTP_DUMP_ENABLE"
	enable           = "true"

	errNotInitializedRequest = errors.Const("request not initialized")
)

const requestDumpTpl = `
============ REQUEST ========================================
%s
============ END REQUEST ====================================
`

const responseDumpTpl = `
============ RESPONSE =======================================
%s
============ END RESPONSE ===================================
`

// R - интерфейс HTTP-запроса
type R interface {
	Context(ctx context.Context) R
	Timeout(t time.Duration) R
	Param(key, value string) R
	Params(params interface{}) R
	Header(key, value string) R
	Token(token string) R
	FormData(data interface{}) R
	Body(in interface{}) R
	Get() (resp *http.Response, err error)
	Post() (resp *http.Response, err error)
	Put() (resp *http.Response, err error)
	Delete() (resp *http.Response, err error)
	Patch() (resp *http.Response, err error)
	Head() (resp *http.Response, err error)
	Options() (resp *http.Response, err error)
}

type request struct {
	r      *http.Request
	cli    *http.Client
	tracer trace.Tracer
	err    error
}

type formSpec struct {
	Form tags.TagProcessor
}

// Request - конструктор HTTP-запроса
func Request(addr string, args ...interface{}) R {
	req := &request{cli: DefaultClient}
	req.r, req.err = http.NewRequest(http.MethodGet, fmt.Sprintf(addr, args...), nil)

	return req
}

// Context - устанавливает контекст
func (r *request) Context(ctx context.Context) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		r.r = r.r.WithContext(ctx)
	}

	return r
}

// Param устанавливает параметр в запрос
func (r *request) Param(key, value string) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		q := r.r.URL.Query()
		q.Set(key, value)

		r.r.URL.RawQuery = q.Encode()
	}

	return r
}

// Params устанавливает параметры в запрос
func (r *request) Params(params interface{}) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		switch ps := params.(type) {
		case url.Values:
			r.r.URL.RawQuery = ps.Encode()
		case map[string][]interface{}:
			q := r.r.URL.Query()

			for k, v := range ps {
				for i := 0; i < len(v); i++ {
					q.Add(k, fmt.Sprintf("%v", v[i]))
				}
			}

			r.r.URL.RawQuery = q.Encode()
		case map[string]interface{}:
			q := r.r.URL.Query()

			for k, v := range ps {
				q.Add(k, fmt.Sprintf("%v", v))
			}

			r.r.URL.RawQuery = q.Encode()
		case map[string]string:
			q := r.r.URL.Query()

			for k, v := range ps {
				q.Add(k, v)
			}

			r.r.URL.RawQuery = q.Encode()
		default:
			q, err := valuesFromObject(params)
			if err != nil {
				r.err = err

				return r
			}

			r.r.URL.RawQuery = q.Encode()
		}
	}

	return r
}

// Header - устанавливает заголовок
func (r *request) Header(key, value string) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		r.r.Header.Set(key, value)
	}

	return r
}

// Token - устанавливает токен авторизации в заголовок
func (r *request) Token(token string) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		r.Header(HeaderAuth, fmt.Sprintf("Bearer %s", token))
	}

	return r
}

// Timeout - устанавливает таймаут запроса
func (r *request) Timeout(t time.Duration) R {
	if r == nil {
		return nil
	}

	if r.r != nil && r.cli != nil {
		r.cli.Timeout = t
	}

	return r
}

// FormData - устанавливает параметры в URL form data
func (r *request) FormData(data interface{}) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		formData := make(url.Values)

		switch data.(type) {
		case map[string]interface{}:
			for k, v := range data.(map[string]string) {
				formData.Add(k, v)
			}
		default:
			var err error

			formData, err = valuesFromObject(data)
			if err != nil {
				r.err = err

				return r
			}
		}

		r.r.Body = io.NopCloser(bytes.NewBuffer([]byte(formData.Encode()))) // nolint: typecheck // todo: убрать, когда обновятся раннеры CI
	}

	return r
}

// Body - маршалит и устанаваливает тело запроса
func (r *request) Body(body interface{}) R {
	if r == nil {
		return nil
	}

	if r.r != nil {
		data, err := json.Marshal(body)
		if err != nil {
			r.err = errors.Wrap(err, "encode body data")

			return r
		}

		buf := &bytes.Buffer{}

		if _, err = buf.Write(data); err != nil {
			r.err = errors.Wrap(err, "write body data to buffer")

			return r
		}

		r.r.Body = io.NopCloser(buf) // nolint: typecheck // todo: убрать, когда обновятся раннеры CI
	}

	return r
}

// Get - выполняет метод GET
func (r *request) Get() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodGet)
}

// Post - выполняет метод POST
func (r *request) Post() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodPost)
}

// Put - выполняет метод PUT
func (r *request) Put() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodPut)
}

// Delete - выполняет метод DELETE
func (r *request) Delete() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodDelete)
}

// Patch - выполняет метод PATCH
func (r *request) Patch() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodPatch)
}

// Head - выполняет метод HEAD
func (r *request) Head() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodHead)
}

// Options - выполняет метод OPTIONS
func (r *request) Options() (*http.Response, error) {
	if r == nil {
		return nil, nil
	}

	return r.doRequest(http.MethodOptions)
}

func (r *request) doRequest(method string) (*http.Response, error) {
	if err := r.err; err != nil {
		return nil, err
	}

	if r.r == nil {
		return nil, errNotInitializedRequest
	}

	if r.tracer == nil {
		r.tracer = trace.NewNoopTracerProvider().Tracer("nethttp.Request")
	}

	r.r.Method = method
	ctx, span := r.tracer.Start(r.r.Context(), "nethttp.doRequest")

	defer span.End()

	if dumpvar := os.Getenv(EnableDumpEnvVar); dumpvar == enable {
		defer func() {
			dumpRequest, _ := httputil.DumpRequest(r.r, true)
			_, _ = fmt.Fprintf(os.Stdout, requestDumpTpl, string(dumpRequest))
		}()
	}

	req := r.r.WithContext(ctx)
	otelhttptrace.Inject(ctx, req)

	cli := r.cli
	if cli == nil {
		cli = DefaultClient
	}

	resp, err := cli.Do(req)
	if err != nil {
		span.RecordError(errors.Formatted(err))
		return nil, errors.Wrap(err, "do request")
	}

	return resp, nil
}

func valuesFromObject(data interface{}) (url.Values, error) {
	formData := make(url.Values)

	spec, err := tags.ParseSpec(formSpec{
		Form: tags.DirectGetter(func(key string, value interface{}) {
			if rv := reflect.ValueOf(value); rv.Type().Kind() == reflect.Slice {
				for i := 0; i < rv.Len(); i++ {
					formData.Add(key, fmt.Sprintf("%v", rv.Index(i).Interface()))
				}

				return
			}

			formData.Add(key, fmt.Sprintf("%v", value))
		}),
	})
	if err != nil {
		return nil, errors.Wrap(err, "parse tags spec")
	}

	if err = spec.Apply(data); err != nil {
		return nil, errors.Wrap(err, "setup form data")
	}

	return formData, nil
}
