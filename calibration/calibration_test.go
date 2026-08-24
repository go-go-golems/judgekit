package calibration

import (
	"strings"
	"testing"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/spec"
)

func ptr(f float64) *float64 { return &f }

func goldClaim(inst, claimID, text string, label assessment.SupportLabel, reviewers ...string) GoldClaim {
	return GoldClaim{
		InstanceID:  inst,
		Claim:       assessment.Claim{ID: claimID, Text: text, Importance: 1, RequiresEvidence: true},
		Label:       label,
		ReviewerIDs: reviewers,
	}
}

func goldSet(claims []GoldClaim) GoldSet {
	return GoldSet{Claims: claims, Digest: "sha256:dataset"}
}

func modelReport(inst string, claims []assessment.Claim, results []assessment.ClaimAssessment) assessment.Report {
	return assessment.Report{
		APIVersion:     assessment.ReportAPIVersion,
		InstanceID:     inst,
		InstanceDigest: "sha256:" + inst,
		ProtocolDigest: "sha256:protocol",
		Claims:         claims,
		ClaimResults:   results,
		Dimensions:     []assessment.DimensionResult{{ConstructID: "faithfulness", Applicable: true, Value: ptr(0.5)}},
	}
}

func TestValidateGoldClaim(t *testing.T) {
	good := goldClaim("inst-1", "c1", "text", assessment.Entailed, "r1")
	if err := ValidateGoldClaim(&good); err != nil {
		t.Fatalf("valid: %v", err)
	}
	badLabel := goldClaim("inst-1", "c1", "text", "bogus", "r1")
	if err := ValidateGoldClaim(&badLabel); err == nil {
		t.Errorf("accepted bogus label")
	}
	noReviewers := goldClaim("inst-1", "c1", "text", assessment.Entailed)
	if err := ValidateGoldClaim(&noReviewers); err == nil {
		t.Errorf("accepted gold claim with no reviewers")
	}
	dupReviewers := goldClaim("inst-1", "c1", "text", assessment.Entailed, "r1", "r1")
	if err := ValidateGoldClaim(&dupReviewers); err == nil {
		t.Errorf("accepted duplicate reviewer")
	}
}

func TestConfusion(t *testing.T) {
	gold := []GoldClaim{
		goldClaim("i", "c1", "a", assessment.Entailed, "r"),
		goldClaim("i", "c2", "b", assessment.Contradicted, "r"),
		goldClaim("i", "c3", "c", assessment.Insufficient, "r"),
	}
	predicted := []assessment.ClaimAssessment{
		{ClaimID: "c1", Label: assessment.Entailed, Reason: "ok"},
		{ClaimID: "c2", Label: assessment.Contradicted, Reason: "ok"},
		{ClaimID: "c3", Label: assessment.Entailed, Reason: "wrong"}, // FP: gold neg, pred pos
	}
	c, err := ConfusionFromClaims(gold, predicted, nil)
	if err != nil {
		t.Fatalf("ConfusionFromClaims: %v", err)
	}
	// c1: gold+ pred+ -> TP
	// c2: gold- pred- -> TN
	// c3: gold- pred+ -> FP
	if c.TruePositives() != 1 || c.TrueNegatives() != 1 || c.FalsePositives() != 1 || c.FalseNegatives() != 0 {
		t.Errorf("confusion = %+v", c)
	}
	if c.Sensitivity() != 1 { // TP/(TP+FN) = 1/1
		t.Errorf("sensitivity = %g, want 1", c.Sensitivity())
	}
	if c.Specificity() != 0.5 { // TN/(TN+FP) = 1/2
		t.Errorf("specificity = %g, want 0.5", c.Specificity())
	}
	if c.FalseSupportRate() != 0.5 { // FP/(TP+FP) = 1/2
		t.Errorf("false support rate = %g, want 0.5", c.FalseSupportRate())
	}
}

func TestExtractionRecall(t *testing.T) {
	gold := []GoldClaim{
		goldClaim("i", "c1", "a", assessment.Entailed, "r"),
		goldClaim("i", "c2", "b", assessment.Contradicted, "r"),
		goldClaim("i", "c3", "c", assessment.Insufficient, "r"),
	}
	// Model extracted c1 and c2 but missed c3.
	predicted := []assessment.ClaimAssessment{
		{ClaimID: "c1", Label: assessment.Entailed, Reason: "ok"},
		{ClaimID: "c2", Label: assessment.Contradicted, Reason: "ok"},
	}
	recall := ExtractionRecall(gold, predicted, nil)
	if recall != 2.0/3.0 {
		t.Errorf("recall = %g, want %g", recall, 2.0/3.0)
	}
	if ExtractionRecall(nil, predicted, nil) != 0 {
		t.Errorf("recall of empty gold should be 0")
	}
}

