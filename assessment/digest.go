package assessment

import (
	"fmt"
	"time"

	"github.com/go-go-golems/judgekit/internal/canonicaljson"
)

// reportDigestInput is Report without the Digest field, so the report digest
// is a non-circular function of report content.
type reportDigestInput struct {
	APIVersion     string            `json:"api_version"`
	InstanceID     string            `json:"instance_id"`
	InstanceDigest string            `json:"instance_digest"`
	ProtocolDigest string            `json:"protocol_digest"`
	Claims         []Claim           `json:"claims,omitempty"`
	ClaimResults   []ClaimAssessment `json:"claim_results,omitempty"`
	Dimensions     []DimensionResult `json:"dimensions"`
	RawArtifacts   []reportArtifact  `json:"raw_artifacts,omitempty"`
	Provenance     RunProvenance     `json:"provenance"`
	StartedAt      string            `json:"started_at"`
	FinishedAt     string            `json:"finished_at"`
}

// reportArtifact mirrors eval.Artifact for digest stability without importing
// a nested struct that already carries its own digest semantics.
type reportArtifact struct {
	MediaType string `json:"media_type"`
	Text      string `json:"text,omitempty"`
	URI       string `json:"uri,omitempty"`
	Digest    string `json:"digest"`
	SizeBytes int64  `json:"size_bytes"`
}

// ReportDigest returns the canonical digest of a report, excluding the report's
// own Digest field. It does not validate.
func ReportDigest(r *Report) (string, error) {
	input := reportDigestInput{
		APIVersion:     r.APIVersion,
		InstanceID:     r.InstanceID,
		InstanceDigest: r.InstanceDigest,
		ProtocolDigest: r.ProtocolDigest,
		Claims:         r.Claims,
		ClaimResults:   r.ClaimResults,
		Dimensions:     r.Dimensions,
		Provenance:     r.Provenance,
		StartedAt:      r.StartedAt.UTC().Format(time.RFC3339Nano),
		FinishedAt:     r.FinishedAt.UTC().Format(time.RFC3339Nano),
	}
	for _, a := range r.RawArtifacts {
		input.RawArtifacts = append(input.RawArtifacts, reportArtifact{
			MediaType: a.MediaType,
			Text:      a.Text,
			URI:       a.URI,
			Digest:    a.Digest,
			SizeBytes: a.SizeBytes,
		})
	}
	return canonicaljson.Sum(input)
}

// Seal validates a report's structure (without requiring a digest yet),
// computes its digest, and sets r.Digest. After Seal, ValidateReport passes.
// allowedEvidence, when non-nil, is used to cross-check evidence references.
func Seal(r *Report, allowedEvidence map[string]struct{}) error {
	if err := validateReportBody(r, allowedEvidence); err != nil {
		return err
	}
	digest, err := ReportDigest(r)
	if err != nil {
		return fmt.Errorf("seal report: %w", err)
	}
	r.Digest = digest
	return nil
}
