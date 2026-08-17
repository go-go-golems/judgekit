package calibration

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/internal/identifier"
	"github.com/go-go-golems/judgekit/spec"
)

// GoldClaim is one human-labeled claim: the claim text, the gold support label,
// the reviewers who produced it, and whether their disagreement was adjudicated.
// ReviewerIDs are retained even after adjudication so inter-rater agreement can
// be measured; adjudication must not erase original disagreement.
type GoldClaim struct {
	InstanceID  string                 `json:"instance_id" yaml:"instance_id"`
	Claim       assessment.Claim       `json:"claim" yaml:"claim"`
	Label       assessment.SupportLabel `json:"label" yaml:"label"`
	ReviewerIDs []string               `json:"reviewer_ids" yaml:"reviewer_ids"`
	Adjudicated bool                   `json:"adjudicated" yaml:"adjudicated"`
}

// GoldDimension is one human-labeled dimension value or label for one instance
// and construct, with the reviewers who produced it.
type GoldDimension struct {
	InstanceID  string            `json:"instance_id" yaml:"instance_id"`
	ConstructID spec.ConstructID  `json:"construct_id" yaml:"construct_id"`
	Value       *float64          `json:"value,omitempty" yaml:"value,omitempty"`
	Label       string            `json:"label,omitempty" yaml:"label,omitempty"`
	ReviewerIDs []string          `json:"reviewer_ids" yaml:"reviewer_ids"`
}

// GoldSet is a collection of gold claims and dimensions for a calibration
// dataset, content-addressed so a calibration report can pin the exact labels
// it was computed against.
type GoldSet struct {
	Claims     []GoldClaim    `json:"claims" yaml:"claims"`
	Dimensions []GoldDimension `json:"dimensions,omitempty" yaml:"dimensions,omitempty"`
	Digest     string         `json:"digest" yaml:"digest"`
}

// ValidateGoldClaim returns nil when g is a well-formed gold claim.
func ValidateGoldClaim(g *GoldClaim) error {
	if err := identifier.Validate(g.InstanceID); err != nil {
		return fmt.Errorf("gold claim: instance id: %w", err)
	}
	if err := assessment.ValidateClaim(&g.Claim); err != nil {
		return fmt.Errorf("gold claim for instance %q: %w", g.InstanceID, err)
	}
	if !assessment.ValidSupportLabel(g.Label) {
		return fmt.Errorf("gold claim %q: label %q is not a support label", g.Claim.ID, g.Label)
	}
	if len(g.ReviewerIDs) == 0 {
		return fmt.Errorf("gold claim %q: at least one reviewer is required", g.Claim.ID)
	}
	seen := make(map[string]struct{}, len(g.ReviewerIDs))
	for _, r := range g.ReviewerIDs {
		if strings.TrimSpace(r) == "" {
			return fmt.Errorf("gold claim %q: blank reviewer id", g.Claim.ID)
		}
		if _, dup := seen[r]; dup {
			return fmt.Errorf("gold claim %q: duplicate reviewer %q", g.Claim.ID, r)
		}
		seen[r] = struct{}{}
	}
	return nil
}

// ValidateGoldDimension returns nil when g is a well-formed gold dimension.
func ValidateGoldDimension(g *GoldDimension) error {
	if err := identifier.Validate(g.InstanceID); err != nil {
		return fmt.Errorf("gold dimension: instance id: %w", err)
	}
	if err := identifier.Validate(string(g.ConstructID)); err != nil {
		return fmt.Errorf("gold dimension: construct id: %w", err)
	}
	if g.Value == nil && strings.TrimSpace(g.Label) == "" {
		return fmt.Errorf("gold dimension for instance %q construct %q: value or label is required", g.InstanceID, g.ConstructID)
	}
	if len(g.ReviewerIDs) == 0 {
		return fmt.Errorf("gold dimension for instance %q construct %q: at least one reviewer is required", g.InstanceID, g.ConstructID)
	}
	return nil
}

// ValidateGoldSet returns nil when s is a well-formed gold set: every record is
// valid and claim IDs are unique within the set.
func ValidateGoldSet(s *GoldSet) error {
	seen := make(map[string]struct{}, len(s.Claims))
	for i := range s.Claims {
		if err := ValidateGoldClaim(&s.Claims[i]); err != nil {
			return err
		}
		key := s.Claims[i].InstanceID + "/" + s.Claims[i].Claim.ID
		if _, dup := seen[key]; dup {
			return fmt.Errorf("gold set: duplicate claim %q for instance %q", s.Claims[i].Claim.ID, s.Claims[i].InstanceID)
		}
		seen[key] = struct{}{}
	}
	for i := range s.Dimensions {
		if err := ValidateGoldDimension(&s.Dimensions[i]); err != nil {
			return err
		}
	}
	if !strings.HasPrefix(s.Digest, "sha256:") {
		return fmt.Errorf("gold set: digest must be a sha256: digest")
	}
	return nil
}
