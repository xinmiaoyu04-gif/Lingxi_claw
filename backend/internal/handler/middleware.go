package handler

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"lingxi-claw/pkg/httpx"
)

// Chain applies middleware outermost-first.
func Chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}

// Recover turns a panic into a 500 INTERNAL_ERROR envelope instead of dropping
// the connection, so the front end always sees the documented shape.
func Recover(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Error("panic recovered", "path", r.URL.Path, "panic", rec)
					httpx.WriteError(w, httpx.ErrInternal("服务器内部错误"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// statusRecorder captures the status code for access logging.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// AccessLog logs one line per request.
func AccessLog(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, r)
			log.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}

// CORS allows the Vite dev server to call the API from another origin
// (API.md §16 sets VITE_API_BASE_URL to this server). Origins come from
// APP_ALLOWED_ORIGINS; "*" allows any origin but then credentials are not
// permitted, per the CORS spec.
func CORS(allowed []string) func(http.Handler) http.Handler {
	allowAll := slices.Contains(allowed, "*")

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			switch {
			case origin == "":
				// Not a browser cross-origin request.
			case allowAll:
				w.Header().Set("Access-Control-Allow-Origin", "*")
			case slices.Contains(allowed, origin):
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Add("Vary", "Origin")
			}

			if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(600))
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NotFoundEnvelope makes unmatched routes and wrong methods return the unified
// envelope rather than Go's plain-text 404/405 pages. A handler's own 404 (for
// example DATASET_NOT_FOUND) already carries httpx.EnvelopeHeader and passes
// through untouched.
func NotFoundEnvelope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &envelopeRecorder{ResponseWriter: w}
		next.ServeHTTP(rec, r)

		if !rec.swallowed {
			return
		}
		out := headerStripper{w}
		switch rec.status {
		case http.StatusNotFound:
			httpx.WriteError(out, httpx.New(http.StatusNotFound, "NOT_FOUND", "接口不存在: "+r.URL.Path))
		case http.StatusMethodNotAllowed:
			httpx.WriteError(out, httpx.New(http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "该接口不支持 "+r.Method))
		}
	})
}

// headerStripper removes the internal envelope marker on the way out so it
// never reaches the client.
type headerStripper struct {
	http.ResponseWriter
}

func (h headerStripper) WriteHeader(code int) {
	h.Header().Del(httpx.EnvelopeHeader)
	h.ResponseWriter.WriteHeader(code)
}

// envelopeRecorder swallows the body of mux-generated 404/405 responses so the
// wrapper can write a JSON envelope instead. Responses that already carry the
// envelope header are passed straight through.
type envelopeRecorder struct {
	http.ResponseWriter
	status    int
	swallowed bool
}

func (e *envelopeRecorder) WriteHeader(code int) {
	e.status = code

	fromHandler := e.Header().Get(httpx.EnvelopeHeader) != ""
	e.Header().Del(httpx.EnvelopeHeader) // internal marker, never sent to clients

	if !fromHandler && (code == http.StatusNotFound || code == http.StatusMethodNotAllowed) {
		e.swallowed = true
		// Drop the headers the mux set; WriteError writes its own.
		e.Header().Del("Content-Type")
		e.Header().Del("Content-Length")
		e.Header().Del("X-Content-Type-Options")
		return
	}
	e.ResponseWriter.WriteHeader(code)
}

func (e *envelopeRecorder) Write(b []byte) (int, error) {
	if e.swallowed {
		return len(b), nil
	}
	return e.ResponseWriter.Write(b)
}
