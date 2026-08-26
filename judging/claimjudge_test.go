package judging

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/protocol"
	"github.com/go-go-golems/judgekit/spec"
)

type testPrompts struct{}

var errExhausted = errors.New("sequence generator exhausted")

func (testPrompts) TemplateDigest(step string) (string, error) {
	switch step {
	case "extract", "support":
		return PromptDigest("test-prompts/" + step + "/v1"), nil
	default:
		return "", errors.New("unknown prompt step")
	}
}

func (testPrompts) ExtractPrompt(input ClaimExtractionInput) (string, error) {
	return "EXTRACT\nQ: " + input.Input.Text + "\nA: " + input.Candidate.Text, nil
}

func (testPrompts) SupportPrompt(inst eval.Instance, claims []assessment.Claim) (string, error) {
	var b strings.Builder
	b.WriteString("SUPPORT\n")
	for _, c := range claims {
		b.WriteString(c.ID + " " + c.Text + "\n")
	}
	return b.String(), nil
}

func buildContract(t *testing.T) spec.ContractDocument {
	t.Helper()
	c := spec.MeasurementContract{
		APIVersion: spec.ContractAPIVersion,
		Name:       "rag-faithfulness",
		Constructs: []spec.Construct{
			{ID: "faithfulness", Name: "Faithfulness", Definition: "fraction of claims entailed by evidence", Unit: "fraction", Direction: spec.Maximize, Range: &spec.Range{Minimum: 0, Maximum: 1}},
			{ID: "relevance", Name: "Relevance", Definition: "how well the answer addresses the question", Unit: "fraction", Direction: spec.Maximize, Range: &spec.Range{Minimum: 0, Maximum: 1}},
			{ID: "abstention", Name: "Abstention", Definition: "whether the answer abstains when ungrounded", Unit: "label", Direction: spec.Descriptive},
		},
		EvidencePolicy: spec.EvidencePolicy{
			AllowedKinds:   []string{"knowledge", "sql"},
			ForbiddenKinds: []string{"model_knowledge"},
		},
		Labels: map[spec.ConstructID][]string{
			"faithfulness": {"entailed", "contradicted", "insufficient"},
			"abstention":   {"attempted", "abstained"},
		},
		Aggregations: map[spec.ConstructID]spec.Aggregation{
			"faithfulness": {Method: spec.MethodFraction, Numerator: "entailed", Denominator: "entailed,contradicted,insufficient", EmptyPolicy: spec.EmptyVacuousPerfect},
			"relevance":    {Method: spec.MethodDirect, EmptyPolicy: spec.EmptyNA},
			"abstention":   {Method: spec.MethodDirect, EmptyPolicy: spec.EmptyNA},
		},
	}
	if err := spec.ValidateContract(&c); err != nil {
		t.Fatalf("validate contract: %v", err)
	}
	digest, err := spec.SemanticDigest(&c)
	if err != nil {
		t.Fatalf("contract digest: %v", err)
	}
	return spec.ContractDocument{Contract: c, Digest: digest}
}

func buildProtocol(t *testing.T, contract spec.ContractDocument, attempts int) protocol.Document {
	t.Helper()
	p := protocol.Protocol{
		APIVersion:        protocol.ProtocolAPIVersion,
		Name:              "gec-faithfulness-v1",
		MeasurementDigest: contract.Digest,
		Model:             protocol.ModelIdentity{Provider: "fake", Model: "fake-1"},
		PromptDigests: map[string]string{
			"extract": PromptDigest("test-prompts/extract/v1"),
			"support": PromptDigest("test-prompts/support/v1"),
		},
		Decoding:          protocol.DecodingPolicy{MaxTokens: 1024},
		EvidenceOrder:     protocol.EvidenceOrderAsGiven,
		ParserVersion:     "strict-json-v1",
		AggregatorVersion: "claim-fraction-v1",
		Retry:             protocol.RetryPolicy{MaximumAttempts: attempts},
	}
	if err := protocol.Validate(&p); err != nil {
		t.Fatalf("validate protocol: %v", err)
	}
	digest, err := protocol.SemanticDigest(&p)
	if err != nil {
		t.Fatalf("protocol digest: %v", err)
	}
	return protocol.Document{Protocol: p, Digest: digest}
}

func buildInstance(t *testing.T, contract spec.ContractDocument) eval.Instance {
	t.Helper()
	policyDigest, err := spec.EvidencePolicyDigest(&contract.Contract.EvidencePolicy)
	if err != nil {
		t.Fatalf("policy digest: %v", err)
	}
	set, err := eval.NewEvidenceSet([]eval.EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: eval.NewTextArtifact("text/plain", "Employees may carry over a maximum of five days."), SourceID: "doc-1"},
	}, policyDigest)
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	inst, err := eval.NewInstance(
		"inst-1",
		eval.NewTextArtifact("text/plain", "How much leave carries over?"),
		eval.NewTextArtifact("text/plain", "You may carry over all unused leave."),
		set, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	return inst
}

