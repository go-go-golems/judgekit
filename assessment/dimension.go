package assessment

import (
	"github.com/go-go-golems/judgekit/spec"
)

// DimensionResult is the aggregated result for one construct: a numeric value,
// a label, or a not-applicable marker. It carries the evidence and diagnostics
// that justify it so a reader can audit how the number was produced.
type DimensionResult struct {
	ConstructID spec.ConstructID `json:"construct_id" yaml:"construct_id"`
	Applicable  bool             `json:"applicable" yaml:"applicable"`
	Value       *float64         `json:"value,omitempty" yaml:"value,omitempty"`
	Label       string           `json:"label,omitempty" yaml:"label,omitempty"`
	Confidence  *float64         `json:"confidence,omitempty" yaml:"confidence,omitempty"`
	EvidenceIDs []string         `json:"evidence_ids,omitempty" yaml:"evidence_ids,omitempty"`
	Diagnostics []string         `json:"diagnostics,omitempty" yaml:"diagnostics,omitempty"`
}
