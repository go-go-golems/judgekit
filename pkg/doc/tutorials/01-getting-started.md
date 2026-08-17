---
Title: "Getting Started with judgekit"
Slug: getting-started
Short: "Run a two-stage claim judge end to end with a fake generator and no provider credentials."
Topics:
- evaluation
- llm
- tutorial
Commands: []
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial walks through judgekit's core loop end to end: define what an
evaluator measures, build one evaluation instance, implement the two prompts
and a generator, run the two-stage claim judge, and read the sealed report. It
uses a fake generator so it runs with no provider credentials.

## What you will build

You will evaluate one answer to one question against one piece of evidence and
produce a report with three dimensions: evidence faithfulness (fraction of
claims entailed by the evidence), answer relevance, and abstention. The same
shape scales to thousands of instances; the point here is to see every link of
the inference chain.

## 1. Define a measurement contract

A measurement contract operationalizes an abstract construct ("faithfulness")
into something measurable: constructs, allowed evidence kinds, labels, and
aggregations. Save it as `contract.yaml`:

```yaml
api_version: judgekit.measurement/v1
name: rag-faithfulness
constructs:
  - id: faithfulness
    name: Evidence faithfulness
    definition: The fraction of extracted claims entailed by the evidence.
    unit: fraction
    direction: maximize
    range: { minimum: 0.0, maximum: 1.0 }
  - id: relevance
    name: Answer relevance
    definition: How well the answer addresses the question.
    unit: fraction
    direction: maximize
    range: { minimum: 0.0, maximum: 1.0 }
  - id: abstention
    name: Appropriate abstention
    definition: Whether the answer abstains when it cannot be grounded.
    unit: label
    direction: descriptive
evidence_policy:
  allowed_kinds: [knowledge]
  forbidden_kinds: [model_knowledge]
labels:
  faithfulness: [entailed, contradicted, insufficient]
aggregations:
  faithfulness:
    method: fraction
    numerator: entailed
    denominator: entailed,contradicted,insufficient
    empty_policy: vacuous_perfect
  relevance:    { method: direct, empty_policy: na }
  abstention:   { method: direct, empty_policy: na }
```

Load and validate it. The contract is content-addressed: `Digest` is its
semantic identity (order-independent), `ByteDigest` proves the exact file.

```go
contract, err := spec.LoadContract("contract.yaml")
if err != nil { log.Fatal(err)
}
```

If validation fails, judgekit tells you why: an unsupported `api_version`, a
duplicate construct id, a label referencing an unknown construct, or a fraction
denominator naming an undeclared label. Fix the contract; do not silence the
error.

## 2. Build an evaluation instance

An instance is one concrete item being judged: input, candidate, evidence,
optional required facts, and metadata. Use `eval.NewTextArtifact` so artifacts
are content-addressed automatically.

```go
evidence, err := eval.NewEvidenceSet([]eval.EvidenceItem{
    {
        ID:      "e1",
        Kind:    "knowledge",
        Content: eval.NewTextArtifact("text/plain", "Employees may carry over a maximum of five days."),
        SourceID: "doc-1",
    },
}, "sha256:policy")
if err != nil { log.Fatal(err) }

instance, err := eval.NewInstance(
    "inst-1",
    eval.NewTextArtifact("text/plain", "How much leave carries over?"),
    eval.NewTextArtifact("text/plain", "You may carry over all unused leave."),
    evidence, nil, nil, nil,
)
if err != nil { log.Fatal(err) }
```

## 3. Describe the protocol

A protocol is the complete reproducible instrument: model, prompt digests,
decoding, evidence ordering, parser and aggregator versions, and retry. A model
name alone is never a protocol.

```go
p := protocol.Protocol{
    APIVersion:        protocol.ProtocolAPIVersion,
    Name:              "gec-faithfulness-v1",
    MeasurementDigest: contract.Digest,
    Model:             protocol.ModelIdentity{Provider: "fake", Model: "fake-1"},
    PromptDigests:     map[string]string{"extract": "sha256:1", "support": "sha256:2"},
    Decoding:          protocol.DecodingPolicy{MaxTokens: 1024},
    EvidenceOrder:     protocol.EvidenceOrderAsGiven,
    ParserVersion:     "strict-json-v1",
    AggregatorVersion: "claim-fraction-v1",
    Retry:             protocol.RetryPolicy{MaximumAttempts: 2},
}
if err := protocol.Validate(&p); err != nil { log.Fatal(err) }
pdigest, err := protocol.SemanticDigest(&p)
if err != nil { log.Fatal(err) }
protoDoc := protocol.Document{Protocol: p, Digest: pdigest}
```

## 4. Implement the prompts and a generator

Implement `judging.ClaimProtocol` to render the two prompts. The extraction
prompt must not reveal the evidence; the support prompt receives the claims and
(through the instance) the evidence. This separation prevents the extractor
from biasing claims toward what is supported.

```go
type myPrompts struct{}

func (myPrompts) ExtractPrompt(inst eval.Instance) (string, error) {
    return "Extract factual claims as {\"statements\":[...]}.\nQ: " + inst.Input.Text + "\nA: " + inst.Candidate.Text, nil
}

func (myPrompts) SupportPrompt(inst eval.Instance, claims []assessment.Claim) (string, error) {
    // render claims + evidence; require {\"verdicts\":[...],\"dimensions\":[...]}
    // ...
    return prompt, nil
}
```

For this tutorial, use the fake generator with canned responses:

```go
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
```

In production you implement `judging.Generator` to call your model and return
its text. Core never imports a provider SDK.

## 5. Run the judge and read the report

```go
judge := &judging.ClaimJudge{
    Contract: contract,
    Protocol: protoDoc,
    Prompts:  myPrompts{},
    Generate: gen,
    Cache:    judging.NewMemoryCache(),
}
report, err := judge.Evaluate(context.Background(), instance)
if err != nil { log.Fatal(err) }
for _, d := range report.Dimensions {
    fmt.Printf("%s: applicable=%v value=%v label=%q\n", d.ConstructID, d.Applicable, d.Value, d.Label)
}
```

The report is sealed: `report.Digest` is a function of its content, and it
carries `report.InstanceDigest` and `report.ProtocolDigest` so a reader can
prove which inputs and protocol produced it.

## Why each step matters

- The **contract** makes "faithfulness" measurable instead of underspecified.
- The **instance** captures everything an evaluator observes, content-addressed.
- The **protocol** makes the instrument reproducible; changing a prompt is a new
  protocol.
- **Two stages** keep extraction honest (no evidence) and support grounded
  (cited evidence only).
- The **sealed report** makes the whole chain auditable.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| `unsupported extension` | Loader got a non-`.yaml/.json` file | Use `.yaml`, `.yml`, or `.json`. |
| `duplicate construct id` | Two constructs share an ID | Construct IDs must be unique. |
| `fraction ... undeclared label` | Numerator/denominator names a label not in `labels` | Add the label to the construct's `labels` list. |
| `cites unknown evidence` | A verdict cited an id not in the instance's evidence | The support judge may only cite evidence in `eval.EvidenceSet`. |
| `entailed verdict requires evidence_ids` | An entailed/contradicted verdict cited nothing | Only `insufficient` may cite no evidence. |
| `no response for step` | The fake generator lacks a step's response | Add the step's canned response. |

## See Also

- `user-guide` for the conceptual model and package map.
- `developer-reference` for the public API, invariants, and integration boundaries.
- `GLOSSARY.md` at the repository root for measurement-theory definitions.
