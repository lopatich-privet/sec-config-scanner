package parser

import (
	"mime"
	"strings"
)

const maxFormatLen = 16

// SupportedContentTypes maps MIME media types to Format values.
// Keys must be lowercase (mime.ParseMediaType returns lowercase).
// To add TOML/XML support, add one entry per media type.
var SupportedContentTypes = map[string]Format{
	"application/json":   FormatJSON,
	"application/x-json": FormatJSON,
	"text/json":          FormatJSON,
	"application/yaml":   FormatYAML,
	"application/x-yaml": FormatYAML,
	"text/yaml":          FormatYAML,
	"text/x-yaml":        FormatYAML,
}

// supportedFormats is a whitelist of valid Format string values.
var supportedFormats = map[Format]bool{
	FormatJSON: true,
	FormatYAML: true,
}

// FormatFromContentType parses a Content-Type header value using
// mime.ParseMediaType (handles "application/json; charset=UTF-8" correctly)
// and looks up the media type in SupportedContentTypes.
func FormatFromContentType(raw string) (Format, bool) {
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return "", false
	}
	f, ok := SupportedContentTypes[mediaType]
	return f, ok
}

// FormatFromString validates a format string (for gRPC).
// It converts to lowercase, checks length <= maxFormatLen,
// and verifies the value against the supportedFormats whitelist.
func FormatFromString(s string) (Format, bool) {
	if len(s) == 0 || len(s) > maxFormatLen {
		return "", false
	}
	f := Format(strings.ToLower(s))
	if !supportedFormats[f] {
		return "", false
	}
	return f, true
}
