package eval

import (
	"github.com/go-go-golems/judgekit/internal/canonicaljson"
)

// evidenceDigestInput is the canonical input over which an evidence set's
// digest is computed. It deliberately excludes the Digest field so the digest
// is a function of content only.
type evidenceDigestInput struct {
	Items        []EvidenceItem `json:"items"`
	PolicyDigest string         `json:"policy_digest"`
}

// EvidenceSetDigest returns the canonical digest of an ordered evidence set
// under a policy digest. Item order is significant because the protocol may
// present evidence in a fixed order.
func EvidenceSetDigest(items []EvidenceItem, policyDigest string) (string, error) {
	return canonicaljson.Sum(evidenceDigestInput{Items: items, PolicyDigest: policyDigest})
}

// instanceDigestInput is Instance without the Digest field, used so the
// instance digest is a pure function of instance content.
type instanceDigestInput struct {
	ID            string            `json:"id"`
	Input         Artifact          `json:"input"`
	Candidate     Artifact          `json:"candidate"`
	Evidence      EvidenceSet       `json:"evidence"`
	Reference     *Artifact         `json:"reference,omitempty"`
	RequiredFacts []RequiredFact    `json:"required_facts,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// InstanceDigest returns the canonical digest of an instance, excluding the
// instance's own Digest field so the computation is non-circular.
func InstanceDigest(i *Instance) (string, error) {
	return canonicaljson.Sum(instanceDigestInput{
		ID:            i.ID,
		Input:         i.Input,
		Candidate:     i.Candidate,
		Evidence:      i.Evidence,
		Reference:     i.Reference,
		RequiredFacts: i.RequiredFacts,
		Metadata:      i.Metadata,
	})
}
