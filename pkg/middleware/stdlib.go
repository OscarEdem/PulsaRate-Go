package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/OscarEdem/PulsaRate-Go/pkg/limiter"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// KeyExtractor defines a function to extract the rate-limiting key from an HTTP request.
type KeyExtractor func(r *http.Request) string

// DefaultKeyExtractor extracts the client IP address as the rate-limiting key.
func DefaultKeyExtractor(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

// RateLimit returns a standard net/http middleware that enforces rate limiting using the given Limiter.
func RateLimit(l limiter.Limiter, extractors ...KeyExtractor) func(http.Handler) http.Handler {
	extractor := DefaultKeyExtractor
	if len(extractors) > 0 && extractors[0] != nil {
		extractor = extractors[0]
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := extractor(r)

			res := l.Reserve(key, 1)

			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))

			if !res.Allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(res.ResetIn.Seconds()), 10))
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(fmt.Sprintf(`{"error":"rate limit exceeded","retry_after_seconds":%d}`, int64(res.ResetIn.Seconds()))))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
