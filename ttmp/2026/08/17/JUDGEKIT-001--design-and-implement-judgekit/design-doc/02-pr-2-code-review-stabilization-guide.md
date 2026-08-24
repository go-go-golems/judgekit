---
Title: PR 2 Code Review Stabilization Guide
Ticket: JUDGEKIT-001
Status: active
Topics:
    - research
    - llm
    - evaluation
    - optimization
    - safety
DocType: design-doc
Intent: short-term
Owners: []
RelatedFiles:
    - Path: repo://.github/workflows/push.yml
      Note: Stale CI target blocking unit tests
    - Path: repo://audit/disagreement.go
      Note: Union-based missing-output comparison
    - Path: repo://audit/panel.go
      Note: Stable panel-member identity
    - Path: repo://audit/reliability.go
      Note: Fresh repeat execution requirements
    - Path: repo://calibration/report.go
      Note: Probability semantics and protocol-population checks
    - Path: repo://judging/claimjudge.go
      Note: Contract, protocol, evidence, prompt, and direct-dimension stabilization
    - Path: repo://suite/suite.go
      Note: Dependency-wave context lifecycle
ExternalSources:
    - https://github.com/go-go-golems/judgekit/pull/2
Summary: Intern-ready analysis and implementation plan for resolving PR 2 review findings without expanding into the fully hardened architecture.
LastUpdated: 2026-08-24T19:25:00-04:00
WhatFor: Stabilize PR 2 by fixing measurement correctness, attribution, audit, orchestration, and CI defects with focused regression tests.
WhenToUse: Before changing code in response to PR 2 review comments or deciding whether the pull request is ready to merge.
---


# PR 2 Code Review Stabilization Guide

## 1. Executive summary

Pull request 2 introduces the first complete judgekit implementation: eight provider-neutral core packages, documentation, a help-host CLI, examples, and ticket artifacts. Automated review found eleven unresolved threads, ten marked P1. The number of findings is not evidence that the package decomposition is fundamentally wrong. The findings mostly expose one repeated implementation pattern: declarations such as contract digests, evidence policies, protocol identities, confidence values, and judge identities were represented in data structures but were not always enforced at the point where a result was produced or consumed.

This guide deliberately proposes a **stabilization pass**, not a full architecture rewrite. The pull request should become truthful about the semantics it already exposes. It does not need private immutable types, signed records, a typestate system, or recursive verification at every function call. Those possibilities are documented separately in the hardened architecture guide.

The merge objective is:

```text
existing public API
  + focused semantic checks
  + explicit run behavior
  + adversarial regression tests
  + green CI
  = research-safe first merge
```

## 2. System orientation for a new intern

Judgekit separates five concerns that are easy to conflate:

1. `spec` defines **what** is measured.
2. `eval` describes **what concrete item** is being measured.
3. `protocol` identifies **how** the measurement is performed.
4. `assessment` represents **what the evaluator observed**.
5. `judging`, `audit`, `calibration`, and `suite` execute or analyze measurements.

The main flow is:

```text
spec.ContractDocument -----------+
                                  |
protocol.Document ---------------+--> judging.ClaimJudge
                                  |          |
eval.Instance -------------------+          v
                                      assessment.Report
                                             |
                        +--------------------+--------------------+
                        |                    |                    |
                    audit                calibration           suite
               consistency tests       gold comparison     evaluator graph
```

The important boundary is not merely whether each input is well-shaped. The inputs must agree with each other. A valid protocol can still be the wrong protocol for a valid contract. A valid report can still come from the wrong protocol for a calibration run. Stabilization therefore focuses on **cross-object invariants**.

## 3. Current source map

Start review in these files:

