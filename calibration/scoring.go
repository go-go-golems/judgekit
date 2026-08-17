package calibration

import (
	"fmt"
	"math"
	"sort"

	"github.com/go-go-golems/judgekit/assessment"
)

// BrierScore is the mean squared error of predicted probabilities against
// binary outcomes. It is a proper scoring rule: in expectation, a forecaster
// minimizes it by reporting its true belief.
//
//	BS = (1/n) * sum_i (p_i - z_i)^2
//
// confidences are predicted probabilities in [0,1]; outcomes are 0 or 1
// (1 = the event occurred, e.g. the claim was entailed by gold). BrierScore
// returns an error if the slices differ in length or a value is out of range.
func BrierScore(confidences []float64, outcomes []float64) (float64, error) {
	if len(confidences) != len(outcomes) {
		return 0, fmt.Errorf("brier: %d confidences but %d outcomes", len(confidences), len(outcomes))
	}
	if len(confidences) == 0 {
		return 0, nil
	}
	var sum float64
	for i := range confidences {
		p := confidences[i]
		z := outcomes[i]
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return 0, fmt.Errorf("brier: confidence %g at %d must be in [0,1]", p, i)
		}
		if z != 0 && z != 1 {
			return 0, fmt.Errorf("brier: outcome %g at %d must be 0 or 1", z, i)
		}
		d := p - z
		sum += d * d
	}
	return sum / float64(len(confidences)), nil
}

// ExpectedCalibrationError bins predicted probabilities and compares the
// average confidence in each bin with the empirical accuracy in that bin.
//
//	ECE = sum_b (|I_b| / n) * |acc(I_b) - conf(I_b)|
//
// bins is the number of equal-width bins over [0,1]. ECE is intuitive but
// depends on binning and can hide within-bin structure; report reliability
// diagrams alongside it for high-stakes use. ECE returns an error on length
// mismatch or out-of-range values.
func ExpectedCalibrationError(confidences []float64, outcomes []float64, bins int) (float64, error) {
	if bins < 1 {
		return 0, fmt.Errorf("ece: bins must be >= 1, got %d", bins)
	}
	if len(confidences) != len(outcomes) {
		return 0, fmt.Errorf("ece: %d confidences but %d outcomes", len(confidences), len(outcomes))
	}
	if len(confidences) == 0 {
		return 0, nil
	}
	type bucket struct {
		confSum float64
		correct int
		count   int
	}
	bs := make([]bucket, bins)
	width := 1.0 / float64(bins)
	for i := range confidences {
		p := confidences[i]
		z := outcomes[i]
		if math.IsNaN(p) || math.IsInf(p, 0) || p < 0 || p > 1 {
			return 0, fmt.Errorf("ece: confidence %g at %d must be in [0,1]", p, i)
		}
		if z != 0 && z != 1 {
			return 0, fmt.Errorf("ece: outcome %g at %d must be 0 or 1", z, i)
		}
		idx := int(p / width)
		if idx >= bins {
			idx = bins - 1
		}
		bs[idx].confSum += p
		bs[idx].count++
		if z == 1 {
			bs[idx].correct++
		}
	}
	n := float64(len(confidences))
	var ece float64
	for _, b := range bs {
		if b.count == 0 {
			continue
		}
		acc := float64(b.correct) / float64(b.count)
		conf := b.confSum / float64(b.count)
		ece += (float64(b.count) / n) * math.Abs(acc-conf)
	}
	return ece, nil
}

// ConfidenceOutcomePairs extracts (confidence, outcome) pairs from matched gold
// claims and predicted verdicts. outcome is 1 when the gold label is entailed.
// Pairs with no predicted confidence are skipped, so Brier and ECE are computed
// only over claims the protocol emitted a confidence for. If no pair has a
// confidence, the returned slices are empty (callers report a nil metric).
func ConfidenceOutcomePairs(gold []GoldClaim, predicted []assessment.ClaimAssessment, matcher ClaimMatcher) (confidences, outcomes []float64) {
	if matcher == nil {
		matcher = MatchByID
	}
	for i := range gold {
		match, ok := findPredicted(predicted, gold[i].Claim, matcher)
		if !ok || match.Confidence == nil {
			continue
		}
		confidences = append(confidences, *match.Confidence)
		if IsEntailed(gold[i].Label) {
			outcomes = append(outcomes, 1)
		} else {
			outcomes = append(outcomes, 0)
		}
	}
	return confidences, outcomes
}

// sortConfidences is a helper retained for future reliability-diagram support;
// it returns the indices that would sort confidences ascending.
func sortConfidences(confidences []float64) []int {
	idx := make([]int, len(confidences))
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(a, b int) bool { return confidences[idx[a]] < confidences[idx[b]] })
	return idx
}
