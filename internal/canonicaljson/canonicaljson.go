// Package canonicaljson produces deterministic JSON encodings used to compute
// stable semantic digests across judgekit documents.
//
// Canonical JSON is the encoding over which semantic identity is computed.
// Two documents that differ only in YAML key ordering, whitespace, or HTML
// escaping must produce the same canonical bytes and therefore the same
// semantic digest.
//
// The canonical form is:
//
//   - object keys are sorted lexicographically by UTF-8 byte order,
//   - no insignificant whitespace (compact encoding),
//   - no HTML escaping of '<', '>', or '&',
//   - standard JSON escaping of quotes, backslashes, and control characters,
//   - NaN and Inf are rejected because they have no stable JSON form.
//
// Numbers are encoded using the standard library's shortest representation so
// that 5, 0.5, and 0.86 round-trip without spurious trailing zeros.
package canonicaljson

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
)

// errNonFinite is returned when a canonical encoding would contain NaN or Inf.
var errNonFinite = errors.New("canonicaljson: NaN or Inf is not representable in canonical JSON")

// Marshal returns the canonical JSON encoding of v. The value is first
// serialized with encoding/json and then re-encoded through a generic
// intermediate so that struct field declaration order and YAML key order do
// not affect the output; only the sorted key order matters.
func Marshal(v any) ([]byte, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("canonicaljson: marshal: %w", err)
	}
	var node any
	if err := json.Unmarshal(raw, &node); err != nil {
		return nil, fmt.Errorf("canonicaljson: normalize: %w", err)
	}
	var buf bytes.Buffer
	if err := encode(&buf, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Sum returns the SHA-256 digest of the canonical JSON encoding of v, prefixed
// with "sha256:" so digests are self-describing.
func Sum(v any) (string, error) {
	b, err := Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// MustSum is like Sum but panics on error. It is intended for fixtures and
// tests where the input is known to be valid.
func MustSum(v any) string {
	s, err := Sum(v)
	if err != nil {
		panic(err)
	}
	return s
}

// encode writes the canonical form of node to buf.
func encode(buf *bytes.Buffer, node any) error {
	switch t := node.(type) {
	case nil:
		buf.WriteString("null")
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) {
			return errNonFinite
		}
		nb, err := json.Marshal(t)
		if err != nil {
			return fmt.Errorf("canonicaljson: encode number: %w", err)
		}
		buf.Write(nb)
	case string:
		encodeString(buf, t)
	case []any:
		buf.WriteByte('[')
		for i := 0; i < len(t); i++ {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := encode(buf, t[i]); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	case map[string]any:
		if err := encodeObject(buf, t); err != nil {
			return err
		}
	default:
		return fmt.Errorf("canonicaljson: unsupported type %T", node)
	}
	return nil
}

// encodeObject writes a map as a sorted-key object.
func encodeObject(buf *bytes.Buffer, m map[string]any) error {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	buf.WriteByte('{')
	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		encodeString(buf, k)
		buf.WriteByte(':')
		if err := encode(buf, m[k]); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

// encodeString writes a JSON string without HTML escaping so that '<', '>',
// and '&' appear literally. Quotes, backslashes, and control characters are
// still escaped per RFC 8259.
func encodeString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		default:
			if c < 0x20 {
				fmt.Fprintf(buf, `\u%04x`, c)
			} else {
				buf.WriteByte(c)
			}
		}
	}
	buf.WriteByte('"')
}