func TestBrierScore(t *testing.T) {
	bs, err := BrierScore([]float64{0.9, 0.1}, []float64{1, 0})
	if err != nil {
		t.Fatalf("brier: %v", err)
	}
	// (0.9-1)^2 + (0.1-0)^2 = 0.01 + 0.01 = 0.02; /2 = 0.01
	if bs > 0.0101 || bs < 0.0099 {
		t.Errorf("brier = %g, want ~0.01", bs)
	}
	if _, err := BrierScore([]float64{0.5}, []float64{0, 1}); err == nil {
		t.Errorf("accepted length mismatch")
	}
	if _, err := BrierScore([]float64{1.5}, []float64{1}); err == nil {
		t.Errorf("accepted out-of-range confidence")
	}
}

func TestECE(t *testing.T) {
	// Perfectly calibrated: two bins, each fully accurate at its confidence.
	conf := []float64{0.0, 1.0}
	out := []float64{0, 1}
	ece, err := ExpectedCalibrationError(conf, out, 2)
	if err != nil {
		t.Fatalf("ece: %v", err)
	}
	if ece != 0 {
		t.Errorf("ece = %g, want 0", ece)
	}
}

func TestCalibrate(t *testing.T) {
	gold := goldSet([]GoldClaim{
		goldClaim("i1", "c1", "the limit is five days", assessment.Entailed, "r1"),
		goldClaim("i1", "c2", "all leave carries over", assessment.Contradicted, "r1"),
	})
	claims := []assessment.Claim{
		{ID: "c1", Text: "the limit is five days", Importance: 1, RequiresEvidence: true},
		{ID: "c2", Text: "all leave carries over", Importance: 1, RequiresEvidence: true},
	}
	results := []assessment.ClaimAssessment{
		{ClaimID: "c1", Label: assessment.Entailed, EvidenceIDs: []string{"e1"}, Reason: "ok", VerdictConfidence: ptr(0.9), EntailedProbability: ptr(0.9)},
		{ClaimID: "c2", Label: assessment.Contradicted, EvidenceIDs: []string{"e1"}, Reason: "ok", VerdictConfidence: ptr(0.8), EntailedProbability: ptr(0.1)},
	}
	reports := map[string]assessment.Report{"i1": modelReport("i1", claims, results)}

	report, err := Calibrate(CalibrateInput{
		Gold:           gold,
		Reports:        reports,
		ProtocolDigest: "sha256:protocol",
		Bins:           5,
	})
	if err != nil {
		t.Fatalf("Calibrate: %v", err)
	}
	if report.ExtractionRecall != 1.0 {
		t.Errorf("recall = %g, want 1.0", report.ExtractionRecall)
	}
	if report.Sensitivity != 1.0 { // one TP, no FN
		t.Errorf("sensitivity = %g, want 1.0", report.Sensitivity)
	}
	if report.Specificity != 1.0 { // one TN, no FP
		t.Errorf("specificity = %g, want 1.0", report.Specificity)
	}
	if report.BrierScore == nil {
		t.Errorf("brier score should be present when confidence is emitted")
	} else if *report.BrierScore > 0.0101 || *report.BrierScore < 0.0099 {
		t.Errorf("brier = %g, want ~0.01 (entailed probabilities 0.9/0.1 vs outcomes 1/0)", *report.BrierScore)
	}
	if report.ECE == nil {
		t.Errorf("ece should be present when confidence is emitted")
	}
	if !strings.HasPrefix(report.Digest, "sha256:") {
		t.Errorf("digest = %q", report.Digest)
	}
	if err := ValidateReport(&report); err != nil {
		t.Errorf("ValidateReport: %v", err)
	}
}

