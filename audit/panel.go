package audit

import (
	"context"
	"fmt"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/internal/identifier"
	"github.com/go-go-golems/judgekit/judging"
	"golang.org/x/sync/errgroup"
)

// AggregationPolicy names how panel member reports are combined. Majority is
// an aggregation, not independent truth; the panel always preserves every
// member report and disagreement so a reader can see what was aggregated.
type AggregationPolicy string

const (
	// MajorityLabel names majority-vote aggregation over labels.
	MajorityLabel AggregationPolicy = "majority_label"
	// MeanValue names arithmetic-mean aggregation over numeric dimensions.
	MeanValue AggregationPolicy = "mean_value"
	// PreserveAll keeps member reports without producing an aggregate.
	PreserveAll AggregationPolicy = "preserve_all"
)

// PanelMember gives one judge a stable identity independent of the instance
// it evaluates. Agreement matrices are keyed by this ID.
type PanelMember struct {
	ID    string
	Judge judging.Judge
}

// Panel runs several named judges over the same instance and preserves every
// member report. It can compute agreement but cannot infer error independence
// without external labels.
type Panel struct {
	Members []PanelMember
	Policy  AggregationPolicy
}

// PanelResult preserves reports by member ID plus the pairwise claim-label
// agreement matrix.
type PanelResult struct {
	InstanceID      string
	Reports         map[string]assessment.Report
	AgreementMatrix map[string]map[string]float64
}

// Evaluate runs every member concurrently and computes pairwise agreement over
// the union of claim IDs. Missing claims therefore count as disagreement.
func (p Panel) Evaluate(ctx context.Context, inst eval.Instance) (PanelResult, error) {
	if err := eval.ValidateInstance(&inst); err != nil {
		return PanelResult{}, fmt.Errorf("panel: invalid instance: %w", err)
	}
	if len(p.Members) == 0 {
		return PanelResult{}, fmt.Errorf("panel: at least one member is required")
	}
	seen := make(map[string]struct{}, len(p.Members))
	for i, member := range p.Members {
		if err := identifier.Validate(member.ID); err != nil {
			return PanelResult{}, fmt.Errorf("panel: member %d id: %w", i, err)
		}
		if _, duplicate := seen[member.ID]; duplicate {
			return PanelResult{}, fmt.Errorf("panel: duplicate member id %q", member.ID)
		}
		seen[member.ID] = struct{}{}
		if member.Judge == nil {
			return PanelResult{}, fmt.Errorf("panel: member %q judge is nil", member.ID)
		}
	}

	byIndex := make([]assessment.Report, len(p.Members))
	g, gctx := errgroup.WithContext(ctx)
	for i, member := range p.Members {
		i, member := i, member
		g.Go(func() error {
			report, err := member.Judge.Evaluate(gctx, inst)
			if err != nil {
				return fmt.Errorf("panel: member %q: %w", member.ID, err)
			}
			byIndex[i] = report
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return PanelResult{}, err
	}
	reports := make(map[string]assessment.Report, len(p.Members))
	for i, member := range p.Members {
		reports[member.ID] = byIndex[i]
	}
	return PanelResult{
		InstanceID:      inst.ID,
		Reports:         reports,
		AgreementMatrix: pairwiseAgreement(p.Members, reports),
	}, nil
}

func pairwiseAgreement(members []PanelMember, reports map[string]assessment.Report) map[string]map[string]float64 {
	matrix := make(map[string]map[string]float64, len(members))
	for _, member := range members {
		matrix[member.ID] = make(map[string]float64, len(members))
		matrix[member.ID][member.ID] = 1
	}
	for i := 0; i < len(members); i++ {
		for j := i + 1; j < len(members); j++ {
			a := indexClaims(reports[members[i].ID])
			b := indexClaims(reports[members[j].ID])
			matches := 0
			total := 0
			for _, cid := range unionKeys(a, b) {
				av, aOK := a[cid]
				bv, bOK := b[cid]
				total++
				if aOK && bOK && av.Label == bv.Label {
					matches++
				}
			}
			var agreement float64
			if total > 0 {
				agreement = float64(matches) / float64(total)
			}
			matrix[members[i].ID][members[j].ID] = agreement
			matrix[members[j].ID][members[i].ID] = agreement
		}
	}
	return matrix
}
