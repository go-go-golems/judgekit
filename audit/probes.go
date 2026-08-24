package audit

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/internal/canonicaljson"
	"github.com/go-go-golems/judgekit/internal/identifier"
)

// ProbeKind names a class of perturbation that should not affect the construct.
type ProbeKind string

const (
	// Repeat re-runs the same instance to measure run-to-run noise.
	Repeat ProbeKind = "repeat"
	// EvidenceOrder reorders the evidence set to measure order sensitivity.
	EvidenceOrder ProbeKind = "evidence_order"
	// CandidateOrder reverses candidate presentation in pairwise settings.
	CandidateOrder ProbeKind = "candidate_order"
	// PromptParaphrase rewords the prompt without changing its meaning.
	PromptParaphrase ProbeKind = "prompt_paraphrase"
	// FormatTransform changes surface formatting (headings, fences) without
	// changing content.
	FormatTransform ProbeKind = "format_transform"
	// CrossJudge runs a different judge over the same instance to measure
	// cross-judge agreement.
	CrossJudge ProbeKind = "cross_judge"
)

// validProbeKinds is the set of accepted probe kinds.
var validProbeKinds = map[ProbeKind]bool{
	Repeat:           true,
	EvidenceOrder:    true,
	CandidateOrder:   true,
	PromptParaphrase: true,
	FormatTransform:  true,
	CrossJudge:       true,
}

// Probe is one base/variant instance pair plus a statement of what remains
// semantically invariant. The invariant list is required: a probe that does not
// state what should not change cannot localize a sensitivity it finds.
type Probe struct {
	ID              string        `json:"id" yaml:"id"`
	Kind            ProbeKind     `json:"kind" yaml:"kind"`
	BaseInstance    eval.Instance `json:"base_instance" yaml:"base_instance"`
	VariantInstance eval.Instance `json:"variant_instance" yaml:"variant_instance"`
	Invariants      []string      `json:"invariants" yaml:"invariants"`
}

// ValidateProbe returns nil when p is a well-formed probe.
func ValidateProbe(p *Probe) error {
	if err := identifier.Validate(p.ID); err != nil {
		return fmt.Errorf("probe: id: %w", err)
	}
	if !validProbeKinds[p.Kind] {
		return fmt.Errorf("probe %q: kind %q is not recognized", p.ID, p.Kind)
	}
	if err := eval.ValidateInstance(&p.BaseInstance); err != nil {
		return fmt.Errorf("probe %q: base instance: %w", p.ID, err)
	}
	if err := eval.ValidateInstance(&p.VariantInstance); err != nil {
		return fmt.Errorf("probe %q: variant instance: %w", p.ID, err)
	}
	if len(p.Invariants) == 0 {
		return fmt.Errorf("probe %q: at least one invariant must be stated", p.ID)
	}
	seen := make(map[string]struct{}, len(p.Invariants))
	for _, inv := range p.Invariants {
		if strings.TrimSpace(inv) == "" {
			return fmt.Errorf("probe %q: blank invariant", p.ID)
		}
		if _, dup := seen[inv]; dup {
			return fmt.Errorf("probe %q: duplicate invariant %q", p.ID, inv)
		}
		seen[inv] = struct{}{}
	}
	return nil
}

// ProbeSet is an ordered collection of probes, content-addressed so a
// reliability report can pin the exact probe set it was computed over.
type ProbeSet struct {
	Probes []Probe `json:"probes" yaml:"probes"`
	Digest string  `json:"digest" yaml:"digest"`
}

// ValidateProbeSet returns nil when s is well-formed: every probe is valid and
// probe IDs are unique.
func ValidateProbeSet(s *ProbeSet) error {
	seen := make(map[string]struct{}, len(s.Probes))
	for i := range s.Probes {
		if err := ValidateProbe(&s.Probes[i]); err != nil {
			return err
		}
		if _, dup := seen[s.Probes[i].ID]; dup {
			return fmt.Errorf("probe set: duplicate probe id %q", s.Probes[i].ID)
		}
		seen[s.Probes[i].ID] = struct{}{}
	}
	if !strings.HasPrefix(s.Digest, "sha256:") {
		return fmt.Errorf("probe set: digest must be a sha256: digest")
	}
	return nil
}

// NewProbeSet validates probes and returns a content-addressed ProbeSet.
func NewProbeSet(probes []Probe) (ProbeSet, error) {
	set := ProbeSet{Probes: probes}
	for i := range probes {
		if err := ValidateProbe(&probes[i]); err != nil {
			return ProbeSet{}, err
		}
	}
	seen := make(map[string]struct{}, len(probes))
	for i := range probes {
		if _, dup := seen[probes[i].ID]; dup {
			return ProbeSet{}, fmt.Errorf("probe set: duplicate probe id %q", probes[i].ID)
		}
		seen[probes[i].ID] = struct{}{}
	}
	digest, err := canonicaljson.Sum(struct {
		Probes []Probe `json:"probes"`
	}{Probes: probes})
	if err != nil {
		return ProbeSet{}, fmt.Errorf("probe set digest: %w", err)
	}
	set.Digest = digest
	return set, nil
}