func fixedClock() time.Time {
	return time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
}

func findDim(t *testing.T, r assessment.Report, id string) assessment.DimensionResult {
	t.Helper()
	for _, d := range r.Dimensions {
		if string(d.ConstructID) == id {
			return d
		}
	}
	t.Fatalf("dimension %q not found", id)
	return assessment.DimensionResult{}
}

func TestClaimJudgeEndToEnd(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	inst := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":["the limit is five days","all leave carries over"]}`,
		"support": `{"verdicts":[{"claim":1,"label":"entailed","evidence_ids":["e1"],"reason":"e1 states the limit"},{"claim":2,"label":"contradicted","evidence_ids":["e1"],"reason":"e1 limits to five"}],"dimensions":[{"construct_id":"relevance","value":0.9},{"construct_id":"abstention","label":"attempted"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}

	report, err := judge.Evaluate(context.Background(), inst)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if len(report.Claims) != 2 {
		t.Fatalf("claims = %d, want 2", len(report.Claims))
	}
	faith := findDim(t, report, "faithfulness")
	if faith.Value == nil || *faith.Value != 0.5 {
		t.Errorf("faithfulness = %v, want 0.5", faith.Value)
	}
	if len(faith.EvidenceIDs) != 1 || faith.EvidenceIDs[0] != "e1" {
		t.Errorf("faithfulness evidence = %v, want [e1]", faith.EvidenceIDs)
	}
	rel := findDim(t, report, "relevance")
	if rel.Value == nil || *rel.Value != 0.9 {
		t.Errorf("relevance = %v, want 0.9", rel.Value)
	}
	abs := findDim(t, report, "abstention")
	if abs.Label != "attempted" {
		t.Errorf("abstention = %q, want attempted", abs.Label)
	}
	if report.ProtocolDigest != proto.Digest {
		t.Errorf("report protocol digest = %q, want %q", report.ProtocolDigest, proto.Digest)
	}
	if report.InstanceDigest != inst.Digest {
		t.Errorf("report instance digest = %q, want %q", report.InstanceDigest, inst.Digest)
	}
	// The extractor must not have seen the evidence.
	if len(gen.Calls) < 1 {
		t.Fatalf("no extract call recorded")
	}
	extractPrompt := gen.Calls[0].Prompt
	if strings.Contains(extractPrompt, "maximum of five days") {
		t.Errorf("extract prompt leaked evidence text:\n%s", extractPrompt)
	}
	if !strings.Contains(extractPrompt, "carry over all unused leave") {
		t.Errorf("extract prompt did not include the candidate answer")
	}
}

func TestClaimJudgeAbstentionIsVacuouslyPerfect(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	inst := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":[]}`,
		"support": `{"verdicts":[],"dimensions":[{"construct_id":"relevance","value":0.2},{"construct_id":"abstention","label":"abstained"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	report, err := judge.Evaluate(context.Background(), inst)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	faith := findDim(t, report, "faithfulness")
	if faith.Value == nil || *faith.Value != 1.0 {
		t.Errorf("faithfulness = %v, want 1.0 (vacuous perfect for zero claims)", faith.Value)
	}
}

func TestClaimJudgeCachesGenerations(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	inst := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":["a claim"]}`,
		"support": `{"verdicts":[{"claim":1,"label":"entailed","evidence_ids":["e1"],"reason":"ok"}],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"attempted"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Cache: NewMemoryCache(), Clock: fixedClock}
	if _, err := judge.Evaluate(context.Background(), inst); err != nil {
		t.Fatalf("first Evaluate: %v", err)
	}
	if len(gen.Calls) != 2 {
		t.Fatalf("expected 2 generations, got %d", len(gen.Calls))
	}
	// Second run over the same instance and prompts must hit the cache.
	if _, err := judge.Evaluate(context.Background(), inst); err != nil {
		t.Fatalf("second Evaluate: %v", err)
	}
	if len(gen.Calls) != 2 {
		t.Errorf("expected cache to prevent regeneration, but generations grew to %d", len(gen.Calls))
	}
}

// sequenceGen returns responses in order per step, to exercise repair.
type sequenceGen struct {
	seq   map[string][]string
	idx   map[string]int
	calls []string
}

func (s *sequenceGen) Generate(_ context.Context, req GenerationRequest) (GenerationResult, error) {
	s.calls = append(s.calls, req.Step)
	i := s.idx[req.Step]
	s.idx[req.Step]++
	if i >= len(s.seq[req.Step]) {
		return GenerationResult{}, errExhausted
	}
	return GenerationResult{Text: s.seq[req.Step][i], Model: protocol.ModelIdentity{Provider: "fake", Model: "fake-1"}}, nil
}

func TestClaimJudgeRepairsStructuralFailure(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 2) // allow one repair
	inst := buildInstance(t, contract)
	gen := &sequenceGen{
		seq: map[string][]string{
			// First extract response is malformed prose; the second is valid.
			"extract": {"sorry, here is text", `{"statements":["a claim"]}`},
			"support": {`{"verdicts":[{"claim":1,"label":"entailed","evidence_ids":["e1"],"reason":"ok"}],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"attempted"}]}`},
		},
		idx: map[string]int{},
	}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	report, err := judge.Evaluate(context.Background(), inst)
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	if len(report.Claims) != 1 {
		t.Fatalf("claims = %d, want 1", len(report.Claims))
	}
	// extract must have been called twice (malformed, then repaired).
	extractCalls := 0
	for _, step := range gen.calls {
		if step == "extract" {
			extractCalls++
		}
	}
	if extractCalls != 2 {
		t.Errorf("extract calls = %d, want 2 (one repair)", extractCalls)
	}
}

func TestClaimJudgeBypassesCacheWhenRequested(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	inst := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":["a claim"]}`,
		"support": `{"verdicts":[{"claim":1,"label":"entailed","evidence_ids":["e1"],"reason":"ok"}],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"attempted"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Cache: NewMemoryCache(), Clock: fixedClock}
	if _, err := judge.Evaluate(context.Background(), inst); err != nil {
		t.Fatalf("warm cache: %v", err)
	}
	if _, err := judge.EvaluateWithOptions(context.Background(), inst, EvaluationOptions{CacheMode: CacheBypass}); err != nil {
		t.Fatalf("bypass: %v", err)
	}
	if len(gen.Calls) != 4 {
		t.Errorf("generator calls = %d, want 4 after two fresh bypass stages", len(gen.Calls))
	}
}

func TestClaimJudgeRejectsMismatchedContractBeforeGeneration(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	proto.Protocol.MeasurementDigest = "sha256:another-contract"
	digest, _ := protocol.SemanticDigest(&proto.Protocol)
	proto.Digest = digest
	gen := &FakeGenerator{Responses: map[string]string{}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Errorf("accepted protocol pinned to another contract")
	}
	if len(gen.Calls) != 0 {
		t.Errorf("generator was called before binding failure")
	}
}

func TestClaimJudgeRejectsMismatchedPromptTemplateBeforeGeneration(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	proto.Protocol.PromptDigests["extract"] = PromptDigest("another-template")
	digest, _ := protocol.SemanticDigest(&proto.Protocol)
	proto.Digest = digest
	gen := &FakeGenerator{Responses: map[string]string{}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Errorf("accepted prompt renderer whose template identity differs from protocol")
	}
	if len(gen.Calls) != 0 {
		t.Errorf("generator was called before prompt binding failure")
	}
}

func TestClaimJudgeEnforcesEvidencePolicyBeforeGeneration(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	policyDigest, _ := spec.EvidencePolicyDigest(&contract.Contract.EvidencePolicy)
	set, err := eval.NewEvidenceSet([]eval.EvidenceItem{{
		ID: "e1", Kind: "model_knowledge", Content: eval.NewTextArtifact("text/plain", "unsafe"), SourceID: "model",
	}}, policyDigest)
	if err != nil {
		t.Fatalf("evidence set: %v", err)
	}
	inst, err := eval.NewInstance("inst-1", eval.NewTextArtifact("text/plain", "q"), eval.NewTextArtifact("text/plain", "a"), set, nil, nil, nil)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	gen := &FakeGenerator{Responses: map[string]string{}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen}
	if _, err := judge.Evaluate(context.Background(), inst); err == nil {
		t.Errorf("accepted forbidden evidence kind")
	}
	if len(gen.Calls) != 0 {
		t.Errorf("generator was called before evidence-policy failure")
	}
}

func TestClaimJudgeRejectsUnrelatedEvidencePolicyDigest(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	set, err := eval.NewEvidenceSet([]eval.EvidenceItem{{
		ID: "e1", Kind: "knowledge", Content: eval.NewTextArtifact("text/plain", "evidence"), SourceID: "doc",
	}}, "sha256:unrelated-policy")
	if err != nil {
		t.Fatalf("evidence set: %v", err)
	}
	inst, err := eval.NewInstance("inst-1", eval.NewTextArtifact("text/plain", "q"), eval.NewTextArtifact("text/plain", "a"), set, nil, nil, nil)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	gen := &FakeGenerator{Responses: map[string]string{}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen}
	if _, err := judge.Evaluate(context.Background(), inst); err == nil {
		t.Errorf("accepted evidence set admitted under another policy")
	}
	if len(gen.Calls) != 0 {
		t.Errorf("generator was called before policy digest failure")
	}
}

func TestClaimJudgeRequiresEvidenceProvenanceBeforeGeneration(t *testing.T) {
	contract := buildContract(t)
	contract.Contract.EvidencePolicy.RequireProvenance = true
	contract.Digest, _ = spec.SemanticDigest(&contract.Contract)
	proto := buildProtocol(t, contract, 1)
	gen := &FakeGenerator{Responses: map[string]string{}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Errorf("accepted evidence without required provenance")
	}
	if len(gen.Calls) != 0 {
		t.Errorf("generator was called before provenance failure")
	}
}

func TestClaimJudgeRejectsOutOfRangeDirectDimension(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":[]}`,
		"support": `{"verdicts":[],"dimensions":[{"construct_id":"relevance","value":100},{"construct_id":"abstention","label":"attempted"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Errorf("accepted direct dimension outside construct range")
	}
}

func TestClaimJudgeRejectsUndeclaredDirectLabel(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":[]}`,
		"support": `{"verdicts":[],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"bogus"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Errorf("accepted undeclared direct label")
	}
}

func TestClaimJudgeFailsClosedOnBadLabel(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	inst := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":["a claim"]}`,
		// "bogus" is not a valid support label; this is a semantic failure that
		// must fail closed rather than be repaired.
		"support": `{"verdicts":[{"claim":1,"label":"bogus","evidence_ids":["e1"],"reason":"x"}],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"attempted"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	if _, err := judge.Evaluate(context.Background(), inst); err == nil {
		t.Errorf("Evaluate accepted an invalid support label, want fail-closed error")
	}
}

func TestClaimJudgeRebindsCurrentInstanceIdentity(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	instance := buildInstance(t, contract)
	staleDigest := instance.Digest
	instance.Metadata = map[string]string{"arm": "changed-after-construction"}
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":[]}`,
		"support": `{"verdicts":[],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"abstained"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Clock: fixedClock}
	report, err := judge.Evaluate(context.Background(), instance)
	if err != nil {
		t.Fatalf("evaluate stale caller instance: %v", err)
	}
	if report.InstanceDigest == staleDigest {
		t.Fatal("report trusted stale caller instance digest")
	}
	if report.Provenance.InstanceDigest != report.InstanceDigest {
		t.Fatalf("provenance instance digest = %q, report = %q", report.Provenance.InstanceDigest, report.InstanceDigest)
	}
}

func TestClaimJudgeRecordsPromptModelAndCacheAttribution(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	instance := buildInstance(t, contract)
	gen := &FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":[]}`,
		"support": `{"verdicts":[],"dimensions":[{"construct_id":"relevance","value":0.5},{"construct_id":"abstention","label":"abstained"}]}`,
	}}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Cache: NewMemoryCache(), Clock: fixedClock}
	first, err := judge.Evaluate(context.Background(), instance)
	if err != nil {
		t.Fatalf("first evaluate: %v", err)
	}
	if first.Provenance.ContractDigest != contract.Digest || first.Provenance.ProtocolDigest != proto.Digest {
		t.Fatalf("unexpected contract/protocol provenance: %+v", first.Provenance)
	}
	if first.Provenance.CacheMode != string(CacheUse) || len(first.Provenance.Generations) != 2 {
		t.Fatalf("unexpected first-run provenance: %+v", first.Provenance)
	}
	for _, generation := range first.Provenance.Generations {
		if generation.CacheHit || generation.RenderedPromptDigest == "" {
			t.Fatalf("unexpected fresh generation provenance: %+v", generation)
		}
		if err := protocol.ValidateObservedModel(proto.Protocol.Model, generation.ObservedModel); err != nil {
			t.Fatalf("observed model: %v", err)
		}
	}

	second, err := judge.Evaluate(context.Background(), instance)
	if err != nil {
		t.Fatalf("cached evaluate: %v", err)
	}
	for _, generation := range second.Provenance.Generations {
		if !generation.CacheHit {
			t.Fatalf("expected cache hit provenance: %+v", generation)
		}
	}
}

func TestClaimJudgeRejectsObservedModelMismatch(t *testing.T) {
	contract := buildContract(t)
	proto := buildProtocol(t, contract, 1)
	gen := &FakeGenerator{
		Responses: map[string]string{"extract": `{"statements":[]}`},
		Model:     protocol.ModelIdentity{Provider: "fake", Model: "wrong-model"},
	}
	judge := &ClaimJudge{Contract: contract, Protocol: proto, Prompts: testPrompts{}, Generate: gen, Cache: NewMemoryCache()}
	if _, err := judge.Evaluate(context.Background(), buildInstance(t, contract)); err == nil {
		t.Fatal("accepted a generation attributed to another model")
	}
	if len(gen.Calls) != 1 {
		t.Fatalf("generator calls = %d, want one failed extract", len(gen.Calls))
	}
}
