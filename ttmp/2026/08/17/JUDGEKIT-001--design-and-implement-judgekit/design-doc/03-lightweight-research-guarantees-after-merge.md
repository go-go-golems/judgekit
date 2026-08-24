---
Title: Lightweight Research Guarantees After Merge
Ticket: JUDGEKIT-001
Status: active
Topics:
    - research
    - llm
    - evaluation
    - optimization
    - safety
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://assessment/report.go
      Note: Durable report snapshot
    - Path: repo://audit/reliability.go
      Note: Research freshness and comparison semantics
    - Path: repo://calibration/report.go
      Note: Research calibration consumer
    - Path: repo://eval/digest.go
      Note: Current stored identity implementation to simplify
    - Path: repo://judging/cache.go
      Note: Cache identity and future run modes
    - Path: repo://judging/interfaces.go
      Note: Provider-neutral execution boundary
    - Path: repo://protocol/protocol.go
      Note: Protocol and prompt identity model
ExternalSources: []
Summary: Post-merge design for simplifying judgekit around trustworthy research attribution without production-grade immutability machinery.
LastUpdated: 2026-08-24T19:25:00-04:00
WhatFor: Define the practical guarantee level judgekit should implement after PR 2 while it remains an internally operated research library.
WhenToUse: After PR 2 merges, before the CoinVault pilot and before adding new identity, persistence, or protocol APIs.
---


# Lightweight Research Guarantees After Merge

## 1. Executive summary

Judgekit is currently an internal research library operated by engineers who can coordinate changes and inspect artifacts. It does not need to defend against malicious mutation, hostile storage, or untrusted plugin authors. It does need to prevent experiments from silently combining the wrong protocol, stale input identity, cached repeat judgments, or mathematically ambiguous confidence values.

The lightweight design adopts this rule:

> Reject mistakes that change the meaning or attribution of a measurement. Do not defend every in-memory value against mutation by trusted research code.

This approach simplifies the implementation in three ways:

- Digests are computed at execution and persistence boundaries rather than treated as continuously verified proofs.
- Public structs remain ordinary Go values; no typestate or immutable wrapper hierarchy is introduced.
- Cross-object semantic checks are centralized in a small research-run preflight.

The result should support reproducible CoinVault experiments without turning judgekit into an integrity platform.

## 2. Threat and error model

### 2.1 Assumptions

The lightweight design assumes:

- repository contributors are trusted;
- callers are Go code owned by the same organization;
- experiment artifacts are stored in controlled local or project storage;
- accidental errors are more likely than malicious tampering;
- API evolution is acceptable before a stable release;
- a human can inspect unusual runs and regenerate them.

### 2.2 Errors that still matter

Trust does not prevent:

- loading the wrong protocol file;
- pairing a protocol with another contract;
- changing an instance after its cache key was computed;
- running a repeat probe through a cache;
- comparing only dimensions that survived both runs;
- calibrating one judge's reports under another judge's name;
- treating verdict confidence as a target-class probability;
- losing provenance needed to explain an unexpected score.

These can invalidate a research conclusion even when every participant is honest.

### 2.3 Explicit non-goals

The lightweight version does not guarantee:

- tamper-evident storage;
- cross-organization artifact verification;
- immutable values after construction;
- signed protocols or reports;
- adversarial plugin isolation;
- long-term binary compatibility;
- automatic migration across schema versions.

## 3. Guarantee levels

Use three levels instead of treating all values as equally sensitive.

### Level 1: structural validity

A value is structurally valid when required fields, ranges, identifiers, references, and API versions are correct.

Examples:

- a claim ID is nonblank and unique;
- evidence references resolve;
- a probability lies in `[0,1]`;
- a contract aggregation references declared labels.

This is handled by existing `Validate...` functions.

### Level 2: research attribution

A run is correctly attributed when the actual instance, contract, protocol, prompt version, and output population are recorded consistently.

Examples:

- protocol measurement digest matches the active contract;
- report protocol digest matches the calibration request;
- cache key uses the current instance content;
- report records the current protocol and rendered prompt identity.

This is the main lightweight guarantee.

### Level 3: hardened integrity

