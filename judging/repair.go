package judging

import (
	"fmt"

	"github.com/go-go-golems/judgekit/internal/strictdecode"
)

// Repairer rewrites a prompt to ask the model to correct a structural failure.
// Only structural failures (bad JSON shape) should be repaired by default; a
// semantically invalid assessment fails closed instead of retrying.
type Repairer interface {
	RepairPrompt(original string, failure *strictdecode.StructuralError) (string, error)
}

// DefaultRepairer appends a correction instruction naming the structural
// failure. It mirrors the proven CoinVault repair strategy without hard-coding
// any product vocabulary.
type DefaultRepairer struct{}

// RepairPrompt returns a prompt that asks for one corrected JSON object.
func (DefaultRepairer) RepairPrompt(original string, failure *strictdecode.StructuralError) (string, error) {
	if failure == nil {
		return original, nil
	}
	return fmt.Sprintf("%s\n\nYour previous response was structurally invalid: %s. Return one corrected JSON object that strictly follows the requested schema.", original, failure.Error()), nil
}
