// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"net/http"
	"strconv"
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

func newRateLimitStore(limit int) *rateLimitStore {
	now := time.Now()
	return &rateLimitStore{
		buckets:   make(map[string]*rateLimitBucket),
		limit:     limit,
		lastSweep: now,
	}
}

func (s *rateLimitStore) allow(key string) (bool, time.Duration) {
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
		return true, 0
	}

	deficit := 1 - bucket.tokens
	wait := time.Duration(deficit / float64(s.limit) * float64(time.Second))
	if wait <= 0 {
		wait = time.Second / time.Duration(s.limit)
		if wait <= 0 {
			wait = time.Second
		}
	}

	return false, wait
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
		allowed, retryAfter := store.allow(key)
		if !allowed {
			secs := int(retryAfter / time.Second)
			if retryAfter%time.Second != 0 {
				secs++
			}
			if secs < 1 {
				secs = 1
			}
			c.Header("Retry-After", strconv.Itoa(secs))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, Fail(
				"rate_limit_exceeded",
				"Too many requests",
			))
			return
		}

		c.Next()
	}
}

func clientRateLimitKey(c *gin.Context) string {
	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	if c.Request != nil && c.Request.RemoteAddr != "" {
		return c.Request.RemoteAddr
	}
	return "unknown"
}
