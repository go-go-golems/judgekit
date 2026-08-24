// Package claimjudge_example shows the full judgekit loop with a fake
// generator: load a contract, build a protocol and instance, render the two
// prompts, run the two-stage claim judge, and read the sealed report.
//
// Run it with:
//
//	GOWORK=off go test ./examples/claim-judge -v
package claimjudge_example

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/judging"
	"github.com/go-go-golems/judgekit/protocol"
	"github.com/go-go-golems/judgekit/spec"
)

// examplePrompts renders the two-stage prompts. The extraction prompt never
// sees the evidence; the support prompt receives the claims and (through the
// instance) the evidence.
type examplePrompts struct{}

func (examplePrompts) TemplateDigest(step string) (string, error) {
	return judging.PromptDigest("example-prompts/" + step + "/v1"), nil
}

func (examplePrompts) ExtractPrompt(inst eval.Instance) (string, error) {
	return "Extract factual claims as {\"statements\":[...]}.\nQ: " + inst.Input.Text + "\nA: " + inst.Candidate.Text, nil
}

func (examplePrompts) SupportPrompt(inst eval.Instance, claims []assessment.Claim) (string, error) {
	var b strings.Builder
	b.WriteString("Judge each claim against the evidence as {\"verdicts\":[...],\"dimensions\":[...]}.\n")
	for _, c := range claims {
		b.WriteString(c.ID + ". " + c.Text + "\n")
	}
	for _, e := range inst.Evidence.Items {
		b.WriteString("[" + e.ID + "] " + e.Content.Text + "\n")
	}
	return b.String(), nil
}

func TestExampleClaimJudge(t *testing.T) {
	// 1. Load the measurement contract.
	contract, err := spec.LoadContract("contract.yaml")
	if err != nil {
		t.Fatalf("load contract: %v", err)
	}

	// 2. Build the protocol, pinning the contract by digest.
	p := protocol.Protocol{
		APIVersion:        protocol.ProtocolAPIVersion,
		Name:              "gec-faithfulness-v1",
		MeasurementDigest: contract.Digest,
		Model:             protocol.ModelIdentity{Provider: "fake", Model: "fake-1"},
		PromptDigests: map[string]string{
			"extract": judging.PromptDigest("example-prompts/extract/v1"),
			"support": judging.PromptDigest("example-prompts/support/v1"),
		},
		Decoding:          protocol.DecodingPolicy{MaxTokens: 1024},
		EvidenceOrder:     protocol.EvidenceOrderAsGiven,
		ParserVersion:     "strict-json-v1",
		AggregatorVersion: "claim-fraction-v1",
		Retry:             protocol.RetryPolicy{MaximumAttempts: 2},
	}
	if err := protocol.Validate(&p); err != nil {
		t.Fatalf("validate protocol: %v", err)
	}
	pdigest, err := protocol.SemanticDigest(&p)
	if err != nil {
		t.Fatalf("protocol digest: %v", err)
	}
	protoDoc := protocol.Document{Protocol: p, Digest: pdigest}

	// 3. Build the evaluation instance under the contract's evidence policy.
	policyDigest, err := spec.EvidencePolicyDigest(&contract.Contract.EvidencePolicy)
	if err != nil {
		t.Fatalf("evidence policy digest: %v", err)
	}
	evidence, err := eval.NewEvidenceSet([]eval.EvidenceItem{
		{
			ID:       "e1",
			Kind:     "knowledge",
			Content:  eval.NewTextArtifact("text/plain", "Employees may carry over a maximum of five days."),
			SourceID: "doc-1",
		},
	}, policyDigest)
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	instance, err := eval.NewInstance(
		"inst-1",
		eval.NewTextArtifact("text/plain", "How much leave carries over?"),
		eval.NewTextArtifact("text/plain", "You may carry over all unused leave."),
		evidence, nil, nil, nil,
	)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}

	// 4. Fake generator with canned model responses.
	gen := &judging.FakeGenerator{Responses: map[string]string{
		"extract": `{"statements":["the limit is five days","all leave carries over"]}`,
		"support": `{"verdicts":[
			{"claim":1,"label":"entailed","evidence_ids":["e1"],"reason":"e1 states the limit"},
			{"claim":2,"label":"contradicted","evidence_ids":["e1"],"reason":"e1 limits to five"}
		],"dimensions":[
			{"construct_id":"relevance","value":0.9},
			{"construct_id":"abstention","label":"attempted"}
		]}`,
	}}

	// 5. Run the two-stage claim judge.
	fixed := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	judge := &judging.ClaimJudge{
		Contract: contract,
		Protocol: protoDoc,
		Prompts:  examplePrompts{},
		Generate: gen,
		Cache:    judging.NewMemoryCache(),
		Clock:    func() time.Time { return fixed },
	}
	report, err := judge.Evaluate(context.Background(), instance)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	// 6. Read the sealed report.
	t.Logf("report digest: %s", report.Digest)
	t.Logf("instance digest: %s", report.InstanceDigest)
	t.Logf("protocol digest: %s", report.ProtocolDigest)
	for _, d := range report.Dimensions {
		t.Logf("dimension %s: applicable=%v value=%v label=%q evidence=%v",
			d.ConstructID, d.Applicable, d.Value, d.Label, d.EvidenceIDs)
	}

	// Sanity: faithfulness is 1 entailed of 2 claims = 0.5.
	for _, d := range report.Dimensions {
		if string(d.ConstructID) == "faithfulness" {
			if d.Value == nil || *d.Value != 0.5 {
				t.Errorf("faithfulness = %v, want 0.5", d.Value)
			}
		}
	}
}
