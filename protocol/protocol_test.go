package protocol

import (
	"path/filepath"
	"strings"
	"testing"
)

func ptrFloat(f float64) *float64 { return &f }
func ptrInt(i int64) *int64       { return &i }

func baseProtocol() Protocol {
	return Protocol{
		APIVersion:        ProtocolAPIVersion,
		Name:              "gec-rag-faithfulness-v2",
		MeasurementDigest: "sha256:abc",
		Model:             ModelIdentity{Provider: "geppetto", Model: "gpt-5.6-luna"},
		PromptDigests:     map[string]string{"extract": "sha256:1", "support": "sha256:2"},
		Decoding:          DecodingPolicy{MaxTokens: 1024, Temperature: ptrFloat(0), Seed: ptrInt(42)},
		EvidenceOrder:     EvidenceOrderAsGiven,
		ParserVersion:     "strict-json-v1",
		AggregatorVersion: "faithfulness-fraction-v1",
		Retry:             RetryPolicy{MaximumAttempts: 2},
	}
}

func TestValidateAcceptsBase(t *testing.T) {
	p := baseProtocol()
	if err := Validate(&p); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestLoadProtocolYAML(t *testing.T) {
	doc, err := LoadProtocol(filepath.Join("testdata", "faithfulness.yaml"))
	if err != nil {
		t.Fatalf("LoadProtocol: %v", err)
	}
	if !strings.HasPrefix(doc.Digest, "sha256:") {
		t.Errorf("digest = %q", doc.Digest)
	}
	if doc.Digest == doc.ByteDigest {
		t.Errorf("semantic and byte digests must differ")
	}
}

func TestRejectsUnsupportedVersion(t *testing.T) {
	p := baseProtocol()
	p.APIVersion = "judgekit.protocol/v9"
	if err := Validate(&p); err == nil {
		t.Errorf("accepted unsupported api_version")
	}
}

func TestRejectsBadName(t *testing.T) {
	p := baseProtocol()
	p.Name = "Bad Name"
	if err := Validate(&p); err == nil {
		t.Errorf("accepted invalid protocol name")
	}
}

func TestRejectsMissingPromptDigests(t *testing.T) {
	p := baseProtocol()
	p.PromptDigests = map[string]string{}
	if err := Validate(&p); err == nil {
		t.Errorf("accepted protocol with no prompt digests")
	}
}

func TestRejectsShuffledWithoutSeed(t *testing.T) {
	p := baseProtocol()
	p.EvidenceOrder = EvidenceOrderShuffled
	p.Decoding.Seed = nil
	if err := Validate(&p); err == nil {
		t.Errorf("accepted shuffled evidence order without seed")
	}
}

func TestRejectsBadDecoding(t *testing.T) {
	p := baseProtocol()
	p.Decoding.MaxTokens = 0
	if err := Validate(&p); err == nil {
		t.Errorf("accepted max_tokens=0")
	}
	p.Decoding.MaxTokens = 10
	p.Decoding.Temperature = ptrFloat(3)
	if err := Validate(&p); err == nil {
		t.Errorf("accepted temperature=3")
	}
}

func TestRejectsBadRetry(t *testing.T) {
	p := baseProtocol()
	p.Retry.MaximumAttempts = 0
	if err := Validate(&p); err == nil {
		t.Errorf("accepted maximum_attempts=0")
	}
}

func TestSemanticDigestChangesOnSemanticField(t *testing.T) {
	a := baseProtocol()
	b := a
	d1, err := SemanticDigest(&a)
	if err != nil {
		t.Fatalf("digest a: %v", err)
	}
	// Change a semantically relevant field: model revision.
	b.Model.Revision = "2026-09-01"
	d2, err := SemanticDigest(&b)
	if err != nil {
		t.Fatalf("digest b: %v", err)
	}
	if d1 == d2 {
		t.Errorf("digest did not change when model revision changed")
	}
	// Change prompt order in the map: digest must be stable (map keys sorted).
	c := baseProtocol()
	c.PromptDigests = map[string]string{"support": "sha256:2", "extract": "sha256:1"}
	d3, err := SemanticDigest(&c)
	if err != nil {
		t.Fatalf("digest c: %v", err)
	}
	if d1 != d3 {
		t.Errorf("digest changed with prompt map insertion order")
	}
}

func TestLoadRejectsUnknownField(t *testing.T) {
	raw := []byte("api_version: judgekit.protocol/v1\nname: p\nbogus: 1\n")
	if _, err := LoadProtocolFromBytes("bad.yaml", raw); err == nil {
		t.Errorf("accepted unknown field")
	}
}
