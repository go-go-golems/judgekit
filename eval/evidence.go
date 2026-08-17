package eval

import "time"

// EvidenceItem is one piece of evidence an evaluator may consult. Kind is a
// free-form string the application defines (for example "knowledge", "sql",
// "reference"); judgekit does not interpret it. Provenance records where the
// evidence came from so an evidence policy that requires provenance can be
// enforced.
type EvidenceItem struct {
	ID         string            `json:"id" yaml:"id"`
	Kind       string            `json:"kind" yaml:"kind"`
	Content    Artifact          `json:"content" yaml:"content"`
	SourceID   string            `json:"source_id" yaml:"source_id"`
	SourceTime *time.Time        `json:"source_time,omitempty" yaml:"source_time,omitempty"`
	Authority  string            `json:"authority,omitempty" yaml:"authority,omitempty"`
	Provenance map[string]string `json:"provenance" yaml:"provenance"`
}

// EvidenceSet is the ordered collection of evidence admitted for one instance.
// PolicyDigest pins the evidence policy under which the set was admitted, and
// Digest makes the ordered set content-addressable.
type EvidenceSet struct {
	Items        []EvidenceItem `json:"items" yaml:"items"`
	PolicyDigest string         `json:"policy_digest" yaml:"policy_digest"`
	Digest       string         `json:"digest" yaml:"digest"`
}