A value is hardened when it is immutable or recursively verified, signed, and safe to exchange across trust boundaries. This is intentionally deferred and described in the hardened guide.

```text
Level 1: shape is valid
          |
          v
Level 2: experiment meaning and attribution are correct   <-- target
          |
          v
Level 3: hostile mutation/storage is detectable           <-- deferred
```

## 4. Simplified identity model

### 4.1 Compute identity when it is used

Mutable in-memory structs should not rely on a digest computed hours earlier. Prefer pure functions or methods:

```go
func (i Instance) ContentDigest() (string, error)
func (s EvidenceSet) ContentDigest() (string, error)
func (c MeasurementContract) SemanticDigest() (string, error)
func (p Protocol) SemanticDigest() (string, error)
```

At execution time:

```go
instanceDigest, err := inst.ContentDigest()
contractDigest, err := contract.SemanticDigest()
protocolDigest, err := protocol.SemanticDigest()
```

The cache and report use these freshly computed values. Mutation therefore produces a new identity instead of a stale stored field.

### 4.2 Store identity only for snapshots

Persisted reports and exported protocol/contract documents may retain digest fields because they represent snapshots.

```go
type Report struct {
    InstanceDigest string
    ProtocolDigest string
    Digest         string
    // observations...
}
```

When loading a persisted report for calibration, verify its own digest once. During ordinary in-memory transformations, do not recursively verify it after every function call.

### 4.3 Reduce duplicate identity fields where possible

A research instance does not need both mutable content and a caller-maintained `Digest` field. After PR 2, consider removing stored digest fields from mutable inputs in a breaking v0 change:

```go
// Preferred research value.
type Instance struct {
    ID            string
    Input         Artifact
    Candidate     Artifact
    Evidence      EvidenceSet
    RequiredFacts []RequiredFact
    Metadata      map[string]string
}
```

The report remains the durable snapshot linking the run to the computed identity.

## 5. Research-run preflight

Centralize semantic checks in one function invoked immediately before generation:

```go
type ResearchRun struct {
    Contract spec.ContractDocument
    Protocol protocol.Document
    Prompts  judging.ClaimProtocol
    Judge    *judging.ClaimJudge
}

type RunIdentity struct {
    ContractDigest string
    ProtocolDigest string
    InstanceDigest string
    PromptVersion  string
}

func (r ResearchRun) Prepare(inst eval.Instance) (RunIdentity, error)
```

Pseudocode:

```text
validate contract structure
validate protocol structure
validate instance structure

compute contract digest from current contract
require protocol.measurement_digest == contract digest

validate evidence kinds and provenance under contract
compute current instance digest
compute current protocol digest
obtain stable prompt version

return RunIdentity
```

The generator receives only a prepared run. This avoids a full `VerifiedInstance` type while still making the critical execution boundary explicit.

## 6. Prompt and cache identity

### 6.1 Stable prompt version

For internal research, prompt identity can be an application-owned version string:

```go
type ClaimProtocol interface {
    Version() string
    ExtractPrompt(eval.Instance) (string, error)
    SupportPrompt(eval.Instance, []assessment.Claim) (string, error)
}
```

CoinVault might return:

```text
coinvault-claim-judge/v3
```

A prompt change must update this version. Code review and tests enforce the convention; judgekit does not attempt to hash executable Go behavior.

### 6.2 Rendered prompt digest

After rendering, compute a separate content digest:

```go
rendered := prompts.ExtractPrompt(inst)
renderedDigest := sha256(rendered)
```

Cache key:

```go
type CacheKey struct {
    ProtocolDigest string
    InstanceDigest string
    Step           string
    PromptVersion  string
    RenderedDigest string
}
```

The version explains the intended template revision; the rendered digest identifies the exact bytes sent for this instance.

### 6.3 Cache modes

Research orchestration must be able to choose:

```go
type CacheMode string
const (
    CacheUse      CacheMode = "use"
    CacheRefresh  CacheMode = "refresh"
    CacheBypass   CacheMode = "bypass"
)
```

- `use`: load and store normally.
- `refresh`: skip load, generate, then replace/store.
- `bypass`: neither load nor store.

Repeat reliability uses `bypass`. Ordinary development uses `use`. Regenerating a known case uses `refresh`.

