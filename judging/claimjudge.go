package judging

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
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
// the claims and (through the instance) the evidence. TemplateDigest returns
// the stable identity of the renderer/template for each step; the protocol
// pins these identities while cache keys separately hash rendered prompt text.
type ClaimExtractionInput struct {
	ID        string            `json:"id"`
	Input     eval.Artifact     `json:"input"`
	Candidate eval.Artifact     `json:"candidate"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ClaimProtocol receives a restricted extraction input that cannot expose
// evidence, references, or required facts. The support stage receives the full
// instance after claims have been fixed.
type ClaimProtocol interface {
	TemplateDigest(step string) (string, error)
	ExtractPrompt(input ClaimExtractionInput) (string, error)
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

var _ ConfigurableJudge = (*ClaimJudge)(nil)

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

// PromptDigest returns the content identity of prompt text. Protocol prompt
// digests identify stable templates/renderers; cache keys use this function on
// the fully rendered, instance-specific prompt.
func PromptDigest(prompt string) string {
	sum := sha256.Sum256([]byte(prompt))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// extractPayload is the model response for claim extraction.
type extractPayload struct {
	Statements []string `json:"statements"`
}

// supportVerdict is one per-claim verdict in the support response.
type supportVerdict struct {
	Claim               int                     `json:"claim"` // 1-based index
	Label               assessment.SupportLabel `json:"label"`
	EvidenceIDs         []string                `json:"evidence_ids"`
	Reason              string                  `json:"reason"`
	VerdictConfidence   *float64                `json:"verdict_confidence,omitempty"`
	EntailedProbability *float64                `json:"entailed_probability,omitempty"`
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

// Evaluate runs the two-stage judge with normal cache behavior.
func (j *ClaimJudge) Evaluate(ctx context.Context, inst eval.Instance) (assessment.Report, error) {
	return j.EvaluateWithOptions(ctx, inst, EvaluationOptions{CacheMode: CacheUse})
}

// EvaluateWithOptions runs the two-stage judge with explicit run-scoped cache
// behavior. It validates and binds the contract, protocol, prompt templates,
// evidence policy, and instance before any generation is attempted.
func (j *ClaimJudge) EvaluateWithOptions(ctx context.Context, inst eval.Instance, opts EvaluationOptions) (assessment.Report, error) {
	if opts.CacheMode == "" {
		opts.CacheMode = CacheUse
	}
	if opts.CacheMode != CacheUse && opts.CacheMode != CacheBypass {
		return assessment.Report{}, fmt.Errorf("claim judge: cache mode %q is not recognized", opts.CacheMode)
	}
	bound, err := eval.BindCurrentIdentity(inst)
	if err != nil {
		return assessment.Report{}, fmt.Errorf("claim judge: bind current instance identity: %w", err)
	}
	inst = bound
	if err := j.validateFor(&inst); err != nil {
		return assessment.Report{}, err
	}
	allowed := assessment.EvidenceIDSet(inst.Evidence.Items)
	started := j.now()
	collector := &runCollector{}

	claims, err := j.extractClaims(ctx, inst, opts, collector)
	if err != nil {
		return assessment.Report{}, err
	}
	results, directs, err := j.judgeSupport(ctx, inst, claims, opts, collector)
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
		Provenance: assessment.RunProvenance{
			ContractDigest:        j.Contract.Digest,
			ProtocolDigest:        j.Protocol.Digest,
			InstanceDigest:        inst.Digest,
			PromptTemplateDigests: maps.Clone(j.Protocol.Protocol.PromptDigests),
			ExpectedModel:         j.Protocol.Protocol.Model,
			CacheMode:             string(opts.CacheMode),
			Generations:           append([]assessment.PromptExecution(nil), collector.generations...),
		},
		StartedAt:  started,
		FinishedAt: j.now(),
	}
	if err := assessment.Seal(&report, allowed); err != nil {
		return assessment.Report{}, fmt.Errorf("claim judge: %w", err)
	}
	return report, nil
}

func (j *ClaimJudge) validateFor(inst *eval.Instance) error {
	if j.Generate == nil {
		return fmt.Errorf("claim judge: generator is required")
	}
	if j.Prompts == nil {
		return fmt.Errorf("claim judge: prompts are required")
	}
	if err := spec.ValidateContract(&j.Contract.Contract); err != nil {
		return fmt.Errorf("claim judge: invalid contract: %w", err)
	}
	contractDigest, err := spec.SemanticDigest(&j.Contract.Contract)
	if err != nil {
		return fmt.Errorf("claim judge: contract digest: %w", err)
	}
	if contractDigest != j.Contract.Digest {
		return fmt.Errorf("claim judge: contract digest %q does not match content (want %q)", j.Contract.Digest, contractDigest)
	}
	if err := protocol.Validate(&j.Protocol.Protocol); err != nil {
		return fmt.Errorf("claim judge: invalid protocol: %w", err)
	}
	protocolDigest, err := protocol.SemanticDigest(&j.Protocol.Protocol)
	if err != nil {
		return fmt.Errorf("claim judge: protocol digest: %w", err)
	}
	if protocolDigest != j.Protocol.Digest {
		return fmt.Errorf("claim judge: protocol digest %q does not match content (want %q)", j.Protocol.Digest, protocolDigest)
	}
	if j.Protocol.Protocol.MeasurementDigest != j.Contract.Digest {
		return fmt.Errorf("claim judge: protocol measurement digest %q does not match contract %q", j.Protocol.Protocol.MeasurementDigest, j.Contract.Digest)
	}
	for _, step := range []string{"extract", "support"} {
		got, err := j.Prompts.TemplateDigest(step)
		if err != nil {
			return fmt.Errorf("claim judge: %s template digest: %w", step, err)
		}
		want, ok := j.Protocol.Protocol.PromptDigests[step]
		if !ok {
			return fmt.Errorf("claim judge: protocol has no prompt digest for step %q", step)
		}
		if got != want {
			return fmt.Errorf("claim judge: %s template digest %q does not match protocol %q", step, got, want)
		}
	}
	if err := eval.ValidateInstance(inst); err != nil {
		return fmt.Errorf("claim judge: invalid instance: %w", err)
	}
	if err := j.validateEvidencePolicy(&inst.Evidence); err != nil {
		return fmt.Errorf("claim judge: evidence policy: %w", err)
	}
	return nil
}

func (j *ClaimJudge) validateEvidencePolicy(set *eval.EvidenceSet) error {
	policy := &j.Contract.Contract.EvidencePolicy
	want, err := spec.EvidencePolicyDigest(policy)
	if err != nil {
		return fmt.Errorf("compute digest: %w", err)
	}
	if set.PolicyDigest != want {
		return fmt.Errorf("policy digest %q does not match contract evidence policy %q", set.PolicyDigest, want)
	}
	allowed := make(map[string]struct{}, len(policy.AllowedKinds))
	for _, kind := range policy.AllowedKinds {
		allowed[kind] = struct{}{}
	}
	forbidden := make(map[string]struct{}, len(policy.ForbiddenKinds))
	for _, kind := range policy.ForbiddenKinds {
		forbidden[kind] = struct{}{}
	}
	for _, item := range set.Items {
		if _, denied := forbidden[item.Kind]; denied {
			return fmt.Errorf("evidence %q has forbidden kind %q", item.ID, item.Kind)
		}
		if len(allowed) > 0 {
			if _, ok := allowed[item.Kind]; !ok {
				return fmt.Errorf("evidence %q has kind %q not present in allowed_kinds", item.ID, item.Kind)
			}
		}
		if policy.RequireProvenance && len(item.Provenance) == 0 {
			return fmt.Errorf("evidence %q requires provenance", item.ID)
		}
	}
	return nil
}

type runCollector struct {
	generations []assessment.PromptExecution
}

func extractionInput(inst eval.Instance) ClaimExtractionInput {
	return ClaimExtractionInput{ID: inst.ID, Input: inst.Input, Candidate: inst.Candidate, Metadata: maps.Clone(inst.Metadata)}
}

func (j *ClaimJudge) extractClaims(ctx context.Context, inst eval.Instance, opts EvaluationOptions, collector *runCollector) ([]assessment.Claim, error) {
	prompt, err := j.Prompts.ExtractPrompt(extractionInput(inst))
	if err != nil {
		return nil, fmt.Errorf("extract prompt: %w", err)
	}
	payload, err := generateAndDecode[extractPayload](ctx, j, inst, "extract", prompt, opts, collector)
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

func (j *ClaimJudge) judgeSupport(ctx context.Context, inst eval.Instance, claims []assessment.Claim, opts EvaluationOptions, collector *runCollector) ([]assessment.ClaimAssessment, []directDimension, error) {
	prompt, err := j.Prompts.SupportPrompt(inst, claims)
	if err != nil {
		return nil, nil, fmt.Errorf("support prompt: %w", err)
	}
	payload, err := generateAndDecode[supportPayload](ctx, j, inst, "support", prompt, opts, collector)
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
			ClaimID:             claims[i].ID,
			Label:               v.Label,
			EvidenceIDs:         v.EvidenceIDs,
			Reason:              v.Reason,
			VerdictConfidence:   v.VerdictConfidence,
			EntailedProbability: v.EntailedProbability,
		}
	}
	return results, payload.Dimensions, nil
}

func (j *ClaimJudge) aggregate(claims []assessment.Claim, results []assessment.ClaimAssessment, directs []directDimension) ([]assessment.DimensionResult, error) {
	constructs := make(map[spec.ConstructID]*spec.Construct, len(j.Contract.Contract.Constructs))
	for i := range j.Contract.Contract.Constructs {
		c := &j.Contract.Contract.Constructs[i]
		constructs[c.ID] = c
	}
	directByConstruct := make(map[spec.ConstructID]directDimension, len(directs))
	for _, d := range directs {
		c, ok := constructs[d.ConstructID]
		if !ok {
			return nil, fmt.Errorf("aggregation: judge emitted unknown direct construct %q", d.ConstructID)
		}
		if j.Contract.Contract.Aggregations[d.ConstructID].Method != spec.MethodDirect {
			return nil, fmt.Errorf("aggregation: judge emitted direct result for non-direct construct %q", d.ConstructID)
		}
		if _, duplicate := directByConstruct[d.ConstructID]; duplicate {
			return nil, fmt.Errorf("aggregation: judge emitted duplicate direct construct %q", d.ConstructID)
		}
		if err := validateDirectDimension(c, j.Contract.Contract.Labels[c.ID], d); err != nil {
			return nil, err
		}
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

func validateDirectDimension(c *spec.Construct, labels []string, d directDimension) error {
	if d.Value == nil && strings.TrimSpace(d.Label) == "" {
		return fmt.Errorf("aggregation: direct construct %q emitted neither value nor label", c.ID)
	}
	if c.Range != nil {
		if d.Value == nil {
			return fmt.Errorf("aggregation: direct construct %q requires a numeric value", c.ID)
		}
		if *d.Value < c.Range.Minimum || *d.Value > c.Range.Maximum {
			return fmt.Errorf("aggregation: direct construct %q value %g is outside [%g,%g]", c.ID, *d.Value, c.Range.Minimum, c.Range.Maximum)
		}
	}
	if strings.TrimSpace(d.Label) != "" && len(labels) == 0 {
		return fmt.Errorf("aggregation: direct construct %q emitted label %q but the contract declares no labels", c.ID, d.Label)
	}
	if len(labels) > 0 {
		if strings.TrimSpace(d.Label) == "" {
			return fmt.Errorf("aggregation: direct construct %q requires a declared label", c.ID)
		}
		allowed := make(map[string]struct{}, len(labels))
		for _, label := range labels {
			allowed[label] = struct{}{}
		}
		if _, ok := allowed[d.Label]; !ok {
			return fmt.Errorf("aggregation: direct construct %q label %q is not declared", c.ID, d.Label)
		}
	}
	return nil
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
func generateAndDecode[T any](ctx context.Context, j *ClaimJudge, inst eval.Instance, step, prompt string, opts EvaluationOptions, collector *runCollector) (T, error) {
	var zero T
	maxAttempts := j.Protocol.Protocol.Retry.MaximumAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	current := prompt
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, cacheHit, err := j.generate(ctx, inst, step, current, opts)
		if err != nil {
			return zero, fmt.Errorf("generate: %w", err)
		}
		collector.generations = append(collector.generations, assessment.PromptExecution{
			Step:                 step,
			Attempt:              attempt,
			RenderedPromptDigest: PromptDigest(current),
			ObservedModel:        result.Model,
			CacheHit:             cacheHit,
			InputTokens:          result.InputTokens,
			OutputTokens:         result.OutputTokens,
			DurationNanos:        result.Duration.Nanoseconds(),
		})
		out, derr := DecodeJSONObjectStrict[T](result.Text)
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

func (j *ClaimJudge) generate(ctx context.Context, inst eval.Instance, step, prompt string, opts EvaluationOptions) (GenerationResult, bool, error) {
	key := CacheKey{
		ProtocolDigest: j.Protocol.Digest,
		InstanceDigest: inst.Digest,
		Step:           step,
		PromptDigest:   PromptDigest(prompt),
	}
	cache := j.cache()
	if opts.CacheMode != CacheBypass {
		var cached GenerationResult
		if hit, err := cache.Load(ctx, key, &cached); err != nil {
			return GenerationResult{}, false, fmt.Errorf("cache load: %w", err)
		} else if hit {
			if err := protocol.ValidateObservedModel(j.Protocol.Protocol.Model, cached.Model); err != nil {
				return GenerationResult{}, false, fmt.Errorf("cached generation: %w", err)
			}
			return cached, true, nil
		}
	}
	result, err := j.Generate.Generate(ctx, GenerationRequest{
		Prompt:     prompt,
		ProtocolID: j.Protocol.Protocol.Name,
		Step:       step,
	})
	if err != nil {
		return GenerationResult{}, false, err
	}
	if err := protocol.ValidateObservedModel(j.Protocol.Protocol.Model, result.Model); err != nil {
		return GenerationResult{}, false, err
	}
	if opts.CacheMode != CacheBypass {
		if err := cache.Store(ctx, key, result); err != nil {
			return GenerationResult{}, false, fmt.Errorf("cache store: %w", err)
		}
	}
	return result, false, nil
}
