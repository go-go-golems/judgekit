package spec

// ContractAPIVersion is the only measurement-contract API version judgekit
// accepts. Rejecting unknown versions fail-closes the loader so a future
// incompatible schema cannot be silently interpreted by old code.
const ContractAPIVersion = "judgekit.measurement/v1"

// ConstructID is a bounded portable identifier (see internal/identifier) that
// names a construct within a measurement contract. It is used as a map key
// for labels, aggregations, and exclusions, so it must be stable and canonical.
type ConstructID string

// Direction states how a construct's value should move to be considered
// better. Descriptive constructs carry no preferred direction (for example,
// a label distribution that is reported but not optimized).
type Direction string

const (
	// Maximize means higher values are better.
	Maximize Direction = "maximize"
	// Minimize means lower values are better.
	Minimize Direction = "minimize"
	// Descriptive means the value is reported without a preferred direction.
	Descriptive Direction = "descriptive"
)

// validDirections is the set of accepted Direction values.
var validDirections = map[Direction]bool{
	Maximize:    true,
	Minimize:    true,
	Descriptive: true,
}

// Range bounds the numeric value of a construct. Both bounds are inclusive.
// A nil Range means the construct is unbounded (for example, a descriptive
// label whose scale is defined by labels rather than numbers).
type Range struct {
	Minimum float64 `json:"minimum" yaml:"minimum"`
	Maximum float64 `json:"maximum" yaml:"maximum"`
}

// Construct is the abstract property an evaluation intends to measure, paired
// with its operational definition so it can be applied consistently.
type Construct struct {
	ID         ConstructID `json:"id" yaml:"id"`
	Name       string      `json:"name" yaml:"name"`
	Definition string      `json:"definition" yaml:"definition"`
	Unit       string      `json:"unit" yaml:"unit"`
	Direction  Direction   `json:"direction" yaml:"direction"`
	Range      *Range      `json:"range,omitempty" yaml:"range,omitempty"`
}
