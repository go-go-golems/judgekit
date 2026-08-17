package calibration

import (
	"strings"

	"github.com/go-go-golems/judgekit/assessment"
)

// ClaimMatcher decides whether a gold claim and a predicted verdict refer to
// the same underlying claim. Matching begins with human linkage in calibration
// data; automated semantic matching may be added later but must itself be
// validated, so the default matcher is deterministic (by ID).
type ClaimMatcher func(gold assessment.Claim, predicted assessment.ClaimAssessment) bool

// MatchByID matches a gold claim to a predicted verdict by claim ID. This is
// the intended primary mechanism: human linkage assigns the same ID to the
// gold claim and the model's extracted claim.
func MatchByID(gold assessment.Claim, predicted assessment.ClaimAssessment) bool {
	return gold.ID != "" && gold.ID == predicted.ClaimID
}

// MatchByText matches a gold claim to a predicted verdict by normalized claim
// text. It requires a lookup from predicted claim ID to the predicted claim
// text (which lives in the report's Claims, not in ClaimAssessment). It is a
// deterministic fallback for cases where IDs are not aligned, but it is brittle
// to paraphrase and should not be the only matching strategy.
func MatchByText(report assessment.Report) ClaimMatcher {
	byID := make(map[string]assessment.Claim, len(report.Claims))
	for i := range report.Claims {
		byID[report.Claims[i].ID] = report.Claims[i]
	}
	return func(gold assessment.Claim, predicted assessment.ClaimAssessment) bool {
		pc, ok := byID[predicted.ClaimID]
		if !ok {
			return false
		}
		return strings.EqualFold(strings.TrimSpace(gold.Text), strings.TrimSpace(pc.Text))
	}
}

// ExtractionRecall is the fraction of human-enumerated claims the model
// extracted: matched model-extracted factual claims divided by
// human-enumerated factual claims. A judge that extracts fewer claims can
// appear accurate on the claims it does extract; recall exposes that.
//
//	recall = matched / len(gold)
//
// A gold claim counts as matched when some predicted verdict matches it under
// the matcher. The matcher defaults to MatchByID.
func ExtractionRecall(gold []GoldClaim, predicted []assessment.ClaimAssessment, matcher ClaimMatcher) float64 {
	if len(gold) == 0 {
		return 0
	}
	if matcher == nil {
		matcher = MatchByID
	}
	matched := 0
	for i := range gold {
		if _, ok := findPredicted(predicted, gold[i].Claim, matcher); ok {
			matched++
		}
	}
	return float64(matched) / float64(len(gold))
}
