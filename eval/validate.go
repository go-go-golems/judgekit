package eval

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/judgekit/internal/identifier"
)

// ValidateArtifact returns nil when a is a well-formed artifact.
func ValidateArtifact(a *Artifact) error {
	if strings.TrimSpace(a.MediaType) == "" {
		return fmt.Errorf("artifact: media_type is required")
	}
	hasText := strings.TrimSpace(a.Text) != ""
	hasURI := strings.TrimSpace(a.URI) != ""
	if hasText == hasURI {
		return fmt.Errorf("artifact: exactly one of text or uri must be set")
	}
	if !strings.HasPrefix(a.Digest, "sha256:") || len(a.Digest) <= len("sha256:") {
		return fmt.Errorf("artifact: digest must be a non-empty sha256: digest")
	}
	if a.SizeBytes < 0 {
		return fmt.Errorf("artifact: size_bytes must be non-negative")
	}
	if hasText {
		want := TextContentDigest(a.Text)
		if a.Digest != want {
			return fmt.Errorf("artifact: text digest %q does not match content (want %q)", a.Digest, want)
		}
		if a.SizeBytes != int64(len(a.Text)) {
			return fmt.Errorf("artifact: size_bytes %d does not match text length %d", a.SizeBytes, len(a.Text))
		}
	}
	return nil
}

// ValidateEvidenceItem returns nil when e is a well-formed evidence item.
func ValidateEvidenceItem(e *EvidenceItem) error {
	if err := identifier.Validate(e.ID); err != nil {
		return fmt.Errorf("evidence item: %w", err)
	}
	if strings.TrimSpace(e.Kind) == "" {
		return fmt.Errorf("evidence item %q: kind is required", e.ID)
	}
	if err := ValidateArtifact(&e.Content); err != nil {
		return fmt.Errorf("evidence item %q: %w", e.ID, err)
	}
	if strings.TrimSpace(e.SourceID) == "" {
		return fmt.Errorf("evidence item %q: source_id is required", e.ID)
	}
	if e.SourceTime != nil && e.SourceTime.IsZero() {
		return fmt.Errorf("evidence item %q: source_time is zero", e.ID)
	}
	for k, v := range e.Provenance {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("evidence item %q: provenance has a blank key", e.ID)
		}
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("evidence item %q: provenance key %q has a blank value", e.ID, k)
		}
	}
	return nil
}

// ValidateEvidenceSet returns nil when s is a well-formed evidence set: every
// item is valid, item IDs are unique, and the stored digest matches the current
// items and policy digest.
func ValidateEvidenceSet(s *EvidenceSet) error {
	if err := validateEvidenceSetBody(s); err != nil {
		return err
	}
	if !strings.HasPrefix(s.Digest, "sha256:") {
		return fmt.Errorf("evidence set: digest must be a sha256: digest")
	}
	want, err := EvidenceSetDigest(s.Items, s.PolicyDigest)
	if err != nil {
		return fmt.Errorf("evidence set: compute digest: %w", err)
	}
	if s.Digest != want {
		return fmt.Errorf("evidence set: digest %q does not match content (want %q)", s.Digest, want)
	}
	return nil
}

func validateEvidenceSetBody(s *EvidenceSet) error {
	if !strings.HasPrefix(s.PolicyDigest, "sha256:") {
		return fmt.Errorf("evidence set: policy_digest must be a sha256: digest")
	}
	seen := make(map[string]struct{}, len(s.Items))
	for i := range s.Items {
		if err := ValidateEvidenceItem(&s.Items[i]); err != nil {
			return err
		}
		if _, dup := seen[s.Items[i].ID]; dup {
			return fmt.Errorf("evidence set: duplicate item id %q", s.Items[i].ID)
		}
		seen[s.Items[i].ID] = struct{}{}
	}
	return nil
}

// ValidateRequiredFact returns nil when f is a well-formed required fact.
func ValidateRequiredFact(f *RequiredFact) error {
	if err := identifier.Validate(f.ID); err != nil {
		return fmt.Errorf("required fact: %w", err)
	}
	if strings.TrimSpace(f.Description) == "" {
		return fmt.Errorf("required fact %q: description is required", f.ID)
	}
	if math.IsNaN(f.Importance) || math.IsInf(f.Importance, 0) || f.Importance < 0 || f.Importance > 1 {
		return fmt.Errorf("required fact %q: importance %g must be in [0,1]", f.ID, f.Importance)
	}
	seen := make(map[string]struct{}, len(f.EvidenceIDs))
	for _, eid := range f.EvidenceIDs {
		if strings.TrimSpace(eid) == "" {
			return fmt.Errorf("required fact %q: blank evidence id", f.ID)
		}
		if _, dup := seen[eid]; dup {
			return fmt.Errorf("required fact %q: duplicate evidence id %q", f.ID, eid)
		}
		seen[eid] = struct{}{}
	}
	return nil
}