## 7. Simplified probability model

Retain the verdict and its confidence as diagnostic information, but expose calibration input separately:

```go
type ClaimAssessment struct {
    ClaimID             string
    Label               SupportLabel
    VerdictConfidence   *float64
    EntailedProbability *float64
    EvidenceIDs         []string
    Reason              string
}
```

Rules:

- `Label` drives confusion matrices.
- `EntailedProbability` drives binary Brier/ECE.
- `VerdictConfidence` is descriptive unless a calibration mapping is defined.
- Missing probability yields nil Brier/ECE contributions, not invented values.

For future three-class calibration, add a distribution only when a protocol actually emits it:

```go
type SupportProbabilities struct {
    Entailed     float64
    Contradicted float64
    Insufficient float64
}
```

Do not add this to v0 merely for symmetry.

## 8. Executable research contract

A lightweight contract must still make the promised measurement operational.

### 8.1 Evidence admission

```go
func AdmitEvidence(policy spec.EvidencePolicy, set eval.EvidenceSet) error
```

This checks kinds and provenance. It does not authenticate sources or verify that provenance itself is truthful.

### 8.2 Direct output validation

```go
func ValidateDirectResult(
    construct spec.Construct,
    labels []string,
    result assessment.DimensionResult,
) error
```

This checks range, labels, applicability, and finiteness. It does not prove the model used the rubric correctly.

### 8.3 Report provenance

A report records:

```go
type RunProvenance struct {
    ContractDigest string
    ProtocolDigest string
    InstanceDigest string
    PromptVersion  string
    PromptDigests  map[string]string
    CacheMode      CacheMode
}
```

This is enough to explain and repeat an internal experiment.

## 9. Audit model

### 9.1 Freshness is a protocol condition

Reliability results must record how they were generated:

```go
type ProbeExecution struct {
    ProbeID       string
    BaseCacheMode CacheMode
    VariantCacheMode CacheMode
}
```

For repeat probes, both must be `bypass`. For order or formatting probes, either bypass or refresh is acceptable as long as the base and variant are not sharing a cached generation under the same key.

### 9.2 Missing output is observable

Comparison is over unions:

```text
same value/label and present on both sides -> agreement
present on both but different              -> disagreement
present only on base                       -> missing_variant
present only on variant                    -> missing_base
```

Reliability reports should expose counts of presence disagreements separately because they indicate extraction/schema instability rather than only changed labels.

### 9.3 Panels preserve identity

```go
type PanelMember struct {
    ID             string
    ProtocolDigest string
    Judge          judging.Judge
}
```

The protocol digest may be derived from the report and checked against the member declaration. Panel majority remains an application decision.

## 10. CoinVault pilot architecture

The CoinVault pilot is the principal test of whether the lightweight guarantees are sufficient.

```text
CoinVault trace
    |
    v
CoinVault adapter -----> eval.Instance
    |                         |
    |                         v
    +---- prompt version -> ResearchRun.Prepare
                                  |
                                  v
                           ClaimJudge execution
                                  |
                                  v
                           assessment.Report
                                  |
                +-----------------+-----------------+
                |                                   |
          CoinVault artifact                  ragopt metrics
```

CoinVault continues to own:

- prompts and prompt versions;
- model/profile resolution;
- evidence meaning and authorization;
- knowledge/SQL adapters;
- product-native artifacts;
- promotion policy.

Judgekit owns:

- structural and cross-object validation;
- current-content identities;
- claim/support orchestration;
- report production;
- audit and calibration calculations.

Pilot steps:

1. Capture fixed CoinVault traces and current judge outputs.
2. Build an adapter into `eval.Instance`.
3. Implement `ClaimProtocol.Version` and rendering using existing prompts.
4. Execute with a fake generator against characterization fixtures.
5. Compare claims, verdicts, evidence links, and aggregate metrics.
6. Run one live provider experiment with cache metadata visible.
7. Remove duplicate generic CoinVault types only after parity.

## 11. API evolution plan

Because no stable release exists, make one v0 cleanup after PR 2:

```text
stored digest on mutable inputs  -> computed ContentDigest methods
PromptDigests map                -> PromptVersion + rendered digests
Confidence                       -> VerdictConfidence + EntailedProbability
Judge Evaluate only              -> optional EvaluateWithOptions
Panel []Judge                    -> []PanelMember
```

