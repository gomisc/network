package nethttp

import (
	"encoding/json"
	"io"
	"net/http"

	"git.eth4.dev/golibs/errors"
)

// ResponseOrError - кастует модель ответа или ошибку из HTTP ответа
func ResponseOrError[T any](resp *http.Response, wanted int, data *T) (*T, error) {
	traceID := resp.Header.Get(HeaderTraceID)

	switch resp.StatusCode {
	case wanted:
		if data != nil {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, errors.Wrap(err, "read response body")
			}

			if err = json.Unmarshal(body, data); err != nil {
				return nil, errors.Wrap(err, "decode response data")
			}

			return data, nil
		}
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound,
		http.StatusRequestTimeout, http.StatusMethodNotAllowed, http.StatusPreconditionFailed:
		statusErr := errors.Ctx().
			Int("code", resp.StatusCode).
			Str("status", http.StatusText(resp.StatusCode)).
			Str("method", resp.Request.Method).
			Str("request-uri", resp.Request.URL.String()).
			Str("trace-request-id", traceID)

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read response body")
		}

		if len(body) != 0 {
			statusErr = statusErr.Str("body", string(body))
		}

		return nil, statusErr.New("get response")
	default:
		var cerr map[string]interface{}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "read response body")
		}

		if err = json.Unmarshal(body, &cerr); err != nil {
			return nil, errors.Ctx().
				Int("code", resp.StatusCode).
				Str("request-uri", resp.Request.RequestURI).
				Str("trace-request-id", traceID).
				New("get response error")
		}

		return nil, errors.Ctx().
			Any("body", cerr).
			Str("trace-request-id", traceID).
			New("unexpected status response")
	}

	return nil, nil
}
