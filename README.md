# judgekit

A provider-neutral Go library for evaluating language-model outputs with
explicit, typed, versioned, and auditable measurement.

judgekit makes machine evaluation explicit without pretending to make it
infallible. It separates **what** an evaluator measures (a construct and its
measurement contract), **how** it measures (a protocol: model, prompts,
decoding, ordering, parser, retry), **what** it observes (an evaluation
instance and evidence), and **what** it produces (a structured, sealed
assessment). A reader can follow the chain from construct to reported number
and see every assumption that produced it.

judgekit is **library-first**. Core packages depend only on the standard
library and internal helpers — never on Glazed, Cobra, Bubble Tea, a provider
SDK, or a sibling product. A thin `cmd/judgekit` CLI exists only to host the
help system.

## What problem does judgekit solve?

An LLM judge is a *measurement instrument*, not an oracle (see
[`GLOSSARY.md`](GLOSSARY.md)). A score is a proxy for a latent construct, and
the proxy can respond to nuisance features (length, position, style) or break
under optimization. Judgekit does not try to make judges correct; it makes the
evaluation *protocol* explicit so that:

- a model name is never treated as a complete evaluator;
- a prompt or decoding change is a different protocol with a different digest;
- a verdict must cite real evidence and cannot invent it;
- contradiction is never conflated with absent evidence (three-way support);
- reports are content-addressed and carry the protocol and instance they were
  produced under;
- evaluation evidence and deployment decisions stay separate (judgekit
  measures; it does not promote).

## What judgekit deliberately does not do

- It does not own product prompts, rubrics, tool names, authorization, or case
  schemas — those belong to applications.
- It does not run optimization campaigns, mutate candidates, or decide
  promotions — that belongs to `ragopt` or the product.
- It does not own retrieval, chunks, reranking, or grounded-answer contracts —
  that belongs to `ragkit`.
- It does not store hidden chain-of-thought or treat any model score as ground
  truth.
- It does not provide a CLI workflow before the library is stable.

## Core packages

```text
judgekit/
  spec/        what an evaluator measures: constructs and measurement contracts
  eval/        what an evaluator observes: artifacts, evidence, instances
  protocol/    how an evaluator measures: model, prompts, decoding, ordering, retry
  assessment/  what an evaluator produces: claims, three-way support, dimensions, reports
  judging/     running evaluators: provider-neutral interfaces and the two-stage claim judge
  audit/       reliability and bias: probes, disagreement reports, panels
  calibration/ calibration: gold records, extraction recall, confusion, Brier, ECE
  suite/       combining evaluators: acyclic graph, concurrent execution
  internal/    canonicaljson, strictdecode, identifier helpers (not importable outside the module)
  cmd/judgekit/  thin CLI hosting the Glazed help system
  pkg/doc/     user guide, getting-started tutorial, developer reference (Glazed help entries)
```

Dependency direction: `spec`, `eval`, and `protocol` are parallel and depend
only on stdlib + internal helpers. `assessment` depends on `spec` and `eval`.
`judging` depends on all four. A boundary test at the module root rejects
forbidden imports (frameworks, provider SDKs, sibling products) in every core
package.

## Quick start

### Define a measurement contract

`examples/claim-judge/contract.yaml`:

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
evidence_policy:
  allowed_kinds: [knowledge, sql]
  forbidden_kinds: [model_knowledge]
labels:
  faithfulness: [entailed, contradicted, insufficient]
aggregations:
  faithfulness:
    method: fraction
    numerator: entailed
    denominator: entailed,contradicted,insufficient
    empty_policy: vacuous_perfect
  relevance:
    method: direct
    empty_policy: na
```

Load and validate it:

```go
doc, err := spec.LoadContract("contract.yaml")
// doc.Digest is the semantic identity; doc.ByteDigest proves the exact file.
```

### Implement a generator adapter

Applications adapt their model runtime to a provider-neutral interface. Core
never imports a provider SDK.

```go
type myGenerator struct{ /* ... */ }

func (g *myGenerator) Generate(ctx context.Context, req judging.GenerationRequest) (judging.GenerationResult, error) {
    // call your model here, return its text
    return judging.GenerationResult{Text: raw, Model: protocol.ModelIdentity{Provider: "mine", Model: "x"}}, nil
}
```

For tests and examples, use `judging.FakeGenerator`, which needs no credentials.

### Run a two-stage claim judge

```go
judge := &judging.ClaimJudge{
    Contract: contractDoc,
    Protocol: protocolDoc,
    Prompts:  myPrompts{},      // implements ExtractPrompt and SupportPrompt
    Generate: gen,
    Cache:    judging.NewMemoryCache(),
}
report, err := judge.Evaluate(ctx, instance)
// report is sealed; report.Dimensions has faithfulness, relevance, ...
```

The extractor never sees evidence; the support judge may only cite evidence in
the instance; the report is sealed with a content-addressed digest.

## Calibrating a protocol

Calibration links judgekit reports to human or objective labels. The
`calibration` package computes extraction recall (a judge can look accurate by
extracting fewer claims), confusion matrices, sensitivity, specificity, the
false-support rate, the Brier score, and expected calibration error. Brier
and ECE apply only when the protocol emits confidence probabilities; a 1-5
ordinal score is not a probability.

```go
report, err := calibration.Calibrate(calibration.CalibrateInput{
    Gold:           goldSet,          // human-labeled claims, content-addressed
    Reports:        reportsByInstance, // the model's reports keyed by instance ID
    ProtocolDigest: protoDoc.Digest,
    Bins:           10,
})
// report.ExtractionRecall, report.Sensitivity, report.BrierScore, report.ECE
```

Reliability is separate from caching: a cached wrong verdict is stable but
unreliable. The `audit` package runs a judge over base/variant instances that
differ only in something that should not affect the construct, compares the
reports, and reports per-construct agreement and mean absolute delta - never
one "reliability score". See `GLOSSARY.md` and the design doc for the full API.

## Help and docs

```bash
go run ./cmd/judgekit help getting-started
go run ./cmd/judgekit help user-guide
go run ./cmd/judgekit help developer-reference
```

## Development

```bash
GOWORK=off go test ./...     # all tests, including the boundary test
GOWORK=off go build ./...
make lint
```

Tests use only fake generators and local fixtures; no provider credentials are
required. See [`AGENT.md`](AGENT.md) for layout and conventions and the
[`ttmp/`](ttmp/) ticket `JUDGEKIT-001` for the design and investigation diary.

## Status

v0 implements the core value packages (`spec`, `eval`, `protocol`,
`assessment`), the two-stage `judging.ClaimJudge`, the `audit` reliability/bias
package, the `calibration` package, and the `suite` package for combining
evaluators. Optional provider adapters are the main remaining piece and are
described in the design doc. No public API stability is promised before the
first CoinVault pilot.
