package spec

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/go-go-golems/judgekit/internal/canonicaljson"
)

// SemanticDigest returns the canonical semantic identity of a validated
// contract: "sha256:" plus the hex digest of its canonical JSON. It is stable
// across harmless YAML key ordering and struct field declaration order.
//
// SemanticDigest does not validate; callers must run ValidateContract first.
func SemanticDigest(c *MeasurementContract) (string, error) {
	return canonicaljson.Sum(c)
}

// EvidencePolicyDigest returns the semantic identity of one evidence policy.
// Evidence sets carry this digest so a judge can prove that their kinds and
// provenance were admitted under the active measurement contract.
func EvidencePolicyDigest(p *EvidencePolicy) (string, error) {
	return canonicaljson.Sum(p)
}

// ByteDigest returns "sha256:" plus the hex digest of the raw source bytes.
// Unlike SemanticDigest, it changes with any byte change, including
// whitespace and key order, so it proves which exact reviewed file was used.
func ByteDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}
