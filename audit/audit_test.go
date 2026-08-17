package audit

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/judging"
	"github.com/go-go-golems/judgekit/spec"
)

func ptr(f float64) *float64 { return &f }

// scriptedJudge returns a canned report keyed by instance ID, so a probe's
// base and variant instances can produce different reports deterministically.
type scriptedJudge struct {
	reports map[string]assessment.Report
}

func (s *scriptedJudge) Evaluate(_ context.Context, inst eval.Instance) (assessment.Report, error) {
	r, ok := s.reports[inst.ID]
	if !ok {
		return assessment.Report{}, nil
	}
	// Re-stamp the instance/digest so the report matches the instance passed in.
	r.InstanceID = inst.ID
	r.InstanceDigest = inst.Digest
	return r, nil
}

var _ judging.Judge = (*scriptedJudge)(nil)

func baseReport(value float64, label assessment.SupportLabel) assessment.Report {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	return assessment.Report{
		APIVersion:     assessment.ReportAPIVersion,
		InstanceID:     "inst",
		InstanceDigest: "sha256:inst",
		ProtocolDigest: "sha256:protocol",
		Claims: []assessment.Claim{
			{ID: "c1", Text: "claim one", Importance: 1, RequiresEvidence: true},
		},
		ClaimResults: []assessment.ClaimAssessment{
			{ClaimID: "c1", Label: label, EvidenceIDs: []string{"e1"}, Reason: "ok"},
		},
		Dimensions: []assessment.DimensionResult{
			{ConstructID: "faithfulness", Applicable: true, Value: ptr(value)},
		},
		StartedAt:  now,
		FinishedAt: now.Add(time.Second),
	}
}

func makeInstance(t *testing.T, id string) eval.Instance {
	t.Helper()
	set, err := eval.NewEvidenceSet([]eval.EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: eval.NewTextArtifact("text/plain", "evidence"), SourceID: "s"},
	}, "sha256:policy")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	inst, err := eval.NewInstance(id,
		eval.NewTextArtifact("text/plain", "q"),
		eval.NewTextArtifact("text/plain", "a"),
		set, nil, nil, nil)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	return inst
}

func makeProbe(t *testing.T, id string, baseValue float64, baseLabel assessment.SupportLabel, variantValue float64, variantLabel assessment.SupportLabel) Probe {
	t.Helper()
	return Probe{
		ID:              id,
		Kind:            Repeat,
		BaseInstance:    makeInstance(t, "inst-base"),
		VariantInstance: makeInstance(t, "inst-variant"),
		Invariants:      []string{"same construct"},
	}
}

func TestValidateProbe(t *testing.T) {
	p := Probe{ID: "p1", Kind: Repeat, BaseInstance: makeInstance(t, "i1"), VariantInstance: makeInstance(t, "i2"), Invariants: []string{"x"}}
	if err := ValidateProbe(&p); err != nil {
		t.Fatalf("valid: %v", err)
	}
	bad := p
	bad.Kind = "bogus"
	if err := ValidateProbe(&bad); err == nil {
		t.Errorf("accepted bogus kind")
	}
	noInv := p
	noInv.Invariants = nil
	if err := ValidateProbe(&noInv); err == nil {
		t.Errorf("accepted probe with no invariants")
	}
}

func TestReliabilityDetectsDisagreement(t *testing.T) {
	base := baseReport(0.5, assessment.Entailed)
	variant := baseReport(0.9, assessment.Contradicted)
	base.InstanceID = "inst-base"
	variant.InstanceID = "inst-variant"
	judge := &scriptedJudge{reports: map[string]assessment.Report{
		"inst-base":    base,
		"inst-variant": variant,
	}}
	probe := Probe{
		ID:              "p1",
		Kind:            Repeat,
		BaseInstance:    makeInstance(t, "inst-base"),
		VariantInstance: makeInstance(t, "inst-variant"),
		Invariants:      []string{"same construct"},
	}
	set, err := NewProbeSet([]Probe{probe})
	if err != nil {
		t.Fatalf("NewProbeSet: %v", err)
	}
	report, err := Reliability(context.Background(), judge, set, "sha256:protocol")
	if err != nil {
		t.Fatalf("Reliability: %v", err)
	}
	if report.TotalPairs != 1 {
		t.Errorf("total pairs = %d, want 1", report.TotalPairs)
	}
	// The two reports differ on faithfulness value (0.5 vs 0.9) and on claim label.
	if len(report.Disagreements) < 1 {
		t.Errorf("expected disagreements, got %d", len(report.Disagreements))
	}
	// Claim-label agreement should be 0 (entailed vs contradicted).
	if report.ClaimLabelAgreement != 0 {
		t.Errorf("claim agreement = %g, want 0", report.ClaimLabelAgreement)
	}
	if got := report.DimensionAgreement["faithfulness"]; got != 0 {
		t.Errorf("dimension agreement = %g, want 0", got)
	}
	if got := report.MeanAbsoluteDelta["faithfulness"]; got != 0.4 {
		t.Errorf("mean abs delta = %g, want 0.4", got)
	}
	if err := ValidateReport(&report); err != nil {
		t.Errorf("ValidateReport: %v", err)
	}
}

