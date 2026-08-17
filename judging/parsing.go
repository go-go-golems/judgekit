package judging

import (
	"github.com/go-go-golems/judgekit/internal/strictdecode"
)

// StripSingleCodeFence removes one code fence that wraps the whole response.
// It is re-exported from internal/strictdecode so judging callers and examples
// have a stable helper without importing an internal package.
func StripSingleCodeFence(raw string) string {
	return strictdecode.StripSingleCodeFence(raw)
}

// DecodeJSONObjectStrict decodes exactly one JSON object from raw into out,
// rejecting prose, unknown fields, and trailing data. It returns a typed
// StructuralError suitable for one repair attempt.
func DecodeJSONObjectStrict[T any](raw string) (T, error) {
	return strictdecode.DecodeJSONObjectStrict[T](raw)
}

// StructuralError is re-exported so callers can use errors.As/isStructural.
type StructuralError = strictdecode.StructuralError

// IsStructural reports whether err is a structural parsing failure.
func IsStructural(err error) bool {
	return strictdecode.IsStructural(err)
}
