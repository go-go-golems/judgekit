package eval

import (
	"strings"
	"testing"
)

func TestNewTextArtifactContentAddressed(t *testing.T) {
	a := NewTextArtifact("text/plain", "hello")
	if err := ValidateArtifact(&a); err != nil {
		t.Fatalf("ValidateArtifact: %v", err)
	}
	if a.Digest != TextContentDigest("hello") {
		t.Errorf("digest mismatch")
	}
	if a.SizeBytes != 5 {
		t.Errorf("size = %d, want 5", a.SizeBytes)
	}
}

func TestValidateArtifactRejectsBothOrNeither(t *testing.T) {
	if err := ValidateArtifact(&Artifact{MediaType: "text/plain", Digest: "sha256:x"}); err == nil {
		t.Errorf("accepted artifact with neither text nor uri")
	}
	if err := ValidateArtifact(&Artifact{MediaType: "text/plain", Text: "a", URI: "b", Digest: "sha256:x"}); err == nil {
		t.Errorf("accepted artifact with both text and uri")
	}
}

func TestValidateArtifactRejectsStaleDigest(t *testing.T) {
	a := NewTextArtifact("text/plain", "hello")
	a.Text = "world"
	if err := ValidateArtifact(&a); err == nil {
		t.Errorf("accepted artifact whose digest does not match its text")
	}
}

func TestNewEvidenceSet(t *testing.T) {
	items := []EvidenceItem{
		{
			ID:       "e1",
			Kind:     "knowledge",
			Content:  NewTextArtifact("text/plain", "evidence one"),
			SourceID: "doc-1",
		},
		{
			ID:       "e2",
			Kind:     "sql",
			Content:  NewTextArtifact("text/plain", "evidence two"),
			SourceID: "sql-1",
		},
	}
	set, err := NewEvidenceSet(items, "sha256:policy")
	if err != nil {
		t.Fatalf("NewEvidenceSet: %v", err)
	}
	if !strings.HasPrefix(set.Digest, "sha256:") {
		t.Errorf("set digest = %q", set.Digest)
	}
	if err := ValidateEvidenceSet(&set); err != nil {
		t.Errorf("ValidateEvidenceSet: %v", err)
	}
}

func TestNewEvidenceSetRejectsDuplicateIDs(t *testing.T) {
	items := []EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "a"), SourceID: "s"},
		{ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "b"), SourceID: "s"},
	}
	if _, err := NewEvidenceSet(items, "sha256:policy"); err == nil {
		t.Errorf("accepted duplicate evidence ids")
	}
}

func TestNewInstance(t *testing.T) {
	set, err := NewEvidenceSet([]EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "ev"), SourceID: "s"},
	}, "sha256:policy")
	if err != nil {
		t.Fatalf("NewEvidenceSet: %v", err)
	}
	inst, err := NewInstance(
		"inst-1",
		NewTextArtifact("text/plain", "what is the policy?"),
		NewTextArtifact("text/plain", "the answer"),
		set,
		nil,
		[]RequiredFact{{ID: "f1", Description: "must mention the limit", Importance: 1, EvidenceIDs: []string{"e1"}}},
		map[string]string{"stratum": "baseline"},
	)
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	if !strings.HasPrefix(inst.Digest, "sha256:") {
		t.Errorf("instance digest = %q", inst.Digest)
	}
	if err := ValidateInstance(&inst); err != nil {
		t.Errorf("ValidateInstance: %v", err)
	}
}

func TestNewInstanceRejectsDanglingFactEvidence(t *testing.T) {
	set, _ := NewEvidenceSet([]EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "ev"), SourceID: "s"},
	}, "sha256:policy")
	_, err := NewInstance(
		"inst-1",
		NewTextArtifact("text/plain", "q"),
		NewTextArtifact("text/plain", "a"),
		set,
		nil,
		[]RequiredFact{{ID: "f1", Description: "d", Importance: 1, EvidenceIDs: []string{"missing"}}},
		nil,
	)
	if err == nil {
		t.Errorf("accepted required fact referencing unknown evidence")
	}
}

func TestNewInstanceRejectsBadID(t *testing.T) {
	set, _ := NewEvidenceSet(nil, "sha256:policy")
	_, err := NewInstance("Bad ID", NewTextArtifact("text/plain", "q"), NewTextArtifact("text/plain", "a"), set, nil, nil, nil)
	if err == nil {
		t.Errorf("accepted invalid instance id")
	}
}

func TestInstanceDigestDeterministic(t *testing.T) {
	set, _ := NewEvidenceSet([]EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "ev"), SourceID: "s"},
	}, "sha256:policy")
	makeInst := func() Instance {
		inst, err := NewInstance("inst-1", NewTextArtifact("text/plain", "q"), NewTextArtifact("text/plain", "a"), set, nil, nil, nil)
		if err != nil {
			t.Fatalf("NewInstance: %v", err)
		}
		return inst
	}
	a := makeInst()
	b := makeInst()
	if a.Digest != b.Digest {
		t.Errorf("instance digest not deterministic: %s vs %s", a.Digest, b.Digest)
	}
	// Changing metadata must change the digest.
	c := makeInst()
	c2, _ := NewInstance("inst-1", NewTextArtifact("text/plain", "q"), NewTextArtifact("text/plain", "a"), set, nil, nil, map[string]string{"k": "v"})
	if c.Digest == c2.Digest {
		t.Errorf("instance digest did not change when metadata changed")
	}
}
