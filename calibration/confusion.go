package calibration

import (
	"fmt"

	"github.com/go-go-golems/judgekit/assessment"
)

// IsEntailed reports whether a support label is a positive (entailed) verdict.
// Non-entailed groups both contradicted and insufficient, the two negative
// outcomes that require different interventions but share the negative side of
// a binary confusion matrix.
func IsEntailed(label assessment.SupportLabel) bool {
	return label == assessment.Entailed
}

// Confusion is a 2x2 contingency table over matched claims, where the gold label
// is the reference and the predicted label is the judge's verdict. Entailed is
// the positive class; non-entailed (contradicted or insufficient) is the
// negative class.
type Confusion struct {
	EntailedAsEntailed       int `json:"entailed_as_entailed" yaml:"entailed_as_entailed"`
	EntailedAsNonEntailed    int `json:"entailed_as_non_entailed" yaml:"entailed_as_non_entailed"`
	NonEntailedAsEntailed    int `json:"non_entailed_as_entailed" yaml:"non_entailed_as_entailed"`
	NonEntailedAsNonEntailed int `json:"non_entailed_as_entailed_neg" yaml:"non_entailed_as_non_entailed"`
}

// Add increments the matching cell.
func (c *Confusion) Add(gold, predicted assessment.SupportLabel) {
	g, p := IsEntailed(gold), IsEntailed(predicted)
	switch {
	case g && p:
		c.EntailedAsEntailed++
	case g && !p:
		c.EntailedAsNonEntailed++
	case !g && p:
		c.NonEntailedAsEntailed++
	default:
		c.NonEntailedAsNonEntailed++
	}
}

// Total returns the total number of matched pairs.
func (c Confusion) Total() int {
	return c.EntailedAsEntailed + c.EntailedAsNonEntailed + c.NonEntailedAsEntailed + c.NonEntailedAsNonEntailed
}

// TruePositives, FalseNegatives, FalsePositives, TrueNegatives name the cells
// in standard confusion-matrix terms for callers who prefer that vocabulary.
func (c Confusion) TruePositives() int  { return c.EntailedAsEntailed }
func (c Confusion) FalseNegatives() int { return c.EntailedAsNonEntailed }
func (c Confusion) FalsePositives() int { return c.NonEntailedAsEntailed }
func (c Confusion) TrueNegatives() int  { return c.NonEntailedAsNonEntailed }

// Sensitivity (recall) is TP / (TP + FN): the fraction of truly entailed claims
// the judge labels entailed. Returns 0 when there are no positive gold cases.
func (c Confusion) Sensitivity() float64 {
	denom := c.TruePositives() + c.FalseNegatives()
	if denom == 0 {
		return 0
	}
	return float64(c.TruePositives()) / float64(denom)
}

// Specificity is TN / (TN + FP): the fraction of truly non-entailed claims the
// judge labels non-entailed. Returns 0 when there are no negative gold cases.
func (c Confusion) Specificity() float64 {
	denom := c.TrueNegatives() + c.FalsePositives()
	if denom == 0 {
		return 0
	}
	return float64(c.TrueNegatives()) / float64(denom)
}

// FalseSupportRate is FP / (TP + FP): the fraction of claims the judge labeled
// entailed that were not entailed by gold (1 - precision). Returns 0 when the
// judge made no positive predictions.
func (c Confusion) FalseSupportRate() float64 {
	denom := c.TruePositives() + c.FalsePositives()
	if denom == 0 {
		return 0
	}
	return float64(c.FalsePositives()) / float64(denom)
}

// ConfusionFromClaims builds a confusion matrix over gold claims matched to
// predicted verdicts. A gold claim with no matching predicted verdict does not
// enter the confusion matrix (the judge failed to extract it); that failure is
// captured separately by extraction recall.
func ConfusionFromClaims(gold []GoldClaim, predicted []assessment.ClaimAssessment, matcher ClaimMatcher) (Confusion, error) {
	if matcher == nil {
		matcher = MatchByID
	}
	var c Confusion
	for i := range gold {
		match, ok := findPredicted(predicted, gold[i].Claim, matcher)
		if !ok {
			continue
		}
		c.Add(gold[i].Label, match.Label)
	}
	return c, nil
}

// findPredicted returns the predicted verdict matching a gold claim, if any.
func findPredicted(predicted []assessment.ClaimAssessment, gold assessment.Claim, matcher ClaimMatcher) (assessment.ClaimAssessment, bool) {
	for i := range predicted {
		if matcher(gold, predicted[i]) {
			return predicted[i], true
		}
	}
	return assessment.ClaimAssessment{}, false
}

// ValidateConfusion is a no-op placeholder kept for API symmetry; a confusion
// matrix is always valid once constructed. It exists so callers can write
// uniform validation paths.
func ValidateConfusion(c Confusion) error {
	if c.Total() < 0 {
		return fmt.Errorf("confusion: negative total")
	}
	return nil
}
