package nethttp

import (
	"encoding/json"
	"io"
	"net/http"

	"git.eth4.dev/golibs/errors"
	"git.eth4.dev/golibs/types/caster"
)

// ResponseOrError - кастует модель ответа или ошибку из HTTP ответа
func ResponseOrError(resp *http.Response, wanted int, data interface{}) error {
	traceID := resp.Header.Get(HeaderTraceID)

	switch resp.StatusCode {
	case wanted:
		if data != nil {
			bodyCaster := caster.Default()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return errors.Wrap(err, "read response body")
			}

			if err = bodyCaster.Cast(body, data); err != nil {
				return errors.Wrap(err, "decode response data")
			}
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
			return errors.Wrap(err, "read response body")
		}

		if len(body) != 0 {
			statusErr = statusErr.Str("body", string(body))
		}

		return statusErr.New("get response")
	default:
		var cerr map[string]interface{}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}

		if err = json.Unmarshal(body, &cerr); err != nil {
			return errors.Ctx().
				Int("code", resp.StatusCode).
				Str("request-uri", resp.Request.RequestURI).
				Str("trace-request-id", traceID).
				New("get response error")
		}

		return errors.Ctx().
			Any("body", cerr).
			Str("trace-request-id", traceID).
			New("unexpected status response")
	}

	return nil
}
