// SPDX-License-Identifier: EUPL-1.2

package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/text/language"
)

// i18nContextKey is the Gin context key for the detected locale string.
const i18nContextKey = "i18n.locale"

// i18nMessagesKey is the Gin context key for the message lookup map.
const i18nMessagesKey = "i18n.messages"

// i18nCatalogKey is the Gin context key for the full locale->message catalog.
const i18nCatalogKey = "i18n.catalog"

// i18nDefaultLocaleKey stores the configured default locale for fallback lookups.
const i18nDefaultLocaleKey = "i18n.default_locale"

// I18nConfig configures the internationalisation middleware.
type I18nConfig struct {
	// DefaultLocale is the fallback locale when the Accept-Language header
	// is absent or does not match any supported locale. Defaults to "en".
	DefaultLocale string

	// Supported lists the locale tags the application supports.
	// Each entry should be a BCP 47 language tag (e.g. "en", "fr", "de").
	// If empty, only the default locale is supported.
	Supported []string

	// Messages maps locale tags to key-value message pairs.
	// For example: {"en": {"greeting": "Hello"}, "fr": {"greeting": "Bonjour"}}
	// This is optional — handlers can use GetLocale() alone for custom logic.
	Messages map[string]map[string]string
}

// WithI18n adds Accept-Language header parsing and locale detection middleware.
// The middleware uses golang.org/x/text/language for RFC 5646 language matching
// with quality weighting support. The detected locale is stored in the Gin
// context and can be retrieved by handlers via GetLocale().
//
// If messages are configured, handlers can look up localised strings via
// GetMessage(). This is a lightweight bridge — the go-i18n grammar engine
// can replace the message map later.
func WithI18n(cfg ...I18nConfig) Option {
	return func(e *Engine) {
		var config I18nConfig
		if len(cfg) > 0 {
			config = cfg[0]
		}
		if config.DefaultLocale == "" {
			config.DefaultLocale = "en"
		}

		// Build the language.Matcher from supported locales.
		tags := []language.Tag{language.Make(config.DefaultLocale)}
		for _, s := range config.Supported {
			tag := language.Make(s)
			// Avoid duplicating the default if it also appears in Supported.
			if tag != tags[0] {
				tags = append(tags, tag)
			}
		}
		matcher := language.NewMatcher(tags)

		e.middlewares = append(e.middlewares, i18nMiddleware(matcher, config))
	}
}

// i18nMiddleware returns Gin middleware that parses Accept-Language, matches
// it against supported locales, and stores the resolved BCP 47 tag in the context.
func i18nMiddleware(matcher language.Matcher, cfg I18nConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		accept := c.GetHeader("Accept-Language")

		var locale string
		if accept == "" {
			locale = cfg.DefaultLocale
		} else {
			tags, _, _ := language.ParseAcceptLanguage(accept)
			tag, _, _ := matcher.Match(tags...)
			locale = tag.String()
		}

		c.Set(i18nContextKey, locale)
		c.Set(i18nDefaultLocaleKey, cfg.DefaultLocale)

		// Attach the message map for this locale if messages are configured.
		if cfg.Messages != nil {
			c.Set(i18nCatalogKey, cfg.Messages)
			if msgs, ok := cfg.Messages[locale]; ok {
				c.Set(i18nMessagesKey, msgs)
			}
		}

		c.Next()
	}
}

// GetLocale returns the detected locale for the current request.
// Returns "en" if the i18n middleware was not applied.
func GetLocale(c *gin.Context) string {
	if v, ok := c.Get(i18nContextKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "en"
}

// GetMessage looks up a localised message by key for the current request.
// Returns the message string and true if found, or empty string and false
// if the key does not exist or the i18n middleware was not applied.
func GetMessage(c *gin.Context, key string) (string, bool) {
	if v, ok := c.Get(i18nMessagesKey); ok {
		if msgs, ok := v.(map[string]string); ok {
			if msg, ok := msgs[key]; ok {
				return msg, true
			}
		}
	}

	catalog, _ := c.Get(i18nCatalogKey)
	msgsByLocale, _ := catalog.(map[string]map[string]string)
	if len(msgsByLocale) == 0 {
		return "", false
	}

	locales := localeFallbacks(GetLocale(c))
	if defaultLocale, ok := c.Get(i18nDefaultLocaleKey); ok {
		if fallback, ok := defaultLocale.(string); ok && fallback != "" {
			locales = append(locales, localeFallbacks(fallback)...)
		}
	}

	seen := make(map[string]struct{}, len(locales))
	for _, locale := range locales {
		if locale == "" {
			continue
		}
		if _, ok := seen[locale]; ok {
			continue
		}
		seen[locale] = struct{}{}
		if msgs, ok := msgsByLocale[locale]; ok {
			if msg, ok := msgs[key]; ok {
				return msg, true
			}
		}
	}

	return "", false
}

// localeFallbacks returns the locale and its parent tags in order from
// most specific to least specific. For example, "fr-CA" yields
// ["fr-CA", "fr"] and "zh-Hant-TW" yields ["zh-Hant-TW", "zh-Hant", "zh"].
func localeFallbacks(locale string) []string {
	locale = strings.TrimSpace(strings.ReplaceAll(locale, "_", "-"))
	if locale == "" {
		return nil
	}

	parts := strings.Split(locale, "-")
	if len(parts) == 0 {
		return []string{locale}
	}

	fallbacks := make([]string, 0, len(parts))
	for i := len(parts); i >= 1; i-- {
		fallbacks = append(fallbacks, strings.Join(parts[:i], "-"))
	}

	return fallbacks
}
