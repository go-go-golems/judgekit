package eval

import "testing"

func TestBindCurrentIdentityRecomputesStaleNestedDigests(t *testing.T) {
	set, err := NewEvidenceSet([]EvidenceItem{{
		ID: "e1", Kind: "knowledge", Content: NewTextArtifact("text/plain", "before"), SourceID: "source-1",
	}}, "sha256:policy")
	if err != nil {
		t.Fatalf("evidence set: %v", err)
	}
	instance, err := NewInstance("instance-1", NewTextArtifact("text/plain", "question"), NewTextArtifact("text/plain", "answer"), set, nil, nil, map[string]string{"arm": "before"})
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	oldInstanceDigest := instance.Digest
	oldEvidenceDigest := instance.Evidence.Digest

	instance.Metadata["arm"] = "after"
	instance.Evidence.Items[0].Content = NewTextArtifact("text/plain", "after")
	if err := ValidateInstance(&instance); err == nil {
		t.Fatal("stale instance unexpectedly validated")
	}

	bound, err := BindCurrentIdentity(instance)
	if err != nil {
		t.Fatalf("bind current identity: %v", err)
	}
	if bound.Digest == oldInstanceDigest {
		t.Fatal("instance digest did not change")
	}
	if bound.Evidence.Digest == oldEvidenceDigest {
		t.Fatal("evidence digest did not change")
	}
	if err := ValidateInstance(&bound); err != nil {
		t.Fatalf("bound instance: %v", err)
	}
}
