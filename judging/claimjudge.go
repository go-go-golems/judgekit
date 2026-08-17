package judging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/protocol"
	"github.com/go-go-golems/judgekit/spec"
)

// ClaimProtocol renders the two prompts for a two-stage claim judge. The
// extraction prompt MUST NOT reveal the evidence; the support prompt receives
// the claims and (through the instance) the evidence.
type ClaimProtocol interface {
	ExtractPrompt(inst eval.Instance) (string, error)
	SupportPrompt(inst eval.Instance, claims []assessment.Claim) (string, error)
}

// ClaimJudge implements Judge with a two-stage decomposed evaluation: extract
// claims with evidence hidden, judge each claim's support against the evidence,
// aggregate per the measurement contract, and seal the report.
type ClaimJudge struct {
	Contract spec.ContractDocument
	Protocol protocol.Document
	Prompts  ClaimProtocol
	Generate Generator
	Cache    Cache
	Repairer Repairer
	Clock    func() time.Time
}

var _ Judge = (*ClaimJudge)(nil)

func (j *ClaimJudge) now() time.Time {
	if j.Clock != nil {
		return j.Clock()
	}
	return time.Now()
}

func (j *ClaimJudge) repairer() Repairer {
	if j.Repairer != nil {
		return j.Repairer
	}
	return DefaultRepairer{}
}

func (j *ClaimJudge) cache() Cache {
	if j.Cache != nil {
		return j.Cache
	}
	return NoopCache{}
}

func promptDigest(prompt string) string {
	sum := sha256.Sum256([]byte(prompt))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// extractPayload is the model response for claim extraction.
type extractPayload struct {
	Statements []string `json:"statements"`
}

// supportVerdict is one per-claim verdict in the support response.
type supportVerdict struct {
	Claim       int                     `json:"claim"` // 1-based index
	Label       assessment.SupportLabel `json:"label"`
	EvidenceIDs []string                `json:"evidence_ids"`
	Reason      string                  `json:"reason"`
	Confidence  *float64                `json:"confidence,omitempty"`
}

// directDimension is a whole-answer dimension emitted by the support judge for
// constructs that are not aggregated from claim labels.
type directDimension struct {
	ConstructID spec.ConstructID `json:"construct_id"`
	Value       *float64         `json:"value,omitempty"`
	Label       string           `json:"label,omitempty"`
	Confidence  *float64         `json:"confidence,omitempty"`
	EvidenceIDs []string         `json:"evidence_ids,omitempty"`
	Reason      string           `json:"reason,omitempty"`
}

// supportPayload is the model response for support judging.
type supportPayload struct {
	Verdicts   []supportVerdict  `json:"verdicts"`
	Dimensions []directDimension `json:"dimensions,omitempty"`
}

// Evaluate runs the two-stage judge over inst and returns a sealed report.
func (j *ClaimJudge) Evaluate(ctx context.Context, inst eval.Instance) (assessment.Report, error) {
	if err := eval.ValidateInstance(&inst); err != nil {
		return assessment.Report{}, fmt.Errorf("claim judge: invalid instance: %w", err)
	}
	allowed := assessment.EvidenceIDSet(inst.Evidence.Items)
	started := j.now()

	claims, err := j.extractClaims(ctx, inst)
	if err != nil {
		return assessment.Report{}, err
	}
	results, directs, err := j.judgeSupport(ctx, inst, claims)
	if err != nil {
		return assessment.Report{}, err
	}
	dims, err := j.aggregate(claims, results, directs)
	if err != nil {
		return assessment.Report{}, err
	}

	report := assessment.Report{
		APIVersion:     assessment.ReportAPIVersion,
		InstanceID:     inst.ID,
		InstanceDigest: inst.Digest,
		ProtocolDigest: j.Protocol.Digest,
		Claims:         claims,
		ClaimResults:   results,
		Dimensions:     dims,
		StartedAt:      started,
		FinishedAt:     j.now(),
	}
	if err := assessment.Seal(&report, allowed); err != nil {
		return assessment.Report{}, fmt.Errorf("claim judge: %w", err)
	}
	return report, nil
}

func (j *ClaimJudge) extractClaims(ctx context.Context, inst eval.Instance) ([]assessment.Claim, error) {
	prompt, err := j.Prompts.ExtractPrompt(inst)
	if err != nil {
		return nil, fmt.Errorf("extract prompt: %w", err)
	}
	payload, err := generateAndDecode[extractPayload](ctx, j, inst, "extract", prompt)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}
	claims := make([]assessment.Claim, 0, len(payload.Statements))
	for _, s := range payload.Statements {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		claims = append(claims, assessment.Claim{
			Text:             s,
			Importance:       1,
			RequiresEvidence: true,
		})
	}
	for i := range claims {
		claims[i].ID = fmt.Sprintf("c%d", i+1)
	}
	return claims, nil
}

