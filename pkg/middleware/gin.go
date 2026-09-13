package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/OscarEdem/PulsaRate-Go/pkg/limiter"
)

// ------------------------------------------------------------------------------------------------------------------                                                                                                                                                                                #*eddiere

// GinRateLimit returns a native Gin middleware handler enforcing rate limits.
func GinRateLimit(l limiter.Limiter, extractors ...KeyExtractor) gin.HandlerFunc {
	extractor := DefaultKeyExtractor
	if len(extractors) > 0 && extractors[0] != nil {
		extractor = extractors[0]
	}

	return func(c *gin.Context) {
		key := extractor(c.Request)
		res := l.Reserve(key, 1)

		c.Header("X-RateLimit-Remaining", strconv.FormatInt(res.Remaining, 10))

		if !res.Allowed {
			c.Header("Retry-After", strconv.FormatInt(int64(res.ResetIn.Seconds()), 10))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":               "rate limit exceeded",
				"retry_after_seconds": int64(res.ResetIn.Seconds()),
				"message":             fmt.Sprintf("Quota exceeded for %s. Try again in %d seconds.", key, int64(res.ResetIn.Seconds())),
			})
			return
		}

		c.Next()
	}
}
