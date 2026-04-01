// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	rateLimitCleanupInterval = time.Minute
	rateLimitStaleAfter      = 10 * time.Minute
)

type rateLimitStore struct {
	mu        sync.Mutex
	buckets   map[string]*rateLimitBucket
	limit     int
	lastSweep time.Time
}

type rateLimitBucket struct {
	mu       sync.Mutex
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

type rateLimitDecision struct {
	allowed    bool
	retryAfter time.Duration
	limit      int
	remaining  int
	resetAt    time.Time
}

func newRateLimitStore(limit int) *rateLimitStore {
	now := time.Now()
	return &rateLimitStore{
		buckets:   make(map[string]*rateLimitBucket),
		limit:     limit,
		lastSweep: now,
	}
}

func (s *rateLimitStore) allow(key string) rateLimitDecision {
	now := time.Now()

	s.mu.Lock()
	bucket, ok := s.buckets[key]
	if !ok || now.Sub(bucket.lastSeen) > rateLimitStaleAfter {
		bucket = &rateLimitBucket{
			tokens:   float64(s.limit),
			last:     now,
			lastSeen: now,
		}
		s.buckets[key] = bucket
	} else {
		bucket.lastSeen = now
	}

	if now.Sub(s.lastSweep) >= rateLimitCleanupInterval {
		for k, candidate := range s.buckets {
			if now.Sub(candidate.lastSeen) > rateLimitStaleAfter {
				delete(s.buckets, k)
			}
		}
		s.lastSweep = now
	}
	s.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	elapsed := now.Sub(bucket.last)
	if elapsed > 0 {
		refill := elapsed.Seconds() * float64(s.limit)
		if bucket.tokens+refill > float64(s.limit) {
			bucket.tokens = float64(s.limit)
		} else {
			bucket.tokens += refill
		}
		bucket.last = now
	}

	if bucket.tokens >= 1 {
		bucket.tokens--
		return rateLimitDecision{
			allowed:   true,
			limit:     s.limit,
			remaining: int(math.Floor(bucket.tokens)),
			resetAt:   now.Add(timeUntilFull(bucket.tokens, s.limit)),
		}
	}

	deficit := 1 - bucket.tokens
	wait := time.Duration(deficit / float64(s.limit) * float64(time.Second))
	if wait <= 0 {
		wait = time.Second / time.Duration(s.limit)
		if wait <= 0 {
			wait = time.Second
		}
	}

	return rateLimitDecision{
		allowed:    false,
		retryAfter: wait,
		limit:      s.limit,
		remaining:  0,
		resetAt:    now.Add(wait),
	}
}

func rateLimitMiddleware(limit int) gin.HandlerFunc {
	if limit <= 0 {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	store := newRateLimitStore(limit)

	return func(c *gin.Context) {
		key := clientRateLimitKey(c)
		decision := store.allow(key)
		if !decision.allowed {
			secs := int(decision.retryAfter / time.Second)
			if decision.retryAfter%time.Second != 0 {
				secs++
			}
			if secs < 1 {
				secs = 1
			}
			setRateLimitHeaders(c, decision.limit, decision.remaining, decision.resetAt)
			c.Header("Retry-After", strconv.Itoa(secs))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, Fail(
				"rate_limit_exceeded",
				"Too many requests",
			))
			return
		}

		c.Next()
		setRateLimitHeaders(c, decision.limit, decision.remaining, decision.resetAt)
	}
}

func setRateLimitHeaders(c *gin.Context, limit, remaining int, resetAt time.Time) {
	if limit > 0 {
		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	}
	if remaining < 0 {
		remaining = 0
	}
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	if !resetAt.IsZero() {
		reset := resetAt.Unix()
		if reset <= time.Now().Unix() {
			reset = time.Now().Add(time.Second).Unix()
		}
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))
	}
}

func timeUntilFull(tokens float64, limit int) time.Duration {
	if limit <= 0 {
		return 0
	}
	missing := float64(limit) - tokens
	if missing <= 0 {
		return 0
	}
	seconds := missing / float64(limit)
	if seconds <= 0 {
		return 0
	}
	return time.Duration(math.Ceil(seconds * float64(time.Second)))
}

// clientRateLimitKey prefers caller-provided credentials for bucket
// isolation, then falls back to the network address.
func clientRateLimitKey(c *gin.Context) string {
	if apiKey := strings.TrimSpace(c.GetHeader("X-API-Key")); apiKey != "" {
		return "api_key:" + apiKey
	}
	if bearer := bearerTokenFromHeader(c.GetHeader("Authorization")); bearer != "" {
		return "bearer:" + bearer
	}
	if ip := c.ClientIP(); ip != "" {
		return "ip:" + ip
	}
	if c.Request != nil && c.Request.RemoteAddr != "" {
		return "ip:" + c.Request.RemoteAddr
	}
	return "ip:unknown"
}

func bearerTokenFromHeader(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
