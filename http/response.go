package nethttp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"

	"git.eth4.dev/golibs/errors"
	"git.eth4.dev/golibs/types/caster"
)

// ResponseOrError - кастует модель ответа или ошибку из HTTP ответа
func ResponseOrError(resp *http.Response, wanted int, data interface{}) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "read response body")
	}

	traceID := resp.Header.Get(HeaderTraceID)

	if dumpvar := os.Getenv(EnableDumpEnvVar); dumpvar == enable {
		defer func() {
			dumpResponse, e := httputil.DumpResponse(resp, true)
			if e != nil {
				dumpResponse, _ = httputil.DumpResponse(resp, false)
			}

			_, _ = fmt.Fprintf(os.Stdout, responseDumpTpl, string(dumpResponse))
		}()
	}

	switch resp.StatusCode {
	case wanted:
		if len(body) != 0 && data != nil {
			caster := caster.Default()

			if err = caster.Cast(body, data); err != nil {
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

		if len(body) != 0 {
			statusErr = statusErr.Str("body", string(body))
		}

		return statusErr.New("get response")
	default:
		var cerr map[string]interface{}

		if err = json.Unmarshal(body, &cerr); err != nil {
			return errors.Ctx().
				Int("code", resp.StatusCode).
				Str("request-uri", resp.Request.RequestURI).
				Str("trace-request-id", traceID).
				New("get response error")
		}

		code := cerr["code"].(string)
		if code == "" {
			code = strconv.Itoa(resp.StatusCode)
		}

		return errors.Ctx().
			Str("code", code).
			Str("status", cerr["message"].(string)).
			Str("trace-request-id", traceID).
			New("unexpected status response")
	}

	return nil
}