func TestCalibrateMissedClaimReducesRecallNotConfusion(t *testing.T) {
	gold := goldSet([]GoldClaim{
		goldClaim("i1", "c1", "a", assessment.Entailed, "r1"),
		goldClaim("i1", "c2", "b", assessment.Contradicted, "r1"), // model misses this one
	})
	claims := []assessment.Claim{{ID: "c1", Text: "a", Importance: 1, RequiresEvidence: true}}
	results := []assessment.ClaimAssessment{{ClaimID: "c1", Label: assessment.Entailed, EvidenceIDs: []string{"e1"}, Reason: "ok"}}
	reports := map[string]assessment.Report{"i1": modelReport("i1", claims, results)}
	report, err := Calibrate(CalibrateInput{Gold: gold, Reports: reports, ProtocolDigest: "sha256:protocol"})
	if err != nil {
		t.Fatalf("Calibrate: %v", err)
	}
	if report.ExtractionRecall != 0.5 {
		t.Errorf("recall = %g, want 0.5 (missed one of two)", report.ExtractionRecall)
	}
	// Confusion is over the one matched claim: TP. Sensitivity = 1 (no FN among matched).
	if report.Sensitivity != 1.0 {
		t.Errorf("sensitivity = %g, want 1.0 (only matched claim is a TP)", report.Sensitivity)
	}
}

func TestCalibrateVerdictConfidenceAloneLeavesBrierNil(t *testing.T) {
	gold := goldSet([]GoldClaim{goldClaim("i1", "c1", "a", assessment.Entailed, "r1")})
	claims := []assessment.Claim{{ID: "c1", Text: "a", Importance: 1, RequiresEvidence: true}}
	results := []assessment.ClaimAssessment{{ClaimID: "c1", Label: assessment.Entailed, EvidenceIDs: []string{"e1"}, Reason: "ok", VerdictConfidence: ptr(0.95)}} // no EntailedProbability
	reports := map[string]assessment.Report{"i1": modelReport("i1", claims, results)}
	report, err := Calibrate(CalibrateInput{Gold: gold, Reports: reports, ProtocolDigest: "sha256:protocol"})
	if err != nil {
		t.Fatalf("Calibrate: %v", err)
	}
	if report.BrierScore != nil || report.ECE != nil {
		t.Errorf("brier/ece must be nil when no entailed probability is emitted")
	}
}

func TestCalibrateRejectsReportFromAnotherProtocol(t *testing.T) {
	gold := goldSet([]GoldClaim{goldClaim("i1", "c1", "a", assessment.Entailed, "r1")})
	report := modelReport("i1", []assessment.Claim{{ID: "c1", Text: "a", Importance: 1}}, []assessment.ClaimAssessment{{ClaimID: "c1", Label: assessment.Entailed, Reason: "ok"}})
	report.ProtocolDigest = "sha256:other"
	_, err := Calibrate(CalibrateInput{
		Gold:           gold,
		Reports:        map[string]assessment.Report{"i1": report},
		ProtocolDigest: "sha256:protocol",
	})
	if err == nil {
		t.Errorf("accepted calibration report from another protocol")
	}
}

func TestCalibrateRejectsBadProtocolDigest(t *testing.T) {
	gold := goldSet([]GoldClaim{goldClaim("i1", "c1", "a", assessment.Entailed, "r1")})
	if _, err := Calibrate(CalibrateInput{Gold: gold, Reports: map[string]assessment.Report{}, ProtocolDigest: "not-a-digest"}); err == nil {
		t.Errorf("accepted bad protocol digest")
	}
}

func TestGoldDimensionValidation(t *testing.T) {
	d := GoldDimension{InstanceID: "i1", ConstructID: spec.ConstructID("faithfulness"), Value: ptr(0.5), ReviewerIDs: []string{"r1"}}
	if err := ValidateGoldDimension(&d); err != nil {
		t.Fatalf("valid: %v", err)
	}
	d2 := d
	d2.Value = nil
	d2.Label = "attempted"
	if err := ValidateGoldDimension(&d2); err != nil {
		t.Errorf("label-only dimension should be valid: %v", err)
	}
	d3 := d
	d3.Value = nil
	d3.Label = ""
	if err := ValidateGoldDimension(&d3); err == nil {
		t.Errorf("accepted dimension with neither value nor label")
	}
}

func TestMatchByText(t *testing.T) {
	gold := []GoldClaim{goldClaim("i1", "gold-1", "The limit is five days.", assessment.Entailed, "r1")}
	claims := []assessment.Claim{{ID: "model-1", Text: "the limit is five days.", Importance: 1, RequiresEvidence: true}}
	results := []assessment.ClaimAssessment{{ClaimID: "model-1", Label: assessment.Entailed, EvidenceIDs: []string{"e1"}, Reason: "ok"}}
	report := modelReport("i1", claims, results)
	matcher := MatchByText(report)
	recall := ExtractionRecall(gold, results, matcher)
	if recall != 1.0 {
		t.Errorf("text-matched recall = %g, want 1.0", recall)
	}
}
