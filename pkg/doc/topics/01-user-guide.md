---
Title: "judgekit User Guide"
Slug: user-guide
Short: "The conceptual model, package map, and invariants for evaluating language-model outputs with judgekit."
Topics:
- evaluation
- llm
- reference
Commands: []
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

judgekit is a provider-neutral Go library for evaluating language-model outputs
with explicit, typed, versioned, and auditable measurement. This guide explains
the conceptual model, the package map, and the invariants that keep evaluation
honest. It is the right place to start before reading the API reference or
running the getting-started tutorial.

## The problem judgekit solves

An LLM judge is a measurement instrument, not an oracle. A score is a proxy for
a latent construct (faithfulness, completeness, helpfulness), and the proxy can
respond to nuisance features — length, position, style — or break under
optimization. judgekit does not try to make judges correct. It makes the
evaluation protocol explicit so the relationship between proxy and construct
can be audited instead of trusted blindly.

The central boundary: judgekit produces auditable measurement evidence; it does
not grant that evidence authority it has not earned. It never decides whether
to deploy or promote.

## The inference chain

judgekit models the middle of this chain:

```text
abstract construct
  -> measurement contract        (spec)
  -> evaluation protocol         (protocol)
  -> evaluation instance + evidence (eval)
  -> structured assessment        (assessment)
  -> statistical audit/calibration (future)
  -> application decision         (outside judgekit)
```

The application owns the construct's meaning and the final decision. judgekit
makes the intermediate steps explicit, typed, and reproducible.

## What judgekit owns

- Construct and measurement-contract types (`spec`).
- Evaluation instances, artifacts, evidence sets, and provenance (`eval`).
- Evaluator protocol identity (`protocol`).
- Claims, three-way support labels, dimension results, and sealed reports
  (`assessment`).
- Provider-neutral judge, critic, verifier, and generator interfaces
  (`judging`).
- Strict structured-output parsing, repair, and caching (`judging`,
  `internal/strictdecode`).
- Deterministic semantic identities and strict file loaders (`internal/`).

## What judgekit does not own

- Product prompts, rubrics, tool names, authorization, or case schemas —
  applications own these.
- Optimization campaigns, candidate mutation, or promotion policy — that
  belongs to `ragopt` or the product.
- Documents, chunks, retrieval, reranking, or grounded-answer contracts — that
  belongs to `ragkit`.
- Provider configuration in core — adapters live in applications or optional
  `provider/` packages.
- A CLI workflow before the library is stable (the `cmd/judgekit` CLI only
  hosts help).
- Hidden chain-of-thought storage or any claim that a model score is ground
  truth.

## Package map

```text
spec/        what an evaluator measures
eval/        what an evaluator observes
protocol/    how an evaluator measures
assessment/  what an evaluator produces
judging/     running evaluators
internal/    canonicaljson, strictdecode, identifier helpers
```

Dependency direction: `spec`, `eval`, and `protocol` are parallel and depend
only on the standard library plus internal helpers. `assessment` depends on
`spec` and `eval`. `judging` depends on all four. A boundary test at the module
root rejects forbidden imports (frameworks, provider SDKs, sibling products)
in every core package.

## Key invariants

### A model name is never a protocol

"We used Model X as a judge" is incomplete. A one-token prompt change, a
different decoding seed, or a new parser version is a different protocol with a
different digest. `protocol.Protocol` captures model, prompt digests, decoding,
evidence ordering, parser and aggregator versions, and retry.

### Measurement and protocol are separate

Prompt or model changes do not always change the intended construct; construct
changes can hide inside prompts. judgekit keeps `spec.MeasurementContract` (what)
and `protocol.Protocol` (how) as separate documents with separate digests. A
report carries the protocol digest and thereby the contract indirectly.

### Support is three-way

A boolean cannot distinguish contradiction from absent evidence, and those
failures need different interventions. `assessment.SupportLabel` is `entailed`,
`contradicted`, or `insufficient`. Entailed and contradicted verdicts must cite
evidence; only `insufficient` may cite none.

### Evidence cannot be invented

The two-stage `judging.ClaimJudge` hides evidence from the extractor and gives
evidence to the support judge. The support judge may only cite evidence in the
instance's `eval.EvidenceSet`; a verdict citing an unknown evidence id fails
closed at seal time.

### Reports are sealed and content-addressed

`assessment.Report` carries the instance and protocol digests it was produced
under and its own content digest. After `assessment.Seal`, a report is treated
as immutable. A reader can prove which inputs and protocol produced a number.

### Caching is not reliability

Caching improves reproducibility and cost control. A cached wrong verdict is
stable but unreliable. Reliability is a future `audit` concern (repeat, order,
and paraphrase probes with disagreement reports).

## Two kinds of construct aggregation

A measurement contract declares how each construct is aggregated:

- **fraction** (for example faithfulness): counts claims by support label and
  computes numerator/denominator over declared labels. The empty case follows
  the contract's `empty_policy` (`vacuous_perfect`, `zero`, or `na`).
- **direct** (for example relevance, abstention): the support judge emits the
  whole-answer dimension directly; the judge takes it as-is.

Other methods (`mean`, `sum`, `count`, `label`) are declared but not yet wired
into the two-stage claim judge; they exist so contracts can name intent now.

## Fail-closed validation

judgekit validates strictly and fails closed: unsupported API versions,
duplicate construct IDs, dangling label/aggregation/exclusion references,
evidence-kind overlap, unknown JSON fields, trailing data, unknown evidence
references, invalid spans, and non-finite values all produce errors rather than
silent best-effort behavior. Only structural failures (bad JSON shape) are
repaired by default; semantic failures surface and stop.

## Troubleshooting

| Problem | Cause | Solution |
| --- | --- | --- |
| Reports differ across runs | Different protocol or instance | Compare `ProtocolDigest` and `InstanceDigest`. |
| A verdict cites unknown evidence | Support judge invented an id | Only cite ids in the instance's evidence set. |
| Digest changed unexpectedly | A semantic field changed | The digest is a function of content; re-pin the protocol. |
| Core won't build with a provider SDK | Boundary violation | Move the SDK import to an adapter, not a core package. |

## See Also

- `getting-started` for a runnable end-to-end tutorial.
- `developer-reference` for the public API and invariants.
- `GLOSSARY.md` at the repository root for measurement-theory definitions.
