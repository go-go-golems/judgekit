package assessment

import (
	"time"

	"github.com/go-go-golems/judgekit/eval"
)

// ReportAPIVersion is the only assessment report API version judgekit accepts.
const ReportAPIVersion = "judgekit.assessment/v1"

// Report is the sealed, content-addressed output of one evaluator over one
// instance. It carries the instance and protocol digests it was produced
// under, the extracted claims, per-claim verdicts, per-construct dimensions,
// optional raw artifacts, timing, and its own digest.
type Report struct {
	APIVersion     string            `json:"api_version" yaml:"api_version"`
	InstanceID     string            `json:"instance_id" yaml:"instance_id"`
	InstanceDigest string            `json:"instance_digest" yaml:"instance_digest"`
	ProtocolDigest string            `json:"protocol_digest" yaml:"protocol_digest"`
	Claims         []Claim           `json:"claims,omitempty" yaml:"claims,omitempty"`
	ClaimResults   []ClaimAssessment `json:"claim_results,omitempty" yaml:"claim_results,omitempty"`
	Dimensions     []DimensionResult `json:"dimensions" yaml:"dimensions"`
	RawArtifacts   []eval.Artifact   `json:"raw_artifacts,omitempty" yaml:"raw_artifacts,omitempty"`
	Provenance     RunProvenance     `json:"provenance" yaml:"provenance"`
	StartedAt      time.Time         `json:"started_at" yaml:"started_at"`
	FinishedAt     time.Time         `json:"finished_at" yaml:"finished_at"`
	Digest         string            `json:"digest" yaml:"digest"`
}
