package audit

import (
	"math"
	"sort"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/spec"
)

// Disagreement records one construct or claim where the base and variant
// reports differ. Presence flags make missing output explicit: a judge that
// drops a difficult claim or dimension is less reliable, not more reliable.
type Disagreement struct {
	ProbeID        string           `json:"probe_id" yaml:"probe_id"`
	InstanceID     string           `json:"instance_id" yaml:"instance_id"`
	Kind           ProbeKind        `json:"kind" yaml:"kind"`
	ConstructID    spec.ConstructID `json:"construct_id,omitempty" yaml:"construct_id,omitempty"`
	ClaimID        string           `json:"claim_id,omitempty" yaml:"claim_id,omitempty"`
	BasePresent    bool             `json:"base_present" yaml:"base_present"`
	VariantPresent bool             `json:"variant_present" yaml:"variant_present"`
	BaseLabel      string           `json:"base_label,omitempty" yaml:"base_label,omitempty"`
	VariantLabel   string           `json:"variant_label,omitempty" yaml:"variant_label,omitempty"`
	BaseValue      *float64         `json:"base_value,omitempty" yaml:"base_value,omitempty"`
	VariantValue   *float64         `json:"variant_value,omitempty" yaml:"variant_value,omitempty"`
}

// CompareReports walks the union of constructs and claims in two reports. A
// missing result is a disagreement because silently intersecting outputs lets
// a partially failing judge appear perfectly stable.
func CompareReports(probeID string, probeKind ProbeKind, base, variant assessment.Report) []Disagreement {
	var out []Disagreement
	instanceID := base.InstanceID

	baseDims := indexDims(base)
	variantDims := indexDims(variant)
	for _, cid := range unionKeys(baseDims, variantDims) {
		bd, baseOK := baseDims[cid]
		vd, variantOK := variantDims[cid]
		if !baseOK || !variantOK || dimDiffers(bd, vd) {
			out = append(out, Disagreement{
				ProbeID:        probeID,
				InstanceID:     instanceID,
				Kind:           probeKind,
				ConstructID:    spec.ConstructID(cid),
				BasePresent:    baseOK,
				VariantPresent: variantOK,
				BaseValue:      bd.Value,
				VariantValue:   vd.Value,
				BaseLabel:      bd.Label,
				VariantLabel:   vd.Label,
			})
		}
	}

	baseClaims := indexClaims(base)
	variantClaims := indexClaims(variant)
	for _, cid := range unionKeys(baseClaims, variantClaims) {
		bc, baseOK := baseClaims[cid]
		vc, variantOK := variantClaims[cid]
		if !baseOK || !variantOK || bc.Label != vc.Label {
			out = append(out, Disagreement{
				ProbeID:        probeID,
				InstanceID:     instanceID,
				Kind:           probeKind,
				ClaimID:        cid,
				BasePresent:    baseOK,
				VariantPresent: variantOK,
				BaseLabel:      string(bc.Label),
				VariantLabel:   string(vc.Label),
			})
		}
	}
	return out
}

// dimDiffers reports whether two dimension results differ in applicability,
// numeric value, or label.
func dimDiffers(a, b assessment.DimensionResult) bool {
	if a.Applicable != b.Applicable || a.Label != b.Label {
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

func unionKeys[T any](a, b map[string]T) []string {
	set := make(map[string]struct{}, len(a)+len(b))
	for key := range a {
		set[key] = struct{}{}
	}
	for key := range b {
		set[key] = struct{}{}
	}
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