// ValidateInstance returns nil when i is a well-formed instance, including
// cross-reference checks that RequiredFact.EvidenceIDs resolve to evidence
// items in i.Evidence.
func ValidateInstance(i *Instance) error {
	if err := identifier.Validate(i.ID); err != nil {
		return fmt.Errorf("instance: %w", err)
	}
	if err := ValidateArtifact(&i.Input); err != nil {
		return fmt.Errorf("instance %q: input: %w", i.ID, err)
	}
	if err := ValidateArtifact(&i.Candidate); err != nil {
		return fmt.Errorf("instance %q: candidate: %w", i.ID, err)
	}
	if err := ValidateEvidenceSet(&i.Evidence); err != nil {
		return fmt.Errorf("instance %q: %w", i.ID, err)
	}
	if i.Reference != nil {
		if err := ValidateArtifact(i.Reference); err != nil {
			return fmt.Errorf("instance %q: reference: %w", i.ID, err)
		}
	}
	evidenceIDs := make(map[string]struct{}, len(i.Evidence.Items))
	for _, e := range i.Evidence.Items {
		evidenceIDs[e.ID] = struct{}{}
	}
	factIDs := make(map[string]struct{}, len(i.RequiredFacts))
	for idx := range i.RequiredFacts {
		f := &i.RequiredFacts[idx]
		if err := ValidateRequiredFact(f); err != nil {
			return err
		}
		if _, dup := factIDs[f.ID]; dup {
			return fmt.Errorf("instance %q: duplicate required fact id %q", i.ID, f.ID)
		}
		factIDs[f.ID] = struct{}{}
		for _, eid := range f.EvidenceIDs {
			if _, ok := evidenceIDs[eid]; !ok {
				return fmt.Errorf("instance %q: required fact %q references unknown evidence %q", i.ID, f.ID, eid)
			}
		}
	}
	for k, v := range i.Metadata {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("instance %q: metadata has a blank key", i.ID)
		}
		_ = v
	}
	if !strings.HasPrefix(i.Digest, "sha256:") {
		return fmt.Errorf("instance %q: digest must be a sha256: digest", i.ID)
	}
	want, err := InstanceDigest(i)
	if err != nil {
		return fmt.Errorf("instance %q: compute digest: %w", i.ID, err)
	}
	if i.Digest != want {
		return fmt.Errorf("instance %q: digest %q does not match content (want %q)", i.ID, i.Digest, want)
	}
	return nil
}

// NewEvidenceSet validates items, computes the set digest, and returns a
// ready EvidenceSet. policyDigest must be the semantic digest of the evidence
// policy under which the set was admitted.
func NewEvidenceSet(items []EvidenceItem, policyDigest string) (EvidenceSet, error) {
	set := EvidenceSet{Items: items, PolicyDigest: policyDigest}
	if err := validateEvidenceSetBody(&set); err != nil {
		return EvidenceSet{}, err
	}
	digest, err := EvidenceSetDigest(items, policyDigest)
	if err != nil {
		return EvidenceSet{}, fmt.Errorf("evidence set digest: %w", err)
	}
	set.Digest = digest
	return set, nil
}

// NewInstance assembles an instance, validates it (including cross-references),
// computes its digest, and returns a ready Instance.
func NewInstance(id string, input, candidate Artifact, evidence EvidenceSet, reference *Artifact, facts []RequiredFact, metadata map[string]string) (Instance, error) {
	inst := Instance{
		ID:            id,
		Input:         input,
		Candidate:     candidate,
		Evidence:      evidence,
		Reference:     reference,
		RequiredFacts: facts,
		Metadata:      metadata,
	}
	digest, err := InstanceDigest(&inst)
	if err != nil {
		return Instance{}, fmt.Errorf("instance digest: %w", err)
	}
	inst.Digest = digest
	if err := ValidateInstance(&inst); err != nil {
		return Instance{}, err
	}
	return inst, nil
}
