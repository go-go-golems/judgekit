// Package identifier validates bounded, portable identifier strings used as
// stable identities across judgekit documents (construct IDs, evidence IDs,
// claim IDs, protocol names, and similar keys).
//
// Identifiers are normalized for digest stability and human readability. An
// identifier must:
//
//   - be 1..128 bytes long,
//   - start and end with a lowercase letter or digit,
//   - contain only lowercase letters, digits, hyphens, and underscores,
//   - never be empty.
//
// Uppercase characters are rejected on purpose: canonical identities are
// case-folded by convention so that YAML key casing cannot create distinct
// semantic identities by accident.
package identifier

import (
	"errors"
	"fmt"
	"regexp"
)

// Length bounds for identifiers.
const (
	MinLen = 1
	MaxLen = 128
)

var (
	// errEmpty is returned when an identifier is the empty string.
	errEmpty = errors.New("identifier: empty")
	// errTooLong is returned when an identifier exceeds MaxLen bytes.
	errTooLong = fmt.Errorf("identifier: longer than %d bytes", MaxLen)
)

// pattern requires the first and last character to be a lowercase letter or
// digit, with hyphens and underscores allowed only in the middle.
var pattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`)

// Validate returns nil when id is a valid bounded portable identifier, or an
// error describing the first violation otherwise.
func Validate(id string) error {
	if id == "" {
		return errEmpty
	}
	if len(id) > MaxLen {
		return errTooLong
	}
	if !pattern.MatchString(id) {
		return fmt.Errorf("identifier: %q is not a valid identifier (use lowercase letters, digits, hyphens, or underscores; start and end with a letter or digit)", id)
	}
	return nil
}

// MustValidate panics when id is invalid. It is intended for compile-time
// fixtures and tests, not for parsing untrusted input.
func MustValidate(id string) string {
	if err := Validate(id); err != nil {
		panic(err)
	}
	return id
}
