package protocol

// ProtocolAPIVersion is the only protocol API version judgekit accepts.
const ProtocolAPIVersion = "judgekit.protocol/v1"

// ModelIdentity names the evaluator model and provider without exposing
// provider-specific request objects. Revision and Settings let an application
// pin a checkpoint or serving configuration so "gpt-5.6-luna" is not ambiguous.
type ModelIdentity struct {
	Provider string            `json:"provider" yaml:"provider"`
	Model    string            `json:"model" yaml:"model"`
	Revision string            `json:"revision,omitempty" yaml:"revision,omitempty"`
	Settings map[string]string `json:"settings,omitempty" yaml:"settings,omitempty"`
}

// DecodingPolicy captures the sampling parameters that affect output. All
// pointers are optional; a nil field means "provider default".
type DecodingPolicy struct {
	Temperature *float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty" yaml:"top_p,omitempty"`
	MaxTokens   int      `json:"max_tokens" yaml:"max_tokens"`
	Seed        *int64   `json:"seed,omitempty" yaml:"seed,omitempty"`
}

// RetryPolicy bounds repair attempts. Only structural failures should be
// repaired by default; MaximumAttempts includes the first attempt.
type RetryPolicy struct {
	MaximumAttempts int      `json:"maximum_attempts" yaml:"maximum_attempts"`
	RepairKinds     []string `json:"repair_kinds,omitempty" yaml:"repair_kinds,omitempty"`
}

// Evidence ordering policies.
const (
	// EvidenceOrderAsGiven presents evidence in the order of the evidence set.
	EvidenceOrderAsGiven = "as_given"
	// EvidenceOrderSorted sorts evidence by ID before presentation.
	EvidenceOrderSorted = "sorted"
	// EvidenceOrderShuffled shuffles evidence using the decoding seed so order
	// randomization is reproducible.
	EvidenceOrderShuffled = "shuffled"
)

// validEvidenceOrders is the set of accepted evidence orderings.
var validEvidenceOrders = map[string]bool{
	EvidenceOrderAsGiven:  true,
	EvidenceOrderSorted:   true,
	EvidenceOrderShuffled: true,
}

// Protocol is the complete reproducible instrument configuration.
type Protocol struct {
	APIVersion        string            `json:"api_version" yaml:"api_version"`
	Name              string            `json:"name" yaml:"name"`
	MeasurementDigest string            `json:"measurement_digest" yaml:"measurement_digest"`
	Model             ModelIdentity     `json:"model" yaml:"model"`
	PromptDigests     map[string]string `json:"prompt_digests" yaml:"prompt_digests"`
	Decoding          DecodingPolicy    `json:"decoding" yaml:"decoding"`
	EvidenceOrder     string            `json:"evidence_order" yaml:"evidence_order"`
	ParserVersion     string            `json:"parser_version" yaml:"parser_version"`
	AggregatorVersion string            `json:"aggregator_version" yaml:"aggregator_version"`
	Retry             RetryPolicy       `json:"retry" yaml:"retry"`
}

// Document pairs a validated protocol with its semantic and byte identities
// and the path it was loaded from.
type Document struct {
	Protocol   Protocol `json:"protocol" yaml:"protocol"`
	Digest     string   `json:"digest" yaml:"digest"`
	ByteDigest string   `json:"byte_digest" yaml:"byte_digest"`
	Path       string   `json:"path" yaml:"path"`
}
