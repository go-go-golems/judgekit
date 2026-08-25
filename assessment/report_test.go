package assessment

import (
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/protocol"
	"github.com/go-go-golems/judgekit/spec"
)

func ptrFloat(f float64) *float64 { return &f }

func sampleReport() Report {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	return Report{
		APIVersion:     ReportAPIVersion,
		InstanceID:     "inst-1",
		InstanceDigest: "sha256:instance",
		ProtocolDigest: "sha256:protocol",
		Claims: []Claim{
			{ID: "c1", Text: "the limit is five days", Importance: 1, RequiresEvidence: true},
			{ID: "c2", Text: "all leave carries over", Importance: 1, RequiresEvidence: true},
		},
		ClaimResults: []ClaimAssessment{
			{ClaimID: "c1", Label: Entailed, EvidenceIDs: []string{"e1"}, Reason: "e1 states the limit"},
			{ClaimID: "c2", Label: Contradicted, EvidenceIDs: []string{"e1"}, Reason: "e1 limits to five days"},
		},
		Dimensions: []DimensionResult{
			{ConstructID: "faithfulness", Applicable: true, Value: ptrFloat(0.5), EvidenceIDs: []string{"e1"}},
			{ConstructID: "relevance", Applicable: true, Value: ptrFloat(0.9)},
		},
		Provenance: RunProvenance{
			ContractDigest:        "sha256:contract",
			ProtocolDigest:        "sha256:protocol",
			InstanceDigest:        "sha256:instance",
			PromptTemplateDigests: map[string]string{"judge": "sha256:template"},
			ExpectedModel:         protocol.ModelIdentity{Provider: "fake", Model: "fake-1"},
			CacheMode:             "use",
			Generations: []PromptExecution{{
				Step: "judge", Attempt: 1, RenderedPromptDigest: "sha256:rendered",
				ObservedModel: protocol.ModelIdentity{Provider: "fake", Model: "fake-1"},
			}},
		},
		StartedAt:  now,
		FinishedAt: now.Add(2 * time.Second),
	}
}

func TestSealAndValidate(t *testing.T) {
	r := sampleReport()
	allowed := map[string]struct{}{"e1": {}}
	if err := Seal(&r, allowed); err != nil {
		t.Fatalf("Seal: %v", err)
	}
	if !strings.HasPrefix(r.Digest, "sha256:") {
		t.Errorf("digest = %q", r.Digest)
	}
	if err := ValidateReport(&r, allowed); err != nil {
		t.Errorf("ValidateReport after seal: %v", err)
	}
}

func TestRejectsEntailedWithoutEvidence(t *testing.T) {
	r := sampleReport()
	r.ClaimResults[0].EvidenceIDs = nil
	if err := Seal(&r, map[string]struct{}{}); err == nil {
		t.Errorf("accepted entailed verdict without evidence")
	}
}

func TestRejectsInsufficientWithoutEvidenceIsAllowed(t *testing.T) {
	r := sampleReport()
	r.ClaimResults[0] = ClaimAssessment{ClaimID: "c1", Label: Insufficient, Reason: "no evidence"}
	if err := Seal(&r, map[string]struct{}{"e1": {}}); err != nil {
		t.Errorf("rejected insufficient verdict without evidence: %v", err)
	}
}

func TestRejectsUnknownEvidenceReference(t *testing.T) {
	r := sampleReport()
	r.ClaimResults[0].EvidenceIDs = []string{"missing"}
	if err := Seal(&r, map[string]struct{}{"e1": {}}); err == nil {
		t.Errorf("accepted verdict citing unknown evidence")
	}
}

func TestRejectsDuplicateClaim(t *testing.T) {
	r := sampleReport()
	r.Claims = append(r.Claims, Claim{ID: "c1", Text: "dup", Importance: 1})
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted duplicate claim id")
	}
}

func TestRejectsClaimResultForUnknownClaim(t *testing.T) {
	r := sampleReport()
	r.ClaimResults = append(r.ClaimResults, ClaimAssessment{ClaimID: "c9", Label: Insufficient, Reason: "x"})
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted claim result for unknown claim")
	}
}

func TestRejectsMissingClaimResult(t *testing.T) {
	r := sampleReport()
	r.ClaimResults = r.ClaimResults[:1]
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted report with a claim missing its result")
	}
}

func TestRejectsInvalidSpan(t *testing.T) {
	r := sampleReport()
	r.Claims[0].CandidateSpan = &Span{Start: 5, End: 5}
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted invalid span")
	}
}

func TestRejectsNotApplicableWithValue(t *testing.T) {
	r := sampleReport()
	r.Dimensions = append(r.Dimensions, DimensionResult{ConstructID: "abstention", Applicable: false, Value: ptrFloat(0.5)})
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted not-applicable dimension with a value")
	}
}

func TestRejectsUnsupportedAPIVersion(t *testing.T) {
	r := sampleReport()
	r.APIVersion = "judgekit.assessment/v9"
	if err := Seal(&r, nil); err == nil {
		t.Errorf("accepted unsupported api_version")
	}
}

func TestReportDigestDeterministic(t *testing.T) {
	r1 := sampleReport()
	r2 := sampleReport()
	if err := Seal(&r1, nil); err != nil {
		t.Fatalf("seal r1: %v", err)
	}
	if err := Seal(&r2, nil); err != nil {
		t.Fatalf("seal r2: %v", err)
	}
	if r1.Digest != r2.Digest {
		t.Errorf("report digest not deterministic: %s vs %s", r1.Digest, r2.Digest)
	}
	// A semantic change must change the digest.
	r2.ClaimResults[0].Label = Insufficient
	r2.ClaimResults[0].EvidenceIDs = nil
	if err := Seal(&r2, nil); err != nil {
		t.Fatalf("seal r2 after change: %v", err)
	}
	if r1.Digest == r2.Digest {
		t.Errorf("report digest did not change when a verdict changed")
	}
}

func TestEvidenceIDSetFromInstance(t *testing.T) {
	set, _ := eval.NewEvidenceSet([]eval.EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: eval.NewTextArtifact("text/plain", "a"), SourceID: "s"},
		{ID: "e2", Kind: "sql", Content: eval.NewTextArtifact("text/plain", "b"), SourceID: "s"},
	}, "sha256:policy")
	allowed := EvidenceIDSet(set.Items)
	if len(allowed) != 2 {
		t.Errorf("expected 2 allowed evidence ids, got %d", len(allowed))
	}
	if _, ok := allowed["e1"]; !ok {
		t.Errorf("e1 missing from allowed set")
	}
	// Avoid unused-import noise by exercising spec.ConstructID identity.
	var _ spec.ConstructID = "faithfulness"
}
