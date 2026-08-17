package assessment

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/internal/identifier"
)

// ValidateSpan returns nil when s is a valid span: non-negative, ordered.
func ValidateSpan(s *Span) error {
	if s.Start < 0 {
		return fmt.Errorf("span: start %d must be non-negative", s.Start)
	}
	if s.End <= s.Start {
		return fmt.Errorf("span: end %d must be greater than start %d", s.End, s.Start)
	}
	return nil
}

// ValidateClaim returns nil when c is a well-formed claim.
func ValidateClaim(c *Claim) error {
	if err := identifier.Validate(c.ID); err != nil {
		return fmt.Errorf("claim: %w", err)
	}
	if strings.TrimSpace(c.Text) == "" {
		return fmt.Errorf("claim %q: text is required", c.ID)
	}
	if math.IsNaN(c.Importance) || math.IsInf(c.Importance, 0) || c.Importance < 0 || c.Importance > 1 {
		return fmt.Errorf("claim %q: importance %g must be in [0,1]", c.ID, c.Importance)
	}
	if c.CandidateSpan != nil {
		if err := ValidateSpan(c.CandidateSpan); err != nil {
			return fmt.Errorf("claim %q: %w", c.ID, err)
		}
	}
	return nil
}

// ValidateClaimAssessment returns nil when a is a well-formed verdict. Entailed
// and contradicted verdicts must cite evidence; only insufficient may cite
// none. If allowedEvidence is non-nil, every cited evidence id must be in it.
func ValidateClaimAssessment(a *ClaimAssessment, allowedEvidence map[string]struct{}) error {
	if err := identifier.Validate(a.ClaimID); err != nil {
		return fmt.Errorf("claim assessment: %w", err)
	}
	if !validSupportLabels[a.Label] {
		return fmt.Errorf("claim assessment %q: label %q is not entailed, contradicted, or insufficient", a.ClaimID, a.Label)
	}
	if strings.TrimSpace(a.Reason) == "" {
		return fmt.Errorf("claim assessment %q: reason is required", a.ClaimID)
	}
	seen := make(map[string]struct{}, len(a.EvidenceIDs))
	for _, eid := range a.EvidenceIDs {
		if strings.TrimSpace(eid) == "" {
			return fmt.Errorf("claim assessment %q: blank evidence id", a.ClaimID)
		}
		if _, dup := seen[eid]; dup {
			return fmt.Errorf("claim assessment %q: duplicate evidence id %q", a.ClaimID, eid)
		}
		seen[eid] = struct{}{}
		if allowedEvidence != nil {
			if _, ok := allowedEvidence[eid]; !ok {
				return fmt.Errorf("claim assessment %q: cites unknown evidence %q", a.ClaimID, eid)
			}
		}
	}
	if a.Label != Insufficient && len(a.EvidenceIDs) == 0 {
		return fmt.Errorf("claim assessment %q: %s verdict requires evidence_ids", a.ClaimID, a.Label)
	}
	if a.Confidence != nil {
		c := *a.Confidence
		if math.IsNaN(c) || math.IsInf(c, 0) || c < 0 || c > 1 {
			return fmt.Errorf("claim assessment %q: confidence %g must be in [0,1]", a.ClaimID, c)
		}
	}
	return nil
}

// ValidateDimension returns nil when d is a well-formed dimension result.
func ValidateDimension(d *DimensionResult, allowedEvidence map[string]struct{}) error {
	if err := identifier.Validate(string(d.ConstructID)); err != nil {
		return fmt.Errorf("dimension: %w", err)
	}
	if !d.Applicable && d.Value != nil {
		return fmt.Errorf("dimension %q: not-applicable dimensions must not carry a value", d.ConstructID)
	}
	if d.Value != nil {
		v := *d.Value
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("dimension %q: value is not finite", d.ConstructID)
		}
	}
	if d.Confidence != nil {
		c := *d.Confidence
		if math.IsNaN(c) || math.IsInf(c, 0) || c < 0 || c > 1 {
			return fmt.Errorf("dimension %q: confidence %g must be in [0,1]", d.ConstructID, c)
		}
	}
	seen := make(map[string]struct{}, len(d.EvidenceIDs))
	for _, eid := range d.EvidenceIDs {
		if strings.TrimSpace(eid) == "" {
			return fmt.Errorf("dimension %q: blank evidence id", d.ConstructID)
		}
		if _, dup := seen[eid]; dup {
			return fmt.Errorf("dimension %q: duplicate evidence id %q", d.ConstructID, eid)
		}
		seen[eid] = struct{}{}
		if allowedEvidence != nil {
			if _, ok := allowedEvidence[eid]; !ok {
				return fmt.Errorf("dimension %q: cites unknown evidence %q", d.ConstructID, eid)
			}
		}
	}
	return nil
}