func (j *ClaimJudge) judgeSupport(ctx context.Context, inst eval.Instance, claims []assessment.Claim) ([]assessment.ClaimAssessment, []directDimension, error) {
	prompt, err := j.Prompts.SupportPrompt(inst, claims)
	if err != nil {
		return nil, nil, fmt.Errorf("support prompt: %w", err)
	}
	payload, err := generateAndDecode[supportPayload](ctx, j, inst, "support", prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("support: %w", err)
	}
	if len(payload.Verdicts) != len(claims) {
		return nil, nil, fmt.Errorf("support: got %d verdicts for %d claims", len(payload.Verdicts), len(claims))
	}
	results := make([]assessment.ClaimAssessment, len(claims))
	for i, v := range payload.Verdicts {
		if v.Claim != i+1 {
			return nil, nil, fmt.Errorf("support: verdict %d references claim %d, want %d", i+1, v.Claim, i+1)
		}
		results[i] = assessment.ClaimAssessment{
			ClaimID:     claims[i].ID,
			Label:       v.Label,
			EvidenceIDs: v.EvidenceIDs,
			Reason:      v.Reason,
			Confidence:  v.Confidence,
		}
	}
	return results, payload.Dimensions, nil
}

func (j *ClaimJudge) aggregate(claims []assessment.Claim, results []assessment.ClaimAssessment, directs []directDimension) ([]assessment.DimensionResult, error) {
	directByConstruct := make(map[spec.ConstructID]directDimension, len(directs))
	for _, d := range directs {
		directByConstruct[d.ConstructID] = d
	}
	labelCount := make(map[string]int, len(results))
	for _, r := range results {
		labelCount[string(r.Label)]++
	}
	dims := make([]assessment.DimensionResult, 0, len(j.Contract.Contract.Constructs))
	for i := range j.Contract.Contract.Constructs {
		c := &j.Contract.Contract.Constructs[i]
		agg := j.Contract.Contract.Aggregations[c.ID]
		dim, err := aggregateConstruct(c, agg, labelCount, results, directByConstruct)
		if err != nil {
			return nil, err
		}
		dims = append(dims, dim)
	}
	return dims, nil
}

func aggregateConstruct(c *spec.Construct, agg spec.Aggregation, labelCount map[string]int, results []assessment.ClaimAssessment, directs map[spec.ConstructID]directDimension) (assessment.DimensionResult, error) {
	switch agg.Method {
	case spec.MethodFraction:
		num := countLabels(agg.Numerator, labelCount)
		den := countLabels(agg.Denominator, labelCount)
		if den == 0 {
			return emptyDimension(c, agg), nil
		}
		v := float64(num) / float64(den)
		ev := unionEvidenceForLabels(results, agg.Numerator)
		return assessment.DimensionResult{ConstructID: c.ID, Applicable: true, Value: &v, EvidenceIDs: ev}, nil
	case spec.MethodDirect:
		d, ok := directs[c.ID]
		if !ok {
			return assessment.DimensionResult{}, fmt.Errorf("aggregation: construct %q is method direct but the judge emitted no dimension for it", c.ID)
		}
		return assessment.DimensionResult{
			ConstructID: c.ID,
			Applicable:  true,
			Value:       d.Value,
			Label:       d.Label,
			Confidence:  d.Confidence,
			EvidenceIDs: d.EvidenceIDs,
		}, nil
	default:
		return assessment.DimensionResult{}, fmt.Errorf("aggregation: method %q is not supported by the claim judge (use fraction or direct)", agg.Method)
	}
}

