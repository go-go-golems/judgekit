# Judgekit Glossary

This glossary defines the measurement-theory terms judgekit is built on. Each
term is grounded in *Language Models as Judges and Optimizers*, Chapter 1
("Foundations of Machine Judgment"), using the clarified reader edition kept
in this workspace's research ticket
(`coinvault/ttmp/2026/08/17/COINVAULT-045--study-self-optimization-and-exploitable-evaluator-errors/design-doc/03-chapter-1-reader-edition-foundations-of-machine-judgment.md`).
After each definition, an **In judgekit** note maps the concept to a concrete
type or package so an intern can connect the theory to the code.

The central idea: an evaluator is a *measurement instrument*, not an oracle. A
score is a proxy for a latent construct, and the relationship between proxy and
construct can break — especially under optimization. Judgekit makes the
intermediate steps explicit so the relationship can be audited.

---

## The inference chain

Judgekit models the middle of this chain (Chapter 1, §1.12 synthesis):

```text
abstract construct
  -> measurement contract        (spec)
  -> evaluation protocol         (protocol)
  -> evaluation instance + evidence (eval)
  -> structured assessment        (assessment)
  -> statistical audit/calibration (audit, calibration)
  -> application decision         (outside judgekit)
```

The application owns the construct's meaning and the final decision. Judgekit
makes the intermediate steps explicit, typed, versioned, and reproducible.

---

## Terms

### Construct
**Definition (1.1).** A construct is the abstract property an evaluation
intends to measure, such as factual correctness, faithfulness to evidence,
helpfulness, or search efficiency. Operationalizing a construct means turning
that abstract property into an explicit measurement procedure: what the
evaluator observes, what questions it answers, what evidence it may use, and
how observations become a reported result.

**In judgekit:** `spec.Construct` (ID, name, definition, unit, direction,
range). A construct is the *what*; a measurement contract operationalizes it.

### Latent utility
**Definition (1.2).** Latent utility is the target value the evaluation is
intended to represent but cannot observe directly in every case. It may be
objective (test success) or constructed from stakeholder preferences (a
tradeoff between completeness and cost).

**In judgekit:** judgekit never claims to observe latent utility directly. A
`assessment.DimensionResult.Value` is a *proxy*, not the utility itself.

### Proxy measurement
**Definition (1.3).** A proxy measurement is an observable score or verdict
used as evidence about latent utility. A proxy is useful to the extent that
its relationship to the target remains valid under the intended use, including
optimization.

**In judgekit:** every judge output is a proxy. The whole point of
`assessment.Report` carrying protocol and instance digests is to make "which
proxy, under which protocol, over which inputs" auditable.

### Measurement contract
The operational definition of one or more constructs: the constructs, allowed
evidence kinds, labels, aggregations, exclusions, and empty-case behavior.

**In judgekit:** `spec.MeasurementContract` + `spec.ContractDocument` (with a
semantic and byte digest). This is what an evaluator measures.

### Evaluation instance
One concrete item being judged: the task input, the candidate artifact, the
admitted evidence, an optional reference, optional required facts, and
metadata (Chapter 1, §1.3: `e = (x, y, c, r, m)`).

**In judgekit:** `eval.Instance` (with a content digest). This is what an
evaluator observes.

### Evaluation protocol
**Definition (1.12).** An evaluation protocol is the complete reproducible
procedure that turns task instances and candidates into reported measurements:
the evaluator model, prompt, allowed evidence, ordering, randomization,
decoding settings, schema validation, aggregation, calibration, and
statistical reporting rule. A model name alone is not an evaluator.

**In judgekit:** `protocol.Protocol` + `protocol.Document`. The protocol pins
the measurement contract by digest and adds model identity, prompt digests,
decoding, evidence ordering, parser/aggregator versions, and retry. A
one-token prompt change is a different protocol with a different digest.

### Judge
**Definition (1.4).** A judge maps an evaluation instance to an assessment used
for reporting or decision-making. Its output may be a scalar, category,
ranking, rationale, structured rubric, or distribution over verdicts.

**In judgekit:** `judging.Judge` (`Evaluate(ctx, eval.Instance) ->
assessment.Report`). `judging.ClaimJudge` is the reference two-stage judge.

### Critic
**Definition (1.5).** A critic returns diagnostic information intended to
explain a failure or guide a revision. It need not rank candidates or produce a
scalar reward.

**In judgekit:** `judging.Critic` (`Critique(...) -> Critique`).

### Verifier
**Definition (1.6).** A verifier checks a proposition, intermediate step,
constraint, or execution result. It is usually narrower than a general judge and
often has access to tools or formal evidence.

**In judgekit:** `judging.Verifier` (`Verify(...) -> VerificationResult`).

### Reward model
**Definition (1.7).** A reward model is a judge whose scalar output is consumed
by an optimization or selection algorithm. The defining property is its role as
an objective proxy, not its architecture.

**In judgekit:** judgekit deliberately does not provide a reward model. Turning
a `assessment.Report` into an optimization target is the application's (or
ragopt's) responsibility; judgekit produces auditable evidence, not deployment
decisions.

### Meta-evaluator
**Definition (1.8).** A meta-evaluator assesses the quality of judgments,
critiques, or evaluators. It may score a judge's reasoning, compare a verdict to
a reference, or select among judges.

**In judgekit:** the `audit` and `calibration` packages provide the
measurement-side of this (reliability, bias probes, confusion matrices).
Meta-evaluation of *reasoning* is out of scope for v0.

