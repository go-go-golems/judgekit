package spec

// EvidencePolicy restricts which evidence an evaluator may admit for a
// contract. Kinds are free-form strings the application defines (for example
// "knowledge", "sql", "reference"); judgekit does not interpret them, it only
// records and checks them so an evaluator cannot silently admit a forbidden
// source type.
type EvidencePolicy struct {
	AllowedKinds      []string `json:"allowed_kinds" yaml:"allowed_kinds"`
	ForbiddenKinds    []string `json:"forbidden_kinds" yaml:"forbidden_kinds"`
	RequireProvenance bool     `json:"require_provenance" yaml:"require_provenance"`
}

// Aggregation defines how per-claim or per-item results are combined into a
// dimension value. Method names the aggregation; for "fraction", Numerator
// and Denominator must name labels defined for the construct.
type Aggregation struct {
	Method      string `json:"method" yaml:"method"`
	Numerator   string `json:"numerator,omitempty" yaml:"numerator,omitempty"`
	Denominator string `json:"denominator,omitempty" yaml:"denominator,omitempty"`
	EmptyPolicy string `json:"empty_policy" yaml:"empty_policy"`
}

// Empty-case policies describe the value reported when an aggregation has no
// items (for example, an answer that makes zero factual claims).
const (
	// EmptyVacuousPerfect reports the maximum value when there is nothing to
	// penalize (e.g. faithfulness of an abstaining answer is 1.0).
	EmptyVacuousPerfect = "vacuous_perfect"
	// EmptyZero reports zero when an empty result is a failure.
	EmptyZero = "zero"
	// EmptyNA reports no value when the construct does not apply.
	EmptyNA = "na"
)

// validEmptyPolicies is the set of accepted EmptyPolicy values.
var validEmptyPolicies = map[string]bool{
	EmptyVacuousPerfect: true,
	EmptyZero:           true,
	EmptyNA:             true,
}

// Aggregation methods.
const (
	// MethodFraction computes Numerator/Denominator over labels.
	MethodFraction = "fraction"
	// MethodMean averages the values of a label across items.
	MethodMean = "mean"
	// MethodSum sums the values of a label across items.
	MethodSum = "sum"
	// MethodCount reports the count of items with a given label.
	MethodCount = "count"
	// MethodLabel reports a single label string rather than a number.
	MethodLabel = "label"
	// MethodDirect takes the dimension value from the judge's emitted
	// dimension for this construct rather than aggregating claim labels.
	MethodDirect = "direct"
)

// validMethods is the set of accepted Method values.
var validMethods = map[string]bool{
	MethodFraction: true,
	MethodMean:     true,
	MethodSum:      true,
	MethodCount:    true,
	MethodLabel:    true,
	MethodDirect:   true,
}

// MeasurementContract is the operational definition of one or more constructs:
// the constructs, evidence policy, labels, aggregations, and exclusions that
// make "quality" measurable instead of underspecified.
type MeasurementContract struct {
	APIVersion     string                      `json:"api_version" yaml:"api_version"`
	Name           string                      `json:"name" yaml:"name"`
	Constructs     []Construct                 `json:"constructs" yaml:"constructs"`
	EvidencePolicy EvidencePolicy              `json:"evidence_policy" yaml:"evidence_policy"`
	Labels         map[ConstructID][]string    `json:"labels" yaml:"labels"`
	Aggregations   map[ConstructID]Aggregation `json:"aggregations" yaml:"aggregations"`
	Exclusions     map[ConstructID][]string    `json:"exclusions,omitempty" yaml:"exclusions,omitempty"`
}

// ContractDocument pairs a validated contract with its semantic and byte
// identities and the path it was loaded from. The digests make a contract
// content-addressable so protocols and reports can pin the exact definition
// they were produced under.
type ContractDocument struct {
	Contract   MeasurementContract `json:"contract" yaml:"contract"`
	Digest     string              `json:"digest" yaml:"digest"`
	ByteDigest string              `json:"byte_digest" yaml:"byte_digest"`
	Path       string              `json:"path" yaml:"path"`
}