// EvidenceIDSet builds the allowed-evidence set from an evidence set's items.
// The judging layer uses it to cross-check verdict evidence references.
func EvidenceIDSet(items []eval.EvidenceItem) map[string]struct{} {
	set := make(map[string]struct{}, len(items))
	for i := range items {
		set[items[i].ID] = struct{}{}
	}
	return set
}

// ValidateReport returns nil when r is a well-formed, sealed report. It checks
// structural integrity and cross-references; evidence references are checked
// against allowedEvidence when it is non-nil.
func ValidateReport(r *Report, allowedEvidence map[string]struct{}) error {
	if err := validateReportBody(r, allowedEvidence); err != nil {
		return err
	}
	if !strings.HasPrefix(r.Digest, "sha256:") {
		return fmt.Errorf("report: digest must be a sha256: digest")
	}
	return nil
}

// validateReportBody performs all report checks except the digest-presence
// check, so Seal can validate structure before the digest is computed.
func validateReportBody(r *Report, allowedEvidence map[string]struct{}) error {
	if r.APIVersion != ReportAPIVersion {
		return fmt.Errorf("report: api_version %q is not supported (want %s)", r.APIVersion, ReportAPIVersion)
	}
	if err := identifier.Validate(r.InstanceID); err != nil {
		return fmt.Errorf("report: instance id: %w", err)
	}
	if !strings.HasPrefix(r.InstanceDigest, "sha256:") {
		return fmt.Errorf("report: instance_digest must be a sha256: digest")
	}
	if !strings.HasPrefix(r.ProtocolDigest, "sha256:") {
		return fmt.Errorf("report: protocol_digest must be a sha256: digest")
	}
	if r.FinishedAt.Before(r.StartedAt) {
		return fmt.Errorf("report: finished_at is before started_at")
	}
	// Claims: unique IDs.
	claimIDs := make(map[string]struct{}, len(r.Claims))
	for i := range r.Claims {
		c := &r.Claims[i]
		if err := ValidateClaim(c); err != nil {
			return err
		}
		if _, dup := claimIDs[c.ID]; dup {
			return fmt.Errorf("report: duplicate claim id %q", c.ID)
		}
		claimIDs[c.ID] = struct{}{}
	}
	// Claim results: each references a real claim, no duplicates, and every
	// claim has exactly one result.
	resultByClaim := make(map[string]int, len(r.ClaimResults))
	for i := range r.ClaimResults {
		a := &r.ClaimResults[i]
		if err := ValidateClaimAssessment(a, allowedEvidence); err != nil {
			return err
		}
		if _, ok := claimIDs[a.ClaimID]; !ok {
			return fmt.Errorf("report: claim result references unknown claim %q", a.ClaimID)
		}
		if _, dup := resultByClaim[a.ClaimID]; dup {
			return fmt.Errorf("report: duplicate claim result for claim %q", a.ClaimID)
		}
		resultByClaim[a.ClaimID] = i
	}
	if len(resultByClaim) != len(claimIDs) {
		return fmt.Errorf("report: %d claims but %d claim results (must be one result per claim)", len(claimIDs), len(resultByClaim))
	}
	// Dimensions: unique construct IDs.
	constructIDs := make(map[string]struct{}, len(r.Dimensions))
	for i := range r.Dimensions {
		d := &r.Dimensions[i]
		if err := ValidateDimension(d, allowedEvidence); err != nil {
			return err
		}
		if _, dup := constructIDs[string(d.ConstructID)]; dup {
			return fmt.Errorf("report: duplicate dimension for construct %q", d.ConstructID)
		}
		constructIDs[string(d.ConstructID)] = struct{}{}
	}
	// Raw artifacts.
	for i := range r.RawArtifacts {
		if err := eval.ValidateArtifact(&r.RawArtifacts[i]); err != nil {
			return fmt.Errorf("report: raw artifact %d: %w", i, err)
		}
	}
	return nil
}