func TestReliabilityPerfectAgreement(t *testing.T) {
	r := baseReport(0.5, assessment.Entailed)
	r.InstanceID = "inst-variant"
	judge := &scriptedJudge{reports: map[string]assessment.Report{
		"inst-base":    baseReport(0.5, assessment.Entailed),
		"inst-variant": r,
	}}
	probe := Probe{
		ID:              "p1",
		Kind:            Repeat,
		BaseInstance:    makeInstance(t, "inst-base"),
		VariantInstance: makeInstance(t, "inst-variant"),
		Invariants:      []string{"same construct"},
	}
	set, _ := NewProbeSet([]Probe{probe})
	report, err := Reliability(context.Background(), judge, set, "sha256:protocol")
	if err != nil {
		t.Fatalf("Reliability: %v", err)
	}
	if report.ClaimLabelAgreement != 1 {
		t.Errorf("claim agreement = %g, want 1", report.ClaimLabelAgreement)
	}
	if got := report.DimensionAgreement["faithfulness"]; got != 1 {
		t.Errorf("dimension agreement = %g, want 1", got)
	}
	if len(report.Disagreements) != 0 {
		t.Errorf("expected no disagreements, got %d", len(report.Disagreements))
	}
}

func TestReliabilityRejectsBadProtocolDigest(t *testing.T) {
	probe := Probe{ID: "p1", Kind: Repeat, BaseInstance: makeInstance(t, "i1"), VariantInstance: makeInstance(t, "i2"), Invariants: []string{"x"}}
	set, _ := NewProbeSet([]Probe{probe})
	if _, err := Reliability(context.Background(), &scriptedJudge{}, set, "nope"); err == nil {
		t.Errorf("accepted bad protocol digest")
	}
}

func TestPanelPreservesAllReports(t *testing.T) {
	r1 := baseReport(0.5, assessment.Entailed)
	r2 := baseReport(0.7, assessment.Contradicted)
	j1 := &scriptedJudge{reports: map[string]assessment.Report{"inst": r1}}
	j2 := &scriptedJudge{reports: map[string]assessment.Report{"inst": r2}}
	panel := Panel{Judges: []judging.Judge{j1, j2}, Policy: PreserveAll}
	inst := makeInstance(t, "inst")
	result, err := panel.Evaluate(context.Background(), inst)
	if err != nil {
		t.Fatalf("panel: %v", err)
	}
	if len(result.Reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(result.Reports))
	}
	// Pairwise agreement between entailed and contradicted is 0.
	if got := result.AgreementMatrix["inst"]["inst"]; got != 0 {
		t.Errorf("self agreement = %g, want 0 (labels differ)", got)
	}
}

func TestPanelRejectsEmpty(t *testing.T) {
	panel := Panel{Judges: nil}
	inst := makeInstance(t, "inst")
	if _, err := panel.Evaluate(context.Background(), inst); err == nil {
		t.Errorf("accepted empty panel")
	}
}

func TestCompareReportsLabelDimension(t *testing.T) {
	base := baseReport(0.5, assessment.Entailed)
	base.Dimensions = append(base.Dimensions, assessment.DimensionResult{ConstructID: "abstention", Applicable: true, Label: "attempted"})
	variant := baseReport(0.5, assessment.Entailed)
	variant.Dimensions = append(variant.Dimensions, assessment.DimensionResult{ConstructID: "abstention", Applicable: true, Label: "abstained"})
	disags := CompareReports("p1", FormatTransform, base, variant)
	// abstention label differs (attempted vs abstained); faithfulness same.
	found := false
	for _, d := range disags {
		if string(d.ConstructID) == "abstention" && d.BaseLabel == "attempted" && d.VariantLabel == "abstained" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected abstention label disagreement, got %v", disags)
	}
	_ = spec.ConstructID("abstention")
}

func TestReliabilityDeterministicDigest(t *testing.T) {
	judge := &scriptedJudge{reports: map[string]assessment.Report{
		"inst-base":    baseReport(0.5, assessment.Entailed),
		"inst-variant": baseReport(0.5, assessment.Entailed),
	}}
	probe := Probe{ID: "p1", Kind: Repeat, BaseInstance: makeInstance(t, "inst-base"), VariantInstance: makeInstance(t, "inst-variant"), Invariants: []string{"x"}}
	set, _ := NewProbeSet([]Probe{probe})
	r1, _ := Reliability(context.Background(), judge, set, "sha256:protocol")
	r2, _ := Reliability(context.Background(), judge, set, "sha256:protocol")
	if r1.Digest != r2.Digest {
		t.Errorf("reliability digest not deterministic: %s vs %s", r1.Digest, r2.Digest)
	}
	if !strings.HasPrefix(r1.Digest, "sha256:") {
		t.Errorf("digest = %q", r1.Digest)
	}
}
