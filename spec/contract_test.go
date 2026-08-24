package spec

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadContractYAML(t *testing.T) {
	doc, err := LoadContract(filepath.Join("testdata", "faithfulness.yaml"))
	if err != nil {
		t.Fatalf("LoadContract: %v", err)
	}
	if doc.Contract.Name != "rag-faithfulness" {
		t.Errorf("name = %q, want rag-faithfulness", doc.Contract.Name)
	}
	if doc.Digest == "" || !strings.HasPrefix(doc.Digest, "sha256:") {
		t.Errorf("digest = %q, want sha256:...", doc.Digest)
	}
	if doc.ByteDigest == "" || !strings.HasPrefix(doc.ByteDigest, "sha256:") {
		t.Errorf("byte_digest = %q, want sha256:...", doc.ByteDigest)
	}
	if doc.Digest == doc.ByteDigest {
		t.Errorf("semantic and byte digests must not be equal")
	}
	if len(doc.Contract.Constructs) != 3 {
		t.Errorf("constructs = %d, want 3", len(doc.Contract.Constructs))
	}
}

func TestLoadContractRejectsUnknownField(t *testing.T) {
	raw := []byte("api_version: judgekit.measurement/v1\nname: bad\nconstructs: []\nbogus: 1\n")
	if _, err := LoadContractFromBytes("bad.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted unknown field, want error")
	}
}

func TestLoadContractRejectsUnsupportedVersion(t *testing.T) {
	raw := []byte("api_version: judgekit.measurement/v9\nname: bad\nconstructs: []\n")
	if _, err := LoadContractFromBytes("bad.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted unsupported api_version, want error")
	}
}

func TestLoadContractRejectsDuplicateConstruct(t *testing.T) {
	raw := []byte(`api_version: judgekit.measurement/v1
name: dup
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
  - id: faithfulness
    name: B
    definition: d
    unit: fraction
    direction: maximize
labels:
  faithfulness: [supported]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: supported
    empty_policy: vacuous_perfect
`)
	if _, err := LoadContractFromBytes("dup.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted duplicate construct id, want error")
	}
}

func TestLoadContractRejectsDanglingLabel(t *testing.T) {
	raw := []byte(`api_version: judgekit.measurement/v1
name: dangle
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
labels:
  missing: [supported]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: supported
    empty_policy: vacuous_perfect
`)
	if _, err := LoadContractFromBytes("dangle.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted labels for unknown construct, want error")
	}
}

func TestLoadContractRejectsMissingAggregation(t *testing.T) {
	raw := []byte(`api_version: judgekit.measurement/v1
name: missing-agg
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
labels:
  faithfulness: [supported]
aggregations: {}
`)
	if _, err := LoadContractFromBytes("missing.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted construct without aggregation, want error")
	}
}

func TestLoadContractRejectsFractionUndeclaredLabel(t *testing.T) {
	raw := []byte(`api_version: judgekit.measurement/v1
name: badfrac
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
labels:
  faithfulness: [supported]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: bogus
    empty_policy: vacuous_perfect
`)
	if _, err := LoadContractFromBytes("badfrac.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted fraction with undeclared denominator label, want error")
	}
}

func TestLoadContractRejectsEvidenceOverlap(t *testing.T) {
	raw := []byte(`api_version: judgekit.measurement/v1
name: overlap
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
evidence_policy:
  allowed_kinds: [knowledge, sql]
  forbidden_kinds: [sql]
labels:
  faithfulness: [supported]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: supported
    empty_policy: vacuous_perfect
`)
	if _, err := LoadContractFromBytes("overlap.yaml", raw); err == nil {
		t.Errorf("LoadContract accepted overlapping evidence kinds, want error")
	}
}

func TestSemanticDigestStableAcrossKeyOrder(t *testing.T) {
	a := []byte(`api_version: judgekit.measurement/v1
name: same
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
labels:
  faithfulness: [supported]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: supported
    empty_policy: vacuous_perfect
evidence_policy:
  allowed_kinds: [knowledge]
`)
	b := []byte(`api_version: judgekit.measurement/v1
name: same
evidence_policy:
  allowed_kinds: [knowledge]
aggregations:
  faithfulness:
    method: fraction
    numerator: supported
    denominator: supported
    empty_policy: vacuous_perfect
labels:
  faithfulness: [supported]
constructs:
  - id: faithfulness
    name: A
    definition: d
    unit: fraction
    direction: maximize
`)
	da, err := LoadContractFromBytes("a.yaml", a)
	if err != nil {
		t.Fatalf("a: %v", err)
	}
	db, err := LoadContractFromBytes("b.yaml", b)
	if err != nil {
		t.Fatalf("b: %v", err)
	}
	if da.Digest != db.Digest {
		t.Errorf("semantic digest changed with key order:\n  %s\n  %s", da.Digest, db.Digest)
	}
	if da.ByteDigest == db.ByteDigest {
		t.Errorf("byte digest should differ when bytes differ")
	}
}

func TestValidateConstructRejectsBadID(t *testing.T) {
	c := &Construct{ID: "Bad ID", Name: "x", Definition: "d", Unit: "u", Direction: Maximize}
	if err := ValidateConstruct(c); err == nil {
		t.Errorf("ValidateConstruct accepted bad id, want error")
	}
}

func TestValidateConstructRejectsBadRange(t *testing.T) {
	c := &Construct{ID: "x", Name: "x", Definition: "d", Unit: "u", Direction: Maximize, Range: &Range{Minimum: 1, Maximum: 0}}
	if err := ValidateConstruct(c); err == nil {
		t.Errorf("ValidateConstruct accepted min>max, want error")
	}
}
