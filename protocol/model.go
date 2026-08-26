package protocol

import (
	"fmt"
	"maps"
	"strings"
)

// ValidateObservedModel checks that the model identity reported by a generator
// exactly matches the protocol's expected provider, model, revision, and
// settings. This is research attribution, not provider authentication.
func ValidateObservedModel(expected, observed ModelIdentity) error {
	if strings.TrimSpace(observed.Provider) == "" || strings.TrimSpace(observed.Model) == "" {
		return fmt.Errorf("observed model provider and model are required")
	}
	if expected.Provider != observed.Provider || expected.Model != observed.Model || expected.Revision != observed.Revision || !maps.Equal(expected.Settings, observed.Settings) {
		return fmt.Errorf("observed model %q/%q revision %q settings %v does not match expected %q/%q revision %q settings %v", observed.Provider, observed.Model, observed.Revision, observed.Settings, expected.Provider, expected.Model, expected.Revision, expected.Settings)
	}
	return nil
}
