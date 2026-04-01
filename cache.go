// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"bytes"
	"container/list"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// cacheEntry holds a cached response body, status code, headers, and expiry.
type cacheEntry struct {
	status  int
	headers http.Header
	body    []byte
	expires time.Time
}

// cacheStore is a simple thread-safe in-memory cache keyed by request URL.
type cacheStore struct {
	mu         sync.RWMutex
	entries    map[string]*cacheEntry
	order      *list.List
	index      map[string]*list.Element
	maxEntries int
}

// newCacheStore creates an empty cache store.
func newCacheStore(maxEntries int) *cacheStore {
	return &cacheStore{
		entries:    make(map[string]*cacheEntry),
		order:      list.New(),
		index:      make(map[string]*list.Element),
		maxEntries: maxEntries,
	}
}

// get retrieves a non-expired entry for the given key.
// Returns nil if the key is missing or expired.
func (s *cacheStore) get(key string) *cacheEntry {
	s.mu.Lock()
	entry, ok := s.entries[key]
	if ok {
		if elem, exists := s.index[key]; exists {
			s.order.MoveToFront(elem)
		}
	}
	s.mu.Unlock()

	if !ok {
		return nil
	}
	if time.Now().After(entry.expires) {
		s.mu.Lock()
		if elem, exists := s.index[key]; exists {
			s.order.Remove(elem)
			delete(s.index, key)
		}
		delete(s.entries, key)
		s.mu.Unlock()
		return nil
	}
	return entry
}

// set stores a cache entry with the given TTL.
func (s *cacheStore) set(key string, entry *cacheEntry) {
	s.mu.Lock()
	if elem, ok := s.index[key]; ok {
		s.order.MoveToFront(elem)
		s.entries[key] = entry
		s.mu.Unlock()
		return
	}

	if s.maxEntries > 0 && len(s.entries) >= s.maxEntries {
		back := s.order.Back()
		if back != nil {
			oldKey := back.Value.(string)
			delete(s.entries, oldKey)
			delete(s.index, oldKey)
			s.order.Remove(back)
		}
	}

	s.entries[key] = entry
	elem := s.order.PushFront(key)
	s.index[key] = elem
	s.mu.Unlock()
}

// cacheWriter intercepts writes to capture the response body and status.
type cacheWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *cacheWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *cacheWriter) WriteString(s string) (int, error) {
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
			if meta := GetRequestMeta(c); meta != nil {
				body = refreshCachedResponseMeta(entry.body, meta)
			}

			for k, vals := range entry.headers {
				if http.CanonicalHeaderKey(k) == "X-Request-ID" {
					continue
				}
				if http.CanonicalHeaderKey(k) == "Content-Length" {
					continue
				}
				for _, v := range vals {
					c.Writer.Header().Set(k, v)
				}
			}
			if requestID := GetRequestID(c); requestID != "" {
				c.Writer.Header().Set("X-Request-ID", requestID)
			} else if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
				c.Writer.Header().Set("X-Request-ID", requestID)
			}
			c.Writer.Header().Set("X-Cache", "HIT")
			c.Writer.Header().Set("Content-Length", strconv.Itoa(len(body)))
			c.Writer.WriteHeader(entry.status)
			_, _ = c.Writer.Write(body)
			c.Abort()
			return
		}

		// Wrap the writer to capture the response.
		cw := &cacheWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = cw

		c.Next()

		// Only cache successful responses.
		status := cw.ResponseWriter.Status()
		if status >= 200 && status < 300 {
			headers := make(http.Header)
			maps.Copy(headers, cw.ResponseWriter.Header())
			store.set(key, &cacheEntry{
				status:  status,
				headers: headers,
				body:    cw.body.Bytes(),
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
	if meta == nil {
		return body
	}

	var payload any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&payload); err != nil {
		return body
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return body
	}

	obj, ok := payload.(map[string]any)
	if !ok {
		return body
	}

	current := map[string]any{}
	if existing, ok := obj["meta"].(map[string]any); ok {
		current = existing
	}

	if meta.RequestID != "" {
		current["request_id"] = meta.RequestID
	}
	if meta.Duration != "" {
		current["duration"] = meta.Duration
	}

	obj["meta"] = current

	updated, err := json.Marshal(obj)
	if err != nil {
		return body
	}
	return updated
}