| Concern | Primary file | Why it matters |
|---|---|---|
| Contract schema | `spec/contract.go` | Evidence policy, labels, aggregations |
| Contract validation | `spec/validate.go` | Internal coherence of measurement definitions |
| Instance validation | `eval/validate.go` | Artifacts, evidence, required facts, instance identity |
| Instance digest | `eval/digest.go` | Canonical identity used by reports and cache keys |
| Protocol schema | `protocol/protocol.go` | Measurement, model, prompt, decoding identity |
| Protocol validation | `protocol/validate.go` | Internal coherence of protocol declarations |
| Report validation | `assessment/validate.go` | Claims, evidence references, dimensions |
| Judge execution | `judging/claimjudge.go` | Main contract/protocol/instance convergence point |
| Judge caching | `judging/cache.go` | Reuse keyed by protocol, instance, step, prompt |
| Reliability | `audit/reliability.go` | Runs base/variant probes |
| Comparisons | `audit/disagreement.go` | Determines what counts as instability |
| Panels | `audit/panel.go` | Preserves multiple judge reports and agreement |
| Calibration | `calibration/report.go` | Matches gold records to predictions |
| Suite runner | `suite/suite.go` | Dependency waves and concurrent execution |
| CI test workflow | `.github/workflows/push.yml` | Currently fails before unit tests |
| Security workflow | `.github/workflows/dependency-scanning.yml` | Govuln and dependency review failures |

## 4. Root-cause classification

### 4.1 Cross-object bindings are missing

Several types validate correctly in isolation but are never checked as a set. Examples include protocol-to-contract and calibration-report-to-protocol relationships.

```text
valid contract A + valid protocol for contract B != valid executable judge
valid report from protocol X + calibration request for protocol Y != valid calibration evidence
```

### 4.2 Contracts are not fully executable

`EvidencePolicy`, construct ranges, and construct labels are declared by `spec`, but the judging path does not consistently apply them. A contract that cannot reject forbidden evidence or an out-of-range direct score is documentation, not an operational measurement definition.

### 4.3 Runtime modes are implicit

Normal judging and reliability auditing use the same `Judge` interface. A cached call is desirable for ordinary evaluation but invalid for a repeatability probe. The caller needs a way to request fresh execution.

### 4.4 Comparison uses shared output only

Reliability currently compares intersections. If one report drops a difficult dimension, the missing result disappears from the denominator. Absence must be observable disagreement.

### 4.5 Some scalar fields have ambiguous semantics

`ClaimAssessment.Confidence` is confidence in the emitted verdict, but binary calibration treats it as the probability of `entailed`. Those meanings coincide only for positive verdicts.

## 5. Review issue matrix

| # | Review finding | Root invariant | Required stabilization |
|---:|---|---|---|
| 1 | Reused errgroup context is canceled after first wave | Context ownership is wave-scoped | Create a fresh errgroup/context per wave |
| 2 | Repeat reliability can hit cache | Repeat probes require fresh generations | Add cache bypass or isolated-cache execution |
| 3 | Negative verdict confidence calibrated backwards | Calibration needs target-class probability | Add explicit entailed probability or transform defined label probabilities |
| 4 | Evidence policy is decorative | Contract governs admitted evidence | Validate evidence against active contract before prompting |
| 5 | Direct dimensions ignore construct range/labels | Output must satisfy construct schema | Validate each direct result against its construct |
| 6 | Protocol can reference another contract | Protocol and contract must be bound | Compare `MeasurementDigest` with contract digest |
| 7 | Prompt declaration can differ from runtime prompt | Protocol must identify prompt implementation | Define template identity and rendered-content identity |
| 8 | Missing dimensions are ignored | Missing output is disagreement | Compare the union of IDs and record presence mismatch |
| 9 | Panel matrix keyed by instance | Matrix requires evaluator identity | Add named panel members |
| 10 | Instance digest is not recomputed | Cache/report identity must reflect actual input | Compute current instance identity at execution |
| 11 | Calibration can consume another protocol's reports | Calibration population must be homogeneous | Validate report protocol before consuming verdicts |

## 6. Implementation plan

### Phase A: restore CI signal

The test workflow currently stops before unit tests because `.github/workflows/push.yml` calls a Makefile target removed during repository normalization.

Actions:

1. Remove `make logcopter-check` from `.github/workflows/push.yml`.
2. Update the Go version/toolchain to 1.26.6.
3. Enable the repository dependency graph, or disable dependency review with a documented reason until it is enabled.
4. Run the same commands locally with `GOWORK=off`.

Validation:

```bash
GOWORK=off go generate ./...
git diff --exit-code
GOWORK=off go test -race -count=1 ./...
GOWORK=off go vet ./...
GOWORK=off golangci-lint run
govulncheck ./...
```

### Phase B: add judge preflight binding

Add one preflight path to `judging.ClaimJudge` rather than scattering checks throughout generation:

```go
func (j *ClaimJudge) validateFor(inst *eval.Instance) error {
    if j.Generate == nil || j.Prompts == nil {
        return errors.New("claim judge: generator and prompts are required")
    }
    if err := spec.ValidateContract(&j.Contract.Contract); err != nil {
        return fmt.Errorf("claim judge: contract: %w", err)
    }
    if err := protocol.Validate(&j.Protocol.Protocol); err != nil {
        return fmt.Errorf("claim judge: protocol: %w", err)
    }
    if j.Protocol.Protocol.MeasurementDigest != j.Contract.Digest {
        return fmt.Errorf("claim judge: protocol measurement digest does not match contract")
    }
    if err := eval.ValidateInstance(inst); err != nil {
        return fmt.Errorf("claim judge: instance: %w", err)
    }
    return validateEvidencePolicy(j.Contract.Contract.EvidencePolicy, inst.Evidence)
}
```

Call this before constructing prompts or consulting caches. No provider call should occur after a failed preflight.

### Phase C: make evidence and dimensions obey the contract

Implement narrow helpers in the package that owns the semantic definition:

```go
func ValidateEvidence(policy EvidencePolicy, items []eval.EvidenceItem) error
func ValidateDimension(c Construct, labels []string, d assessment.DimensionResult) error
```

Evidence pseudocode:

```text
allowed = set(policy.AllowedKinds)
forbidden = set(policy.ForbiddenKinds)
for each evidence item:
    if item.kind in forbidden: reject
    if allowed is nonempty and item.kind not in allowed: reject
    if require_provenance and provenance is empty: reject
```

Direct-dimension pseudocode:

```text
for every direct result:
    find construct by construct_id or reject
    reject duplicate output
    if numeric range declared:
        require numeric value and check min <= value <= max
    if labels declared:
        require label to belong to declared set
for every direct construct in contract:
    require exactly one direct result
```

### Phase D: correct confidence and calibration

The preferred research-stage API is explicit target-class probability:

```go
type ClaimAssessment struct {
    ClaimID             string
    Label               SupportLabel
    VerdictConfidence   *float64
    EntailedProbability *float64
    // evidence and reason...
}
```

Calibration consumes only `EntailedProbability`. If it is nil, Brier/ECE omit that claim. The confusion matrix still uses `Label`.

Before matching claims, verify:

```go
if report.ProtocolDigest != in.ProtocolDigest {
    return Report{}, fmt.Errorf("calibrate: instance %q used protocol %q, want %q", ...)
}
```

### Phase E: make audit runs genuinely fresh

Avoid changing every consumer at once. Add an optional execution interface that auditing can detect:

```go
type CacheMode string
const (
    CacheUse CacheMode = "use"
    CacheBypass CacheMode = "bypass"
)

type EvaluationOptions struct {
    CacheMode CacheMode
}

type ConfigurableJudge interface {
    EvaluateWithOptions(context.Context, eval.Instance, EvaluationOptions) (assessment.Report, error)
}
```

`audit.Reliability` requests `CacheBypass` for `Repeat` probes. If a judge cannot guarantee fresh execution, fail clearly rather than report repeat reliability.

### Phase F: fix report comparison and panel identity

Represent panel members explicitly:

```go
type PanelMember struct {
    ID    string
    Judge judging.Judge
}

type Panel struct {
    Members []PanelMember
    Policy  AggregationPolicy
}
```

Build disagreement input from the union of IDs:

```text
all_construct_ids = union(base dimensions, variant dimensions)
for id in all_construct_ids:
    if id missing from either side:
        record presence disagreement
    else compare value, label, and applicability

all_claim_ids = union(base claims, variant claims)
repeat the same presence rule
```

### Phase G: repair suite context ownership

Create a group inside each wave:

```go
for len(pending) > 0 {
    ready := findReady(pending, remaining)
    waveResults := snapshotResults(results)
    g, waveCtx := errgroup.WithContext(ctx)

    for _, evaluator := range ready {
        evaluator := evaluator
        g.Go(func() error {
            report, err := evaluator.Evaluate(waveCtx, inst, waveResults)
            // store under lock
            return err
        })
    }
    if err := g.Wait(); err != nil {
        return Results{}, err
    }
    markWaveComplete(ready, pending, remaining)
}
```

Do not mutate dependency state from worker goroutines. Mark the completed wave after `Wait`; this makes scheduling state single-threaded and easier to reason about.

### Phase H: simplify prompt identity

Do not compare a static protocol digest to a prompt containing instance text. Use:

```go
type ClaimProtocol interface {
    Version() string
    ExtractPrompt(eval.Instance) (string, error)
    SupportPrompt(eval.Instance, []assessment.Claim) (string, error)
}
```

The protocol pins `Version()` or a template digest. The cache key computes a fresh SHA-256 of the rendered prompt. This is enough for internal research and avoids a premature renderer framework.

## 7. Regression-test plan

Add one test per defect. Each test should fail on the reviewed commit and pass after the fix.

```text
judging:
  rejects mismatched protocol/contract
  rejects forbidden evidence before generation
  rejects missing required provenance
  rejects direct value outside range
  rejects direct label outside contract
  changes cache identity when instance content changes

calibration:
  rejects report from another protocol
  correctly scores contradicted verdict probability

audit:
  repeat probe bypasses cache
  missing dimension is disagreement
  missing claim is disagreement
  three panel members retain separate matrix rows

suite:
  dependent evaluator receives a live context after first wave

CI:
  test job reaches and runs unit tests
```

## 8. What not to add in this pull request

Do not add:

- private immutable wrappers;
- `VerifiedInstance` or `CompiledJudge` typestate;
- cryptographic signatures;
- provider adapters;
- advanced calibration slices;
- new panel aggregation algorithms;
- CoinVault changes;
- a generalized asynchronous DAG scheduler.

The review response should reduce risk and ambiguity, not add another framework layer.

## 9. Commit structure

Use focused commits so reviewers can map code to threads:

```text
fix(ci): restore test and vulnerability checks
fix(judging): bind contracts and enforce evidence/direct dimensions
fix(calibration): define probability semantics and protocol population
fix(audit): bypass cache and count missing output
fix(audit): give panel members stable identities
fix(suite): scope errgroups to dependency waves
fix(protocol): distinguish template and rendered prompt identity
test: add PR 2 adversarial regression suite
docs: align guarantees with research-stage behavior
```

## 10. Merge gate

PR 2 is ready when:

- all eleven threads are resolved with tests;
- all CI jobs are green or a repository-setting limitation is explicitly documented;
- no report can silently be attributed to the wrong contract or protocol;
- direct outputs and evidence obey the measurement contract;
- calibration consumes an explicit target probability;
- repeat reliability cannot use cached generations;
- missing output lowers reliability;
- suite dependents receive a live context;
- documentation no longer promises stronger immutability than the implementation provides.

## 11. Review checklist for the intern

Before requesting re-review:

- Read every thread at `https://github.com/go-go-golems/judgekit/pull/2`.
- Link each commit or regression test in the corresponding response.
- Confirm no fix merely suppresses an error.
- Confirm fake generators can prove that invalid preflight inputs make zero model calls.
- Run tests with and without the workspace (`GOWORK=off`).
- Inspect `git diff --stat` and keep unrelated ticket/docs edits out of stabilization commits.

## 12. References

- `judging/claimjudge.go`
- `judging/cache.go`
- `eval/validate.go`
- `eval/digest.go`
- `spec/contract.go`
- `spec/validate.go`
- `protocol/protocol.go`
- `assessment/validate.go`
- `calibration/report.go`
- `audit/reliability.go`
- `audit/disagreement.go`
- `audit/panel.go`
- `suite/suite.go`
- `.github/workflows/push.yml`
- `.github/workflows/dependency-scanning.yml`
- Pull request 2 review threads