### Rubric
A rubric converts a broad construct into conditions the judge can apply. Good
rubrics separate dimensions, are observable, anchored, and actionable (Chapter
1, §1.3.1).

**In judgekit:** a rubric lives in the *application*, not in judgekit core. The
contract names constructs and labels; the application renders the prompt that
applies them. `protocol.Protocol.PromptDigests` pins each prompt renderer/template identity;
the cache separately hashes the fully rendered, instance-specific text.

### Rubric leakage
**Definition (1.13).** Rubric leakage occurs when the wording or examples in a
rubric disclose the desired answer or favor a candidate format in a way that
changes the task rather than merely measuring it.

**In judgekit:** a prompt-rendering concern owned by the application. The
two-stage `ClaimJudge` reduces one form of leakage by deriving claims before
seeing evidence, but rubric leakage in the prompt text is the application's
responsibility.

### Pointwise / pairwise / listwise judgment
**Definitions (1.9–1.11).** Pointwise evaluates one candidate against a rubric
or reference; pairwise compares two candidates; listwise jointly ranks three or
more. Comparative judgments are often easier but can be intransitive and amplify
position bias.

**In judgekit:** v0 focuses on pointwise two-stage claim judgment. Pairwise and
listwise protocols are future work; their interfaces will preserve candidate
identity separately from display order (Chapter 1, §1.10's `observed_order`
warning).

### Reliability
**Definition (1.16).** Reliability is the degree to which repeated measurements
of the same underlying object agree when irrelevant conditions vary (decoding
randomness, candidate order, prompt paraphrase, judge model, task sampling).
Reliability is consistency, not correctness.

**In judgekit:** the `audit` package (repeat/order/paraphrase probes,
disagreement reports). Cache determinism is *not* reliability — a cached wrong
verdict is stable but unreliable.

### Validity
**Definition (1.17).** Validity is the strength of the evidence that an
evaluation score supports the intended interpretation and use. It is a
property of an inference under a protocol, not an eternal property of a model.
Consequential validity asks what happens when people optimize against the
score.

**In judgekit:** judgekit supports validity by making the protocol explicit and
the report auditable. It does not certify validity; that requires calibration
data and human labels.

### Calibration
**Definition (1.18).** Calibration is agreement between stated confidence and
empirical frequency under a specified population. It concerns the meaning of
probabilities, not merely whether the most likely label is correct.

**In judgekit:** the `calibration` package (gold records, extraction
recall, confusion matrices, Brier score, ECE). A 1–5 ordinal score must not be
treated as a probability without an explicit calibrated mapping.

### Bias and nuisance variable
**Definition (1.19).** A nuisance variable is an observed or latent factor that
affects the measurement but is not part of the intended construct (position,
length, formatting, model identity, style). A bias is a systematic dependence
on such a feature, or a misweighting of a feature that should matter.

**In judgekit:** the `audit` bias probes (position, verbosity, style,
self-preference, authority). A probe must state what remains semantically
invariant.

### Construct shift / proxy breakdown under optimization
**Chapter 1, §1.6.1.** Optimization changes the distribution of candidate
outputs, and the old relationship between the judge's proxy and the intended
construct may no longer hold (a form of Goodhart's law). A judge valid for
retrospective reporting can be invalid as an optimization target.

**In judgekit:** the reason judgekit refuses to make promotion decisions.
Reports feed an external gate (ragopt/product) that must interpret them
explicitly; judgekit does not collapse "measured 0.86" into "deploy."

### De-anchored protocol
**Chapter 1, §1.8.4.** The judge first derives required facts or an answer key
from the question and evidence, then sees the candidate, so polished candidate
language does not anchor what the judge thinks the answer should contain.

**In judgekit:** the two-stage `ClaimJudge` is partially de-anchored: the
extractor derives claims from the candidate before the support judge sees
evidence. Full de-anchoring (answer key before candidate) is an application
prompt-design choice; judgekit provides the staging, not the prompt.

---

## How the terms map to packages

| Term | Package | Type |
| --- | --- | --- |
| Construct | `spec` | `Construct` |
| Measurement contract | `spec` | `MeasurementContract`, `ContractDocument` |
| Evaluation instance | `eval` | `Instance` |
| Evidence | `eval` | `EvidenceItem`, `EvidenceSet` |
| Required facts | `eval` | `RequiredFact` |
| Evaluation protocol | `protocol` | `Protocol`, `Document` |
| Model identity | `protocol` | `ModelIdentity` |
| Claim | `assessment` | `Claim` |
| Support verdict | `assessment` | `ClaimAssessment`, `SupportLabel` |
| Dimension result | `assessment` | `DimensionResult` |
| Assessment / report | `assessment` | `Report` |
| Judge | `judging` | `Judge`, `ClaimJudge` |
| Critic | `judging` | `Critic` |
| Verifier | `judging` | `Verifier` |
| Generator (provider-neutral) | `judging` | `Generator` |
| Evaluator suite | `suite` | `Suite`, `Evaluator`, `Results` |
| Reliability | `audit` | `ReliabilityReport`, `Probe`, `Panel` |
| Calibration | `calibration` | `GoldSet`, `Report`, `Confusion` |

---

## Further reading

- *Language Models as Judges and Optimizers*, Chapter 1 (reader edition) — the
  source of Definitions 1.1–1.19 above.
- `design-doc/01-judgekit-architecture-and-implementation-guide.md` in this
  ticket — the full architecture and decision records.
- `pkg/doc/` — the user guide, getting-started tutorial, and developer
  reference.
