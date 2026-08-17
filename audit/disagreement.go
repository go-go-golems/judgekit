package audit

import (
	"math"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/spec"
)

// Disagreement records one construct or claim where the base and variant
// reports differ. It localizes the sensitivity a probe finds. BaseLabel and
// VariantLabel are strings so they can carry both support labels (for claim
// disagreements) and free-form dimension labels (for label dimensions such as
// abstention).
type Disagreement struct {
	ProbeID      string           `json:"probe_id" yaml:"probe_id"`
	InstanceID   string           `json:"instance_id" yaml:"instance_id"`
	Kind         ProbeKind        `json:"kind" yaml:"kind"`
	ConstructID  spec.ConstructID `json:"construct_id,omitempty" yaml:"construct_id,omitempty"`
	ClaimID      string           `json:"claim_id,omitempty" yaml:"claim_id,omitempty"`
	BaseLabel    string           `json:"base_label,omitempty" yaml:"base_label,omitempty"`
	VariantLabel string           `json:"variant_label,omitempty" yaml:"variant_label,omitempty"`
	BaseValue    *float64         `json:"base_value,omitempty" yaml:"base_value,omitempty"`
	VariantValue *float64         `json:"variant_value,omitempty" yaml:"variant_value,omitempty"`
}

// CompareReports walks two reports for the same instance and returns the
// disagreements: per-construct value or label differences, and per-claim
// support-label differences. Reports over different instances are not
// compared here; the caller pairs base and variant per probe.
func CompareReports(probeID string, probeKind ProbeKind, base, variant assessment.Report) []Disagreement {
	var out []Disagreement
	instanceID := base.InstanceID

	// Dimensions by construct.
	baseDims := indexDims(base)
	variantDims := indexDims(variant)
	for cid, bd := range baseDims {
		vd, ok := variantDims[cid]
		if !ok {
			continue
		}
		if dimDiffers(bd, vd) {
			out = append(out, Disagreement{
				ProbeID:      probeID,
				InstanceID:   instanceID,
				Kind:         probeKind,
				ConstructID:  spec.ConstructID(cid),
				BaseValue:    bd.Value,
				VariantValue: vd.Value,
				BaseLabel:    bd.Label,
				VariantLabel: vd.Label,
			})
		}
	}

	// Claim results by claim ID.
	baseClaims := indexClaims(base)
	variantClaims := indexClaims(variant)
	for cid, bc := range baseClaims {
		vc, ok := variantClaims[cid]
		if !ok {
			continue
		}
		if bc.Label != vc.Label {
			out = append(out, Disagreement{
				ProbeID:      probeID,
				InstanceID:   instanceID,
				Kind:         probeKind,
				ClaimID:      cid,
				BaseLabel:    string(bc.Label),
				VariantLabel: string(vc.Label),
			})
		}
	}
	return out
}

// dimDiffers reports whether two dimension results differ in value or label.
func dimDiffers(a, b assessment.DimensionResult) bool {
	if a.Label != b.Label {
		return true
	}
	if (a.Value == nil) != (b.Value == nil) {
		return true
	}
	if a.Value != nil && b.Value != nil {
		return math.Abs(*a.Value-*b.Value) > 1e-9
	}
	return false
}

func indexDims(r assessment.Report) map[string]assessment.DimensionResult {
	m := make(map[string]assessment.DimensionResult, len(r.Dimensions))
	for _, d := range r.Dimensions {
		m[string(d.ConstructID)] = d
	}
	return m
}

func indexClaims(r assessment.Report) map[string]assessment.ClaimAssessment {
	m := make(map[string]assessment.ClaimAssessment, len(r.ClaimResults))
	for _, c := range r.ClaimResults {
		m[c.ClaimID] = c
	}
	return m
}
