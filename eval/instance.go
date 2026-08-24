package eval

// RequiredFact is a fact the answer ought to establish, used to measure
// completeness independently of answer length. Importance is a 0..1 weight.
// EvidenceIDs optionally links the fact to the evidence that establishes it.
type RequiredFact struct {
	ID          string   `json:"id" yaml:"id"`
	Description string   `json:"description" yaml:"description"`
	Importance  float64  `json:"importance" yaml:"importance"`
	EvidenceIDs []string `json:"evidence_ids,omitempty" yaml:"evidence_ids,omitempty"`
}

// Instance is one concrete item being judged: the input, the candidate, the
// admitted evidence, an optional reference, optional required facts, and
// metadata. Digest makes the instance content-addressable so reports and
// caches can pin exact inputs.
type Instance struct {
	ID            string            `json:"id" yaml:"id"`
	Input         Artifact          `json:"input" yaml:"input"`
	Candidate     Artifact          `json:"candidate" yaml:"candidate"`
	Evidence      EvidenceSet       `json:"evidence" yaml:"evidence"`
	Reference     *Artifact         `json:"reference,omitempty" yaml:"reference,omitempty"`
	RequiredFacts []RequiredFact    `json:"required_facts,omitempty" yaml:"required_facts,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Digest        string            `json:"digest" yaml:"digest"`
}
