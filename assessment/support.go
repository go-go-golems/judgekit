package assessment

// SupportLabel is the three-way verdict a support judge assigns to a claim
// against evidence. A boolean cannot distinguish contradiction from absent
// evidence, and those failures need different interventions.
type SupportLabel string

const (
	// Entailed means the evidence entails the claim.
	Entailed SupportLabel = "entailed"
	// Contradicted means the evidence contradicts the claim.
	Contradicted SupportLabel = "contradicted"
	// Insufficient means the evidence neither entails nor contradicts the
	// claim; there is not enough evidence to decide.
	Insufficient SupportLabel = "insufficient"
)

// validSupportLabels is the set of accepted support labels.
var validSupportLabels = map[SupportLabel]bool{
	Entailed:     true,
	Contradicted: true,
	Insufficient: true,
}

// ValidSupportLabel reports whether label is one of the accepted three-way
// support labels. It is exported so adjacent packages (calibration, audit) can
// validate labels without re-declaring the set.
func ValidSupportLabel(label SupportLabel) bool {
	return validSupportLabels[label]
}

// ClaimAssessment is the verdict for one claim: its support label, the
// evidence cited, an optional confidence, and a required reason. Entailed and
// contradicted verdicts must cite at least one evidence item; only
// insufficient may cite none.
type ClaimAssessment struct {
	ClaimID     string       `json:"claim_id" yaml:"claim_id"`
	Label       SupportLabel `json:"label" yaml:"label"`
	EvidenceIDs []string     `json:"evidence_ids" yaml:"evidence_ids"`
	Confidence  *float64     `json:"confidence,omitempty" yaml:"confidence,omitempty"`
	Reason      string       `json:"reason" yaml:"reason"`
}