Do not preserve compatibility aliases unless a real consumer has already shipped against the old API.

## 12. Testing strategy

### Unit tests

- structural validation tables;
- current-content digest changes after mutation;
- map-order-independent semantic contract/protocol identity;
- prompt version included in cache identity;
- all cache modes;
- binary calibration from explicit entailed probability;
- evidence admission;
- direct-result ranges and labels;
- union comparison.

### Integration tests

- prepared run rejects contract/protocol mismatch before generator invocation;
- report provenance reflects exact rendered prompts;
- cache refresh creates a new generation;
- repeat probe calls the fake generator twice;
- CoinVault fixture projection preserves existing aggregate values.

### Research artifact tests

Round-trip persisted reports and verify their snapshot digest once on load.

## 13. Documentation language

Use language that matches the implementation:

| Avoid | Prefer |
|---|---|
| immutable instance | ordinary Go value; identity computed at execution |
| digest proves origin | digest fingerprints the recorded snapshot |
| sealed report cannot change | report should be treated as immutable after sealing |
| protocol guarantees reproducibility | protocol records the configuration needed for reproduction |
| evidence provenance is trusted | provenance is recorded and structurally required |

## 14. Decision records

### Decision: Optimize for research attribution, not adversarial integrity

- **Context:** Judgekit is currently used by trusted internal engineers.
- **Options considered:** No validation; full immutable/verified architecture; boundary-focused research guarantees.
- **Decision:** Enforce semantic attribution at execution and analysis boundaries while retaining ordinary mutable Go values.
- **Rationale:** Most practical risk comes from accidental experiment composition, not hostile mutation.
- **Consequences:** The API remains easy to iterate, but artifacts are not suitable as cross-trust attestations.
- **Status:** proposed

### Decision: Compute mutable-input identity on demand

- **Context:** Stored digests beside mutable structs become stale.
- **Options considered:** Reject stale digests everywhere; private immutable wrappers; compute current identity when needed.
- **Decision:** Prefer on-demand identity for instances and evidence; store digests on persisted reports.
- **Rationale:** This removes stale-state classes without a typestate system.
- **Consequences:** Digest computation occurs at run boundaries; large inputs may require profiling.
- **Status:** proposed

### Decision: Version prompt behavior explicitly

- **Context:** Static template and rendered prompt identity are different concepts.
- **Options considered:** Hash renderer code; compare rendered text with static digest; application-owned version plus rendered digest.
- **Decision:** Pin an application-owned prompt version and compute rendered-content digests.
- **Rationale:** It is clear, inspectable, and adequate for controlled research.
- **Consequences:** Engineers must bump the version when prompt behavior changes.
- **Status:** proposed

## 15. Implementation phases after merge

1. Update documentation claims.
2. Introduce on-demand digest methods and run identity.
3. Add `ResearchRun.Prepare` preflight.
4. Simplify prompt identity.
5. Add cache modes.
6. Clarify confidence/probability fields.
7. Update audit and panel identity.
8. Migrate examples.
9. Run the CoinVault pilot.
10. Review whether any hardened features are justified by evidence.

## 16. Definition of done

The lightweight design is complete when:

- current-content identity is used for cache and report attribution;
- one preflight checks contract, protocol, instance, and evidence together;
- prompt version and rendered prompt digest are both recorded;
- calibration uses explicit target probability;
- audit runs can bypass caching and count missing output;
- panel members have stable identity;
- CoinVault can reproduce its current judge path through judgekit;
- documentation makes no adversarial-integrity claim;
- all tests and CI are green.

## 17. References

- `eval/instance.go`
- `eval/digest.go`
- `judging/claimjudge.go`
- `judging/cache.go`
- `protocol/protocol.go`
- `assessment/report.go`
- `audit/reliability.go`
- `audit/disagreement.go`
- `audit/panel.go`
- `calibration/report.go`
- `README.md`
- `GLOSSARY.md`
- `pkg/doc/reference/01-developer-reference.md`
- `ttmp/.../design-doc/02-pr-2-code-review-stabilization-guide.md`
