package assessment

// Span locates a claim within the candidate text. Start is inclusive and End
// is exclusive, in byte offsets.
type Span struct {
	Start int `json:"start" yaml:"start"`
	End   int `json:"end" yaml:"end"`
}

// Claim is one factual statement extracted from a candidate, to be verified
// independently against evidence. Importance is a 0..1 weight. RequiresEvidence
// records whether the contract requires the claim to be grounded.
type Claim struct {
	ID               string  `json:"id" yaml:"id"`
	Text             string  `json:"text" yaml:"text"`
	Importance       float64 `json:"importance" yaml:"importance"`
	RequiresEvidence bool    `json:"requires_evidence" yaml:"requires_evidence"`
	CandidateSpan    *Span   `json:"candidate_span,omitempty" yaml:"candidate_span,omitempty"`
}