func emptyDimension(c *spec.Construct, agg spec.Aggregation) assessment.DimensionResult {
	switch agg.EmptyPolicy {
	case spec.EmptyVacuousPerfect:
		v := 1.0
		if c.Range != nil {
			v = c.Range.Maximum
		}
		return assessment.DimensionResult{ConstructID: c.ID, Applicable: true, Value: &v}
	case spec.EmptyZero:
		v := 0.0
		if c.Range != nil {
			v = c.Range.Minimum
		}
		return assessment.DimensionResult{ConstructID: c.ID, Applicable: true, Value: &v}
	default:
		return assessment.DimensionResult{ConstructID: c.ID, Applicable: false}
	}
}

func countLabels(list string, labelCount map[string]int) int {
	total := 0
	for _, l := range strings.Split(list, ",") {
		total += labelCount[strings.TrimSpace(l)]
	}
	return total
}

func unionEvidenceForLabels(results []assessment.ClaimAssessment, list string) []string {
	wanted := make(map[string]struct{})
	for _, l := range strings.Split(list, ",") {
		wanted[strings.TrimSpace(l)] = struct{}{}
	}
	set := make(map[string]struct{})
	for _, r := range results {
		if _, ok := wanted[string(r.Label)]; !ok {
			continue
		}
		for _, e := range r.EvidenceIDs {
			set[e] = struct{}{}
		}
	}
	out := make([]string, 0, len(set))
	for e := range set {
		out = append(out, e)
	}
	sort.Strings(out)
	return out
}

// generateAndDecode generates (with cache) and strictly decodes one JSON
// object, retrying once per repair attempt on a structural failure. Semantic
// failures are not retried; they surface to the caller and fail closed at seal.
func generateAndDecode[T any](ctx context.Context, j *ClaimJudge, inst eval.Instance, step, prompt string) (T, error) {
	var zero T
	maxAttempts := j.Protocol.Protocol.Retry.MaximumAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	current := prompt
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		raw, err := j.generate(ctx, inst, step, current)
		if err != nil {
			return zero, fmt.Errorf("generate: %w", err)
		}
		out, derr := DecodeJSONObjectStrict[T](raw)
		if derr == nil {
			return out, nil
		}
		lastErr = derr
		var se *StructuralError
		if attempt < maxAttempts && errors.As(derr, &se) {
			repaired, rerr := j.repairer().RepairPrompt(current, se)
			if rerr != nil {
				return zero, rerr
			}
			current = repaired
			continue
		}
	}
	return zero, lastErr
}

func (j *ClaimJudge) generate(ctx context.Context, inst eval.Instance, step, prompt string) (string, error) {
	key := CacheKey{
		ProtocolDigest: j.Protocol.Digest,
		InstanceDigest: inst.Digest,
		Step:           step,
		PromptDigest:   promptDigest(prompt),
	}
	cache := j.cache()
	var cached string
	if hit, err := cache.Load(ctx, key, &cached); err == nil && hit {
		return cached, nil
	}
	res, err := j.Generate.Generate(ctx, GenerationRequest{
		Prompt:     prompt,
		ProtocolID: j.Protocol.Protocol.Name,
		Step:       step,
	})
	if err != nil {
		return "", err
	}
	if err := cache.Store(ctx, key, res.Text); err != nil {
		return "", fmt.Errorf("cache store: %w", err)
	}
	return res.Text, nil
}
