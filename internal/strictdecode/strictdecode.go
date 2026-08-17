// Package strictdecode parses language-model structured output strictly.
//
// Judgekit judges request JSON objects from models. Models sometimes wrap the
// object in a single code fence or add stray prose. strictdecode defines a
// single, opinionated parsing path that:
//
//   - optionally strips one code fence that wraps the whole response,
//   - requires exactly one JSON object (the first non-whitespace byte must be
//     '{'),
//   - rejects unknown fields,
//   - rejects trailing data after the single JSON value,
//   - returns a typed StructuralError suitable for one repair attempt.
//
// The default path never searches arbitrary prose for the first matching
// "{...}". That tolerance hides model failures; callers that need it must
// implement their own tolerant parser.
package strictdecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Kind classifies a structural failure so callers can decide whether a repair
// is appropriate.
type Kind string

const (
	KindNoObject     Kind = "no_json_object"
	KindDecode       Kind = "decode"
	KindUnknownField Kind = "unknown_field"
	KindMissingField Kind = "missing_field"
	KindTrailingData Kind = "trailing_data"
	KindNonFinite    Kind = "non_finite"
)

// StructuralError describes a structural failure in model output. It is the
// only error type returned by this package so callers can distinguish
// repairable structural problems from semantic validation failures.
type StructuralError struct {
	Kind  Kind
	Msg   string
	Raw   string
	cause error
}

// Error implements the error interface.
func (e *StructuralError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("strictdecode: %s: %s: %v", e.Kind, e.Msg, e.cause)
	}
	return fmt.Sprintf("strictdecode: %s: %s", e.Kind, e.Msg)
}

// Unwrap returns the wrapped cause, if any.
func (e *StructuralError) Unwrap() error { return e.cause }

// IsStructural reports whether err is a *StructuralError.
func IsStructural(err error) bool {
	var target *StructuralError
	return errors.As(err, &target)
}

// structural builds a StructuralError with an excerpt of raw for debugging.
func structural(kind Kind, msg string, raw string, cause error) *StructuralError {
	excerpt := raw
	if len(excerpt) > 200 {
		excerpt = excerpt[:200]
	}
	return &StructuralError{Kind: kind, Msg: msg, Raw: excerpt, cause: cause}
}

// StripSingleCodeFence removes one code fence that wraps the entire response.
// If the response is prose with an embedded fence, or has no fence, it is
// returned unchanged so the strict decoder can reject the surrounding prose.
func StripSingleCodeFence(raw string) string {
	trimmed := strings.TrimLeft(raw, " \t\r\n")
	if !strings.HasPrefix(trimmed, "```") {
		return raw
	}
	rest := trimmed[len("```"):]
	// An optional language tag runs to the first newline.
	nl := strings.IndexByte(rest, '\n')
	if nl < 0 {
		return raw
	}
	body := rest[nl+1:]
	bodyRight := strings.TrimRight(body, " \t\r\n")
	if !strings.HasSuffix(bodyRight, "```") {
		return raw
	}
	inner := bodyRight[:len(bodyRight)-len("```")]
	inner = strings.TrimRight(inner, "\r\n")
	return inner
}

// firstNonSpace returns the first non-whitespace byte of s, or 0 if s is all
// whitespace.
func firstNonSpace(s string) byte {
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ' ', '\t', '\r', '\n':
		default:
			return s[i]
		}
	}
	return 0
}

// DecodeJSONObjectStrict decodes exactly one JSON object from raw into out.
// It strips one wrapping code fence, requires the remaining content to be a
// single JSON object, rejects unknown fields, and rejects trailing data.
func DecodeJSONObjectStrict[T any](raw string) (T, error) {
	var zero T
	cleaned := StripSingleCodeFence(raw)
	trimmed := strings.TrimSpace(cleaned)
	if firstNonSpace(trimmed) != '{' {
		return zero, structural(KindNoObject, "expected a single JSON object", raw, nil)
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	dec.DisallowUnknownFields()
	var t T
	if err := dec.Decode(&t); err != nil {
		return zero, structural(classifyDecode(err), "decode failed", raw, err)
	}
	var skip any
	if err := dec.Decode(&skip); err != io.EOF {
		return zero, structural(KindTrailingData, "expected exactly one JSON object, found trailing data", raw, err)
	}
	return t, nil
}

// classifyDecode maps a json error to a more specific Kind when possible.
func classifyDecode(err error) Kind {
	var te *json.UnmarshalTypeError
	if errors.As(err, &te) {
		if te.Field != "" {
			return KindDecode
		}
		return KindDecode
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "unknown field"):
		return KindUnknownField
	case strings.Contains(msg, "cannot find") || strings.Contains(msg, "missing"):
		return KindMissingField
	case strings.Contains(msg, "NaN") || strings.Contains(msg, "infinity") || strings.Contains(msg, "Inf"):
		return KindNonFinite
	default:
		return KindDecode
	}
}

// ValidateSingleJSONValue reports whether raw contains exactly one JSON value
// and nothing else after an optional wrapping code fence. It does not decode
// into a typed value.
func ValidateSingleJSONValue(raw string) error {
	cleaned := StripSingleCodeFence(raw)
	trimmed := strings.TrimSpace(cleaned)
	if trimmed == "" {
		return structural(KindNoObject, "empty response", raw, nil)
	}
	dec := json.NewDecoder(strings.NewReader(trimmed))
	var skip any
	if err := dec.Decode(&skip); err != nil {
		return structural(classifyDecode(err), "decode failed", raw, err)
	}
	if err := dec.Decode(&skip); err != io.EOF {
		return structural(KindTrailingData, "expected exactly one JSON value, found trailing data", raw, err)
	}
	return nil
}
