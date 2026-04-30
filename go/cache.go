// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"container/list" // Note: AX-6 - LRU ordering needs linked-list semantics; no core primitive.
	"net/http"       // Note: AX-6 - Gin cache middleware must handle HTTP headers/methods directly.
	"time"

	core "dappco.re/go"

	"github.com/gin-gonic/gin"
)

// cacheEntry holds a cached response body, status code, headers, and expiry.
type cacheEntry struct {
	status  int
	headers http.Header
	body    []byte
	size    int
	expires time.Time
}

// cacheStore is a simple thread-safe in-memory cache keyed by request URL.
type cacheStore struct {
	mu           core.RWMutex
	entries      map[string]*cacheEntry
	order        *list.List
	index        map[string]*list.Element
	maxEntries   int
	maxBytes     int
	currentBytes int
}

// newCacheStore creates an empty cache store.
// At least one of maxEntries or maxBytes must be positive; if both are
// non-positive the store would be unbounded and newCacheStore returns nil so
// callers can skip registering the middleware.
func newCacheStore(maxEntries, maxBytes int) *cacheStore {
	if maxEntries <= 0 && maxBytes <= 0 {
		return nil
	}
	return &cacheStore{
		entries:    make(map[string]*cacheEntry),
		order:      list.New(),
		index:      make(map[string]*list.Element),
		maxEntries: maxEntries,
		maxBytes:   maxBytes,
	}
}

// get retrieves a non-expired entry for the given key.
// Returns nil if the key is missing or expired.
func (s *cacheStore) get(key string) *cacheEntry {
	s.mu.Lock()
	entry, ok := s.entries[key]
	if !ok {
		s.mu.Unlock()
		return nil
	}

	// Check expiry before promoting in the LRU order so we never move a stale
	// entry to the front. All expiry checking and eviction happen inside the
	// same critical section to avoid a TOCTOU race.
	if time.Now().After(entry.expires) {
		if elem, exists := s.index[key]; exists {
			s.order.Remove(elem)
			delete(s.index, key)
		}
		s.currentBytes -= entry.size
		if s.currentBytes < 0 {
			s.currentBytes = 0
		}
		delete(s.entries, key)
		s.mu.Unlock()
		return nil
	}

	// Only promote to LRU front after confirming the entry is still valid.
	if elem, exists := s.index[key]; exists {
		s.order.MoveToFront(elem)
	}
	s.mu.Unlock()
	return entry
}

// set stores a cache entry with the given TTL.
func (s *cacheStore) set(key string, entry *cacheEntry) {
	s.mu.Lock()
	if entry.size <= 0 {
		entry.size = cacheEntrySize(entry.headers, entry.body)
	}

	if elem, ok := s.index[key]; ok {
		// Reject an oversized replacement before touching LRU state so the
		// existing entry remains intact when the new value cannot fit.
		if s.maxBytes > 0 && entry.size > s.maxBytes {
			s.mu.Unlock()
			return
		}
		if existing, exists := s.entries[key]; exists {
			s.currentBytes -= existing.size
			if s.currentBytes < 0 {
				s.currentBytes = 0
			}
		}
		s.order.MoveToFront(elem)
		s.entries[key] = entry
		s.currentBytes += entry.size
		s.evictBySizeLocked()
		s.mu.Unlock()
		return
	}

	if s.maxBytes > 0 && entry.size > s.maxBytes {
		s.mu.Unlock()
		return
	}

	for (s.maxEntries > 0 && len(s.entries) >= s.maxEntries) || s.wouldExceedBytesLocked(entry.size) {
		if !s.evictOldestLocked() {
			break
		}
	}

	if s.maxBytes > 0 && s.wouldExceedBytesLocked(entry.size) {
		s.mu.Unlock()
		return
	}

	s.entries[key] = entry
	elem := s.order.PushFront(key)
	s.index[key] = elem
	s.currentBytes += entry.size
	s.mu.Unlock()
}

func (s *cacheStore) wouldExceedBytesLocked(nextSize int) bool {
	if s.maxBytes <= 0 {
		return false
	}
	return s.currentBytes+nextSize > s.maxBytes
}

