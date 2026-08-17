package audit

import (
	"context"
	"fmt"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/judging"
)

// AggregationPolicy names how panel member reports are combined. Majority is
// an aggregation, not independent truth; the panel always preserves every
// member report and disagreement so a reader can see what was aggregated.
type AggregationPolicy string

const (
	// MajorityLabel reports the most common claim label across panel members.
	MajorityLabel AggregationPolicy = "majority_label"
	// MeanValue reports the mean dimension value across panel members.
	MeanValue AggregationPolicy = "mean_value"
	// PreserveAll reports nothing aggregated; it returns every member report.
	PreserveAll AggregationPolicy = "preserve_all"
)

// Panel runs several judges over the same instance and preserves every member
// report. It can compute a pairwise agreement matrix but cannot infer error
// independence without external labels.
type Panel struct {
	Judges []judging.Judge
	Policy AggregationPolicy
}

// PanelResult is the output of one panel evaluation: every member report is
// retained, plus the pairwise claim-label agreement matrix.
type PanelResult struct {
	InstanceID      string
	Reports         []assessment.Report
	AgreementMatrix map[string]map[string]float64 // judge i name -> judge j name -> agreement
}

// Evaluate runs every judge over inst concurrently, preserves each report,
// and computes pairwise claim-label agreement. It does not collapse the
// reports into one; that is the application's decision.
func (p Panel) Evaluate(ctx context.Context, inst eval.Instance) (PanelResult, error) {
	if err := eval.ValidateInstance(&inst); err != nil {
		return PanelResult{}, fmt.Errorf("panel: invalid instance: %w", err)
	}
	if len(p.Judges) == 0 {
		return PanelResult{}, fmt.Errorf("panel: at least one judge is required")
	}
	reports := make([]assessment.Report, len(p.Judges))
	errs := make([]error, len(p.Judges))
	for i := range p.Judges {
		if p.Judges[i] == nil {
			return PanelResult{}, fmt.Errorf("panel: judge %d is nil", i)
		}
		reports[i], errs[i] = p.Judges[i].Evaluate(ctx, inst)
	}
	for i := range errs {
		if errs[i] != nil {
			return PanelResult{}, fmt.Errorf("panel: judge %d: %w", i, errs[i])
		}
	}
	result := PanelResult{
		InstanceID:      inst.ID,
		Reports:         reports,
		AgreementMatrix: pairwiseAgreement(reports),
	}
	return result, nil
}

// pairwiseAgreement computes claim-label agreement between every pair of
// reports over the claims they share. A judge pair with no shared claims has
// agreement 0, not 1, so the matrix does not over-report agreement.
func pairwiseAgreement(reports []assessment.Report) map[string]map[string]float64 {
	matrix := make(map[string]map[string]float64, len(reports))
	for i := range reports {
		matrix[reports[i].InstanceID] = make(map[string]float64)
	}
	for i := 0; i < len(reports); i++ {
		for j := i + 1; j < len(reports); j++ {
			a := indexClaims(reports[i])
			b := indexClaims(reports[j])
			matches := 0
			total := 0
			for cid, av := range a {
				bv, ok := b[cid]
				if !ok {
					continue
				}
				total++
				if av.Label == bv.Label {
					matches++
				}
			}
			var agree float64
			if total > 0 {
				agree = float64(matches) / float64(total)
			}
			matrix[reports[i].InstanceID][reports[j].InstanceID] = agree
			matrix[reports[j].InstanceID][reports[i].InstanceID] = agree
		}
	}
	return matrix
}