func (s *cacheStore) evictBySizeLocked() {
	for s.maxBytes > 0 && s.currentBytes > s.maxBytes {
		if !s.evictOldestLocked() {
			return
		}
	}
}

func (s *cacheStore) evictOldestLocked() bool {
	back := s.order.Back()
	if back == nil {
		return false
	}
	oldKey := back.Value.(string)
	if existing, ok := s.entries[oldKey]; ok {
		s.currentBytes -= existing.size
		if s.currentBytes < 0 {
			s.currentBytes = 0
		}
	}
	delete(s.entries, oldKey)
	delete(s.index, oldKey)
	s.order.Remove(back)
	return true
}

type cacheBodyBuffer interface {
	Write([]byte) (int, error)
	WriteString(string) (int, error)
	Bytes() []byte
}

// cacheWriter intercepts writes to capture the response body and status.
type cacheWriter struct {
	gin.ResponseWriter
	body cacheBodyBuffer
}

func (w *cacheWriter) Write(data []byte) (
	int,
	error,
) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *cacheWriter) WriteString(s string) (
	int,
	error,
) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// cacheMiddleware returns Gin middleware that caches GET responses in memory.
// Only successful responses (2xx) are cached. Non-GET methods pass through.
func cacheMiddleware(store *cacheStore, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only cache GET requests.
		if c.Request.Method != http.MethodGet {
			c.Next()
			return
		}

		key := c.Request.URL.RequestURI()

		// Serve from cache if a valid entry exists.
		if entry := store.get(key); entry != nil {
			body := entry.body
			metaRewritten := false
			if meta := GetRequestMeta(c); meta != nil {
				body = refreshCachedResponseMeta(entry.body, meta)
				metaRewritten = true
			}

			// staleValidatorHeader returns true for headers that describe the
			// exact bytes of the cached body and must be dropped when the body
			// has been rewritten by refreshCachedResponseMeta.
			staleValidatorHeader := func(canonical string) bool {
				if !metaRewritten {
					return false
				}
				switch canonical {
				case "Etag", "Content-Md5", "Digest":
					return true
				}
				return false
			}

			for k, vals := range entry.headers {
				canonical := http.CanonicalHeaderKey(k)
				if canonical == "X-Request-Id" {
					continue
				}
				if canonical == "Content-Length" {
					continue
				}
				if staleValidatorHeader(canonical) {
					continue
				}
				for _, v := range vals {
					c.Writer.Header().Add(k, v)
				}
			}
			if requestID := GetRequestID(c); requestID != "" {
				c.Writer.Header().Set("X-Request-ID", requestID)
			} else if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
				c.Writer.Header().Set("X-Request-ID", requestID)
			}
			c.Writer.Header().Set("X-Cache", "HIT")
			c.Writer.Header().Set("Content-Length", core.Itoa(len(body)))
			c.Writer.WriteHeader(entry.status)
			_, _ = c.Writer.Write(body)
			c.Abort()
			return
		}

		// Wrap the writer to capture the response.
		cw := &cacheWriter{
			ResponseWriter: c.Writer,
			body:           core.NewBuffer(),
		}
		c.Writer = cw

		c.Next()

		// Only cache successful responses.
		status := cw.ResponseWriter.Status()
		if status >= 200 && status < 300 {
			headers := make(http.Header)
			for key, vals := range cw.ResponseWriter.Header() {
				headers[key] = append([]string(nil), vals...)
			}
			store.set(key, &cacheEntry{
				status:  status,
				headers: headers,
				body:    cw.body.Bytes(),
				size:    cacheEntrySize(headers, cw.body.Bytes()),
				expires: time.Now().Add(ttl),
			})
		}
	}
}

// refreshCachedResponseMeta updates the meta envelope in a cached JSON body so
// request-scoped metadata reflects the current request instead of the cache fill.
// Non-JSON bodies, malformed JSON, and responses without a top-level object are
// returned unchanged.
func refreshCachedResponseMeta(body []byte, meta *Meta) []byte {
	return refreshResponseMetaBody(body, meta)
}

func cacheEntrySize(headers http.Header, body []byte) int {
	size := len(body)
	for key, vals := range headers {
		size += len(key)
		for _, val := range vals {
			size += len(val)
		}
	}
	return size
}
