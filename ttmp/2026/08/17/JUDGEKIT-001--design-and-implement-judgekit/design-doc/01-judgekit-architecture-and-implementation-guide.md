---
Title: Judgekit Architecture and Implementation Guide
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
    - Path: repo://AGENT.md
      Note: Repository implementation constraints
    - Path: repo://GLOSSARY.md
      Note: Textbook-grounded construct definitions
    - Path: repo://README.md
      Note: Current template README requiring replacement
    - Path: repo://assessment/report.go
      Note: Sealed assessment reports
    - Path: repo://assessment/validate.go
      Note: Three-way support and cross-reference validation
    - Path: repo://audit/panel.go
      Note: Panels preserving every member report
    - Path: repo://audit/probes.go
      Note: Reliability probes with stated invariants
    - Path: repo://audit/reliability.go
      Note: Reliability aggregation and sealed report
    - Path: repo://boundary_test.go
      Note: Core dependency boundary test
    - Path: repo://calibration/confusion.go
      Note: 2x2 confusion and sensitivity/specificity/false-support-rate
    - Path: repo://calibration/gold.go
      Note: Gold records retaining reviewer identity
    - Path: repo://calibration/report.go
      Note: Calibrate entry point and sealed calibration report
    - Path: repo://calibration/scoring.go
      Note: Brier score and expected calibration error
    - Path: repo://cmd/judgekit/main.go
      Note: Thin Glazed help-host CLI
    - Path: repo://eval/instance.go
      Note: Evaluation instance types
    - Path: repo://go.mod
      Note: Current unnormalized module and dependency baseline
    - Path: repo://judging/claimjudge.go
      Note: Two-stage claim judge
    - Path: repo://judging/interfaces.go
      Note: Provider-neutral Judge/Generator interfaces
    - Path: repo://pkg/doc/tutorials/01-getting-started.md
      Note: Getting-started Glazed help entry
    - Path: repo://protocol/protocol.go
      Note: Complete protocol identity
    - Path: repo://spec/contract.go
      Note: Measurement contracts and aggregation methods
    - Path: repo://spec/validate.go
      Note: Fail-closed contract validation
    - Path: repo://suite/suite.go
      Note: Acyclic evaluator suites with concurrent execution
    - Path: ws://coinvault/internal/knowledge/judge.go
      Note: Primary extraction source for judging mechanics
    - Path: ws://coinvault/ttmp/2026/08/17/COINVAULT-045--study-self-optimization-and-exploitable-evaluator-errors/design-doc/09-structured-evaluation-and-optimization-refactor-for-coinvault.md
      Note: Upstream package-boundary design
    - Path: ws://ragopt/pkg/eval/types.go
      Note: Generic experiment boundary judgekit must not absorb
    - Path: ws://ragopt/pkg/review/review.go
      Note: Blinded human review integration evidence
ExternalSources: []
Summary: An intern-oriented design and implementation plan for building judgekit as a provider-neutral Go library for measurement contracts, evaluator protocols, structured judgments, reliability audits, and calibration.
LastUpdated: 2026-08-17T18:55:00-04:00
WhatFor: Define judgekit's scope, package boundaries, APIs, invariants, migration path, tests, and integration responsibilities before implementation begins.
WhenToUse: When bootstrapping judgekit, reviewing its public API, extracting evaluation code from CoinVault, or integrating judgekit with ragopt and other products.
---




# Judgekit Architecture and Implementation Guide

## Executive summary

Judgekit will be a provider-neutral Go library for defining what an evaluator measures, identifying how it measures it, representing the evidence and assessments it produces, and testing whether those assessments are reliable and calibrated. Its purpose is not to provide one universal LLM judge. Its purpose is to make evaluation protocols explicit, typed, versioned, reproducible, and auditable.

The project begins from a concrete need in CoinVault. CoinVault already extracts factual claims from answers, judges those claims against retrieved and SQL evidence, computes faithfulness, scores answer relevance and abstention, caches judge calls, and records claim-level reasons. Those mechanisms are useful outside CoinVault, but their current types and execution logic are embedded in the product package. At the same time, CoinVault-specific meaning—authorized evidence, tool routes, required facts, employee-assistant policy, and prompts—must not be moved into a generic library.

Judgekit therefore owns **structures and mechanics**, while applications own **construct meaning and protocols**.

Judgekit will own:

- construct and measurement-contract types;
- evaluation instances, artifacts, evidence sets, and provenance;
- evaluator protocol identities;
- claims, support labels, dimension results, and reports;
- provider-neutral judge, critic, verifier, and generator interfaces;
- strict structured-output parsing and validation helpers;
- repeat/panel orchestration and disagreement reports;
- reliability, bias-probe, and calibration calculations;
- deterministic semantic identities and strict file loaders.

Judgekit will not own:

- CoinVault prompts, rubrics, tool names, authorization, or case schemas;
- optimization campaigns, candidate mutation, or promotion policy—those belong in ragopt;
- documents, chunks, retrieval, reranking, or grounded-answer contracts—those belong in ragkit;
- Geppetto, OpenAI, Anthropic, or other provider configuration in core packages;
- a CLI in the first implementation phase;
- hidden chain-of-thought storage or interpretation;
- claims that one model score is ground truth.

The initial repository is an unnormalized Go template. Its module still reads `github.com/go-go-golems/XXX`, its README is template text, and it contains a placeholder CLI. The first implementation phase must normalize it into a library module named `github.com/go-go-golems/judgekit`, remove binary-only scaffolding, reduce dependencies, initialize documentation locally, and establish package-boundary tests.

The minimum viable product should be deliberately narrow:

1. `spec`: strict construct and measurement-contract definitions.
2. `eval`: immutable evaluation instances and evidence sets.
3. `protocol`: complete evaluator protocol identity.
4. `assessment`: claims and structured results.
5. `judging`: provider-neutral interfaces and a two-stage claim judge orchestration.
6. `calibration`: human-gold confusion matrices and extraction recall.

Reliability probes, panels, and optional provider adapters follow after the core types have been exercised by one CoinVault integration.

## 1. Current repository state

The repository exists at:

```text
/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit
```

It currently contains template project plumbing:

```text
judgekit/
  AGENT.md
  README.md                 # GO GO TEMPLATE placeholder
  go.mod                    # module github.com/go-go-golems/XXX
  go.sum
  Makefile
  lefthook.yml
  .golangci.yml
  .goreleaser.yaml
  logcopter_generate.go
  cmd/XXX/main.go           # placeholder binary
  pkg/doc.go
  pkg/logcopter.go
  ttmp/                     # initialized by this ticket
```

`GOWORK=off go test ./...` currently passes only because the placeholder packages contain no substantive behavior. This is a bootstrap baseline, not implementation evidence.

### 1.1 Why judgekit should be a library-first repository

The first consumers are Go applications and experiment systems. A CLI would force command frameworks and output concerns into the dependency graph before the domain model stabilizes. A library-first design gives us:

- small provider-neutral core packages;
- precise unit tests;
- easy use from CoinVault and future systems;
- no accidental Cobra, Glazed, Bubble Tea, or provider dependency in core;
- freedom to add a separate CLI later when concrete operational workflows exist.

The placeholder `cmd/XXX` should be removed rather than renamed during the library bootstrap. `.goreleaser.yaml` can be removed or rewritten for a source-only library release; there is no binary to publish initially.

### 1.2 Repository normalization checklist

The bootstrap implementation should:

```text
module path:  github.com/go-go-golems/XXX
          ->  github.com/go-go-golems/judgekit

log area:    go-go-golems.XXX
          -> go-go-golems.judgekit

README:      template banners
          -> project purpose, boundaries, package map, quick start

cmd/XXX:     placeholder binary
          -> remove
```

Then run:

```bash
GOWORK=off go mod tidy
GOWORK=off go test ./...
make logcopter-generate
make logcopter-check
```

Do not keep aliases for the `XXX` module path or placeholder packages.

## 2. Conceptual model

### 2.1 The inference chain

Judgekit models the middle of this chain:

```text
abstract construct
  -> measurement contract
  -> evaluation protocol
  -> evaluation instance + evidence
  -> structured assessment
  -> statistical audit/calibration
  -> application decision
```

The application owns the construct's substantive meaning and the eventual decision. Judgekit makes the intermediate steps explicit and reproducible.

### 2.2 Core terms

**Construct.** An abstract property intended to be measured, such as evidence faithfulness, completeness, or appropriate abstention.

**Measurement contract.** The operational definition of a construct: unit of judgment, allowed evidence, labels, aggregation, exclusions, and empty-case behavior.

**Evaluation instance.** One concrete item: input, candidate artifact, evidence, optional reference, required facts, and metadata.

**Protocol.** The complete reproducible instrument configuration: measurement-contract digest, model identity, prompts, decoding, evidence ordering, parser, retries, and aggregation.

**Judge.** An evaluator producing a score, label, ranking, or structured assessment used for reporting or decision support.

**Critic.** An evaluator producing diagnostic feedback intended to guide change.

**Verifier.** A narrow evaluator checking a proposition or constraint.

**Assessment.** The structured observation produced by an evaluator, including provenance and protocol identity.

**Reliability.** Stability under changes that should not affect the construct.

**Calibration.** Agreement between stated confidence and empirical frequency on a specified population.

Judgekit should use these names consistently in public APIs and documentation.

### 2.3 What judgekit does not decide

Judgekit does not decide whether a model should be deployed. It may compute measurements and calibration evidence, but promotion policy belongs to an application or ragopt.

```text
judgekit: “protocol P measured faithfulness 0.86 with these claims”
ragopt:   “candidate improved 0.04 over incumbent on paired cases”
product:  “this confidence-bound evidence is sufficient to deploy”
```

Collapsing those statements into one `Score()` method would recreate the ambiguity judgekit is intended to remove.

## 3. Package architecture

```text
judgekit/
  spec/
    construct.go
    contract.go
    digest.go
    load.go
    validate.go

  eval/
    artifact.go
    evidence.go
    instance.go
    facts.go
    digest.go
    validate.go

  protocol/
    protocol.go
    model.go
    decoding.go
    retry.go
    digest.go
    load.go

  assessment/
    claim.go
    support.go
    dimension.go
    report.go
    validate.go

  judging/
    interfaces.go
    claimjudge.go
    aggregate.go
    parsing.go
    repair.go

  suite/
    suite.go
    runner.go
    dependency.go

  audit/
    probes.go
    reliability.go
    disagreement.go
    bias.go

  calibration/
    gold.go
    confusion.go
    report.go
    slices.go

  internal/
    canonicaljson/
    strictdecode/
    identifier/

  provider/                  # optional, later
    geppetto/
```

### 3.1 Dependency direction

```text
spec        eval        protocol
  \          |          /
   \         |         /
        assessment
             |
          judging
          /      \
       suite    audit
          \      /
         calibration
```

More precisely:

- `spec` imports only the standard library and internal helpers.
- `eval` may refer to construct IDs but should not depend on judging execution.
- `protocol` refers to a measurement digest, not an application package.
- `assessment` imports `eval`, `spec`, and `protocol` value identifiers.
- `judging` imports all core value packages and exposes execution interfaces.
- `audit` and `calibration` consume reports but do not call providers directly.
- provider adapters import `judging`, never the reverse.

Add a boundary test that parses imports and rejects core dependencies on:

```text
geppetto
pinocchio
glazed
cobra
bubbletea
coinvault
ragopt
ragkit
provider SDKs
```

## 4. Core APIs

### 4.1 Constructs

```go
package spec

type ConstructID string

type Direction string

const (
    Maximize Direction = "maximize"
    Minimize Direction = "minimize"
    Descriptive Direction = "descriptive"
)

type Range struct {
    Minimum float64 `json:"minimum" yaml:"minimum"`
    Maximum float64 `json:"maximum" yaml:"maximum"`
}

type Construct struct {
    ID         ConstructID `json:"id" yaml:"id"`
    Name       string      `json:"name" yaml:"name"`
    Definition string      `json:"definition" yaml:"definition"`
    Unit       string      `json:"unit" yaml:"unit"`
    Direction  Direction   `json:"direction" yaml:"direction"`
    Range      *Range      `json:"range,omitempty" yaml:"range,omitempty"`
}
```

Validation rules:

- IDs are bounded portable identifiers.
- Names and definitions are non-empty.
- Units are explicit.
- Range values are finite and ordered.
- Direction is one of the declared constants.
- Duplicate construct IDs are rejected.

### 4.2 Measurement contracts

```go
type EvidencePolicy struct {
    AllowedKinds      []string `json:"allowed_kinds" yaml:"allowed_kinds"`
    ForbiddenKinds    []string `json:"forbidden_kinds" yaml:"forbidden_kinds"`
    RequireProvenance bool     `json:"require_provenance" yaml:"require_provenance"`
}

type Aggregation struct {
    Method      string `json:"method" yaml:"method"`
    Numerator   string `json:"numerator,omitempty" yaml:"numerator,omitempty"`
    Denominator string `json:"denominator,omitempty" yaml:"denominator,omitempty"`
    EmptyPolicy string `json:"empty_policy" yaml:"empty_policy"`
}

type MeasurementContract struct {
    APIVersion     string                       `json:"api_version" yaml:"api_version"`
    Name           string                       `json:"name" yaml:"name"`
    Constructs     []Construct                  `json:"constructs" yaml:"constructs"`
    EvidencePolicy EvidencePolicy               `json:"evidence_policy" yaml:"evidence_policy"`
    Labels         map[ConstructID][]string     `json:"labels" yaml:"labels"`
    Aggregations   map[ConstructID]Aggregation  `json:"aggregations" yaml:"aggregations"`
    Exclusions     map[ConstructID][]string     `json:"exclusions,omitempty" yaml:"exclusions,omitempty"`
}

type ContractDocument struct {
    Contract   MeasurementContract `json:"contract"`
    Digest     string              `json:"digest"`
    ByteDigest string              `json:"byte_digest"`
    Path       string              `json:"path"`
}
```

The semantic digest is computed from canonical JSON after validation. The byte digest records exact source bytes. Both matter: semantic identity ignores harmless YAML ordering, while byte identity proves which reviewed file was used.

### 4.3 Evaluation artifacts

```go
package eval

type Artifact struct {
    MediaType string `json:"media_type"`
    Text      string `json:"text,omitempty"`
    URI       string `json:"uri,omitempty"`
    Digest    string `json:"digest"`
    SizeBytes int64  `json:"size_bytes"`
}
```

An artifact may carry inline text or an immutable URI reference, but its digest is always required. Judgekit should not become an arbitrary file loader; applications resolve private artifacts before evaluation.

### 4.4 Evidence

```go
type EvidenceItem struct {
    ID         string            `json:"id"`
    Kind       string            `json:"kind"`
    Content    Artifact          `json:"content"`
    SourceID   string            `json:"source_id"`
    SourceTime *time.Time        `json:"source_time,omitempty"`
    Authority  string            `json:"authority,omitempty"`
    Provenance map[string]string `json:"provenance"`
}

type EvidenceSet struct {
    Items        []EvidenceItem `json:"items"`
    PolicyDigest string         `json:"policy_digest"`
    Digest       string         `json:"digest"`
}
```

Evidence IDs must be unique within a set. The digest covers ordered normalized items. If the protocol declares order-insensitive presentation, it may sort a copy before rendering; it must not mutate the original set.

### 4.5 Evaluation instances

```go
type RequiredFact struct {
    ID          string   `json:"id"`
    Description string   `json:"description"`
    Importance  float64  `json:"importance"`
    EvidenceIDs []string `json:"evidence_ids,omitempty"`
}

type Instance struct {
    ID            string            `json:"id"`
    Input         Artifact          `json:"input"`
    Candidate     Artifact          `json:"candidate"`
    Evidence      EvidenceSet       `json:"evidence"`
    Reference     *Artifact         `json:"reference,omitempty"`
    RequiredFacts []RequiredFact    `json:"required_facts,omitempty"`
    Metadata      map[string]string `json:"metadata,omitempty"`
    Digest        string            `json:"digest"`
}
```

Required facts are optional because some constructs do not have one canonical checklist. When present, they enable completeness measurement independent of answer length.

### 4.6 Protocol identity

```go
package protocol

type ModelIdentity struct {
    Provider string            `json:"provider" yaml:"provider"`
    Model    string            `json:"model" yaml:"model"`
    Revision string            `json:"revision,omitempty" yaml:"revision,omitempty"`
    Settings map[string]string `json:"settings,omitempty" yaml:"settings,omitempty"`
}

type DecodingPolicy struct {
    Temperature *float64 `json:"temperature,omitempty" yaml:"temperature,omitempty"`
    TopP        *float64 `json:"top_p,omitempty" yaml:"top_p,omitempty"`
    MaxTokens   int      `json:"max_tokens" yaml:"max_tokens"`
    Seed        *int64   `json:"seed,omitempty" yaml:"seed,omitempty"`
}

type RetryPolicy struct {
    MaximumAttempts int      `json:"maximum_attempts" yaml:"maximum_attempts"`
    RepairKinds     []string `json:"repair_kinds,omitempty" yaml:"repair_kinds,omitempty"`
}

type Protocol struct {
    APIVersion         string            `json:"api_version" yaml:"api_version"`
    Name               string            `json:"name" yaml:"name"`
    MeasurementDigest  string            `json:"measurement_digest" yaml:"measurement_digest"`
    Model              ModelIdentity     `json:"model" yaml:"model"`
    PromptDigests      map[string]string `json:"prompt_digests" yaml:"prompt_digests"`
    Decoding           DecodingPolicy    `json:"decoding" yaml:"decoding"`
    EvidenceOrder      string            `json:"evidence_order" yaml:"evidence_order"`
    ParserVersion      string            `json:"parser_version" yaml:"parser_version"`
    AggregatorVersion  string            `json:"aggregator_version" yaml:"aggregator_version"`
    Retry              RetryPolicy       `json:"retry" yaml:"retry"`
}
```

A model name alone is never a protocol identity.

### 4.7 Claims and verdicts

```go
package assessment

type Span struct {
    Start int `json:"start"`
    End   int `json:"end"`
}

type Claim struct {
    ID               string  `json:"id"`
    Text             string  `json:"text"`
    Importance       float64 `json:"importance"`
    RequiresEvidence bool    `json:"requires_evidence"`
    CandidateSpan    *Span   `json:"candidate_span,omitempty"`
}

type SupportLabel string

const (
    Entailed     SupportLabel = "entailed"
    Contradicted SupportLabel = "contradicted"
    Insufficient SupportLabel = "insufficient"
)

type ClaimAssessment struct {
    ClaimID     string       `json:"claim_id"`
    Label       SupportLabel `json:"label"`
    EvidenceIDs []string     `json:"evidence_ids"`
    Confidence  *float64     `json:"confidence,omitempty"`
    Reason      string       `json:"reason"`
}
```

Use three-way support labels from the beginning. A boolean cannot distinguish contradiction from absent evidence, and those failures require different interventions.

### 4.8 Dimension results and reports

```go
type DimensionResult struct {
    ConstructID spec.ConstructID `json:"construct_id"`
    Applicable  bool             `json:"applicable"`
    Value       *float64         `json:"value,omitempty"`
    Label       string           `json:"label,omitempty"`
    Confidence  *float64         `json:"confidence,omitempty"`
    EvidenceIDs []string         `json:"evidence_ids,omitempty"`
    Diagnostics []string         `json:"diagnostics,omitempty"`
}

type Report struct {
    APIVersion      string            `json:"api_version"`
    InstanceID      string            `json:"instance_id"`
    InstanceDigest  string            `json:"instance_digest"`
    ProtocolDigest  string            `json:"protocol_digest"`
    Claims          []Claim           `json:"claims,omitempty"`
    ClaimResults    []ClaimAssessment `json:"claim_results,omitempty"`
    Dimensions      []DimensionResult `json:"dimensions"`
    RawArtifacts    []eval.Artifact   `json:"raw_artifacts,omitempty"`
    StartedAt       time.Time         `json:"started_at"`
    FinishedAt      time.Time         `json:"finished_at"`
    Digest          string            `json:"digest"`
}
```

Reports are immutable values after validation and sealing. Raw provider text can be retained as a content-addressed artifact when policy allows it; it should not be required in every report.

## 5. Judging execution

### 5.1 Provider-neutral interfaces

```go
package judging

type GenerationRequest struct {
    Prompt      string
    MediaType   string
    ProtocolID  string
    Step        string
}

type GenerationResult struct {
    Text         string
    InputTokens  int
    OutputTokens int
    Model        protocol.ModelIdentity
    Duration     time.Duration
}

type Generator interface {
    Generate(context.Context, GenerationRequest) (GenerationResult, error)
}

type Judge interface {
    Evaluate(context.Context, eval.Instance) (assessment.Report, error)
}

type Critic interface {
    Critique(context.Context, eval.Instance, assessment.Report) (Critique, error)
}

type Verifier interface {
    Verify(context.Context, VerificationRequest) (VerificationResult, error)
}
```

The `Generator` does not expose provider-specific request objects. Adapters translate provider configuration into the stable interface.

### 5.2 Two-stage claim judge

The first reusable evaluator should implement the proven decomposition:

```text
input + candidate
  -> claim extractor (evidence hidden)
  -> validated claims
  -> claims + evidence
  -> support judge
  -> validated claim assessments
  -> contract-defined aggregation
  -> assessment report
```

```go
type ClaimProtocol interface {
    ExtractPrompt(eval.Instance) (string, error)
    SupportPrompt(eval.Instance, []assessment.Claim) (string, error)
}

type ClaimJudge struct {
    Contract spec.ContractDocument
    Protocol protocol.Document
    Prompts  ClaimProtocol
    Generate Generator
    Cache    Cache
}

var _ Judge = (*ClaimJudge)(nil)
```

The extractor must not see evidence unless the measurement contract explicitly defines evidence-conditioned claim extraction. The support judge may only cite IDs in the provided evidence set.

### 5.3 Structured parsing

Judgekit should provide strict parsing primitives, not hard-code one vendor's structured-output API.

```go
func DecodeJSONObjectStrict[T any](raw string) (T, error)
func StripSingleCodeFence(raw string) string
func ValidateSingleJSONValue(raw string) error
```

Decoding rules:

- optionally strip one whole-response code fence;
- require exactly one JSON object;
- reject unknown fields;
- reject missing required pointer fields;
- reject trailing values;
- return a typed structural error suitable for one repair attempt;
- never search arbitrary prose for the first `{...}` in the strict default path.

CoinVault currently tolerates prose surrounding JSON. Judgekit should default to stricter behavior. A compatibility mode is not needed for the new package; applications can deliberately configure a tolerant parser if required.

### 5.4 Repair policy

A repair is part of protocol identity.

```go
type Repairer interface {
    RepairPrompt(original string, failure StructuralError) (string, error)
}
```

Only structural failures should be repairable by default. A semantically invalid assessment—unknown evidence IDs, wrong claim ordering, out-of-range confidence—should not trigger repeated unconstrained retries unless the protocol explicitly allows it.

### 5.5 Caching

```go
type CacheKey struct {
    ProtocolDigest string
    InstanceDigest string
    Step           string
    PromptDigest   string
}

type Cache interface {
    Load(context.Context, CacheKey, any) (bool, error)
    Store(context.Context, CacheKey, any) error
}
```

Caching improves reproducibility and cost control. It does not count as reliability measurement. Audit runners need an explicit cache-bypass mode.

## 6. Evaluator suites

A suite combines multiple evaluators without collapsing their outputs prematurely.

```go
package suite

type Evaluator interface {
    Name() string
    DependsOn() []string
    Evaluate(context.Context, eval.Instance, Results) (assessment.Report, error)
}

type Suite struct {
    APIVersion string
    Name       string
    Evaluators []Evaluator
}

type Results struct {
    Reports map[string]assessment.Report
}

func (s Suite) Run(context.Context, eval.Instance) (Results, error)
```

The suite validates an acyclic dependency graph and runs independent evaluators concurrently with `errgroup`. Deterministic verifiers can run alongside LLM judges. One evaluator may consume another's claim extraction only when that dependency is declared.

Example application composition:

```text
claim_extractor
  -> support_judge
required_fact_verifier
citation_resolver
abstention_judge
style_judge
```

Each report retains its own protocol identity. A suite digest identifies the evaluator graph and versions.

## 7. Reliability and bias audits

### 7.1 Probe model

```go
package audit

type ProbeKind string

const (
    Repeat            ProbeKind = "repeat"
    EvidenceOrder     ProbeKind = "evidence_order"
    CandidateOrder    ProbeKind = "candidate_order"
    PromptParaphrase  ProbeKind = "prompt_paraphrase"
    FormatTransform   ProbeKind = "format_transform"
    CrossJudge        ProbeKind = "cross_judge"
)

type Probe struct {
    ID             string
    Kind           ProbeKind
    BaseInstance   eval.Instance
    VariantInstance eval.Instance
    Invariants     []string
}
```

A probe must state what remains semantically invariant. Removing headings, reversing pair order, or shuffling evidence should not silently change the intended construct.

### 7.2 Reliability report

```go
type ReliabilityReport struct {
    APIVersion          string
    ProtocolDigest      string
    ProbeSetDigest      string
    TotalPairs          int
    ClaimLabelAgreement float64
    DimensionAgreement  map[spec.ConstructID]float64
    MeanAbsoluteDelta   map[spec.ConstructID]float64
    Disagreements       []Disagreement
    Digest              string
}
```

Do not report one “reliability score” without per-construct and per-probe breakdowns.

### 7.3 Panels and correlated errors

```go
type Panel struct {
    Judges []Judge
    Policy AggregationPolicy
}
```

Panel output must preserve every member report and disagreement. Majority vote is an aggregation, not independent truth. Judgekit can compute pairwise agreement but cannot infer error independence without external labels.

## 8. Calibration

### 8.1 Gold records

```go
package calibration

type GoldClaim struct {
    InstanceID  string
    Claim       assessment.Claim
    Label       assessment.SupportLabel
    ReviewerIDs []string
    Adjudicated bool
}

type GoldDimension struct {
    InstanceID  string
    ConstructID spec.ConstructID
    Value       *float64
    Label       string
    ReviewerIDs []string
}
```

Human labels must retain reviewer identity or pseudonymous stable IDs so agreement can be measured. Adjudication should not erase original disagreement.

### 8.2 Extraction recall

A claim judge can appear accurate by failing to extract difficult claims. Calibration must compare model-extracted claims with human-enumerated claims.

```text
extraction recall
  = matched model-extracted factual claims
    / human-enumerated factual claims
```

Matching can begin with human linkage in calibration data. Automated semantic matching may be added later but must itself be validated.

### 8.3 Confusion and probability calibration

```go
type Confusion struct {
    EntailedAsEntailed         int
    EntailedAsNonEntailed      int
    NonEntailedAsEntailed      int
    NonEntailedAsNonEntailed   int
}

type Report struct {
    ProtocolDigest    string
    DatasetDigest     string
    ExtractionRecall float64
    Sensitivity       float64
    Specificity       float64
    FalseSupportRate  float64
    BrierScore        *float64
    ECE               *float64
    ByGroup           map[string]SliceReport
}
```

Calibration metrics apply only when the protocol emits confidence probabilities. A 1–5 ordinal score must not be treated as a probability without an explicit calibrated mapping.

## 9. Integration boundaries

### 9.1 CoinVault

CoinVault retains:

- prompts and rubrics;
- allowed evidence kinds;
- knowledge and SQL evidence adapters;
- required facts and case strata;
- authorization and tool contracts;
- judge model/profile resolution;
- native artifact integration.

CoinVault replaces local generic types with judgekit types and maps judgekit reports into ragopt metrics.

```go
instance := evaluationadapter.FromCoinVaultTrace(caseInput, trace)
report, err := claimJudge.Evaluate(ctx, instance)
product := projectCoinVaultEvaluation(report, contracts)
outcome := projectRagoptOutcome(product)
```

### 9.2 Ragopt

Ragopt should not call judge providers. It receives:

- protocol and measurement digests in run identity;
- metric projections in outcomes;
- full judgekit reports through native artifacts;
- optional calibration-report digests as gate prerequisites.

Future ragopt policy may require:

```yaml
required_evaluator_evidence:
  protocol_digest: sha256:...
  calibration_report_digest: sha256:...
  maximum_false_support_rate: 0.05
```

### 9.3 Ragkit

Ragkit remains independent. Applications adapt `rag.Evidence` into judgekit evidence. Judgekit should not depend on RAG-specific fields because it must also evaluate code, support answers, plans, and tool trajectories.

### 9.4 Provider adapters

Provider adapters are optional and explicit:

```text
judgekit/provider/geppetto
judgekit/provider/openai       # only if direct adapter is justified
```

A provider adapter implements `judging.Generator`; it does not define constructs, prompts, or product defaults.

## 10. Implementation phases

### Phase 0: normalize the repository

1. Rename module to `github.com/go-go-golems/judgekit`.
2. Replace logcopter area placeholders.
3. Remove `cmd/XXX`.
4. Rewrite README.
5. Reduce `go.mod` with `go mod tidy`.
6. Confirm local docmgr root points to `judgekit/ttmp`.
7. Add a boundary test.
8. Commit bootstrap separately from domain code.

### Phase 1: internal canonical primitives

Implement:

- bounded identifier validation;
- strict JSON/YAML decode;
- canonical semantic JSON;
- SHA-256 identity helpers;
- finite numeric validation;
- immutable cloning helpers.

These primitives should remain internal until repeated public use justifies exposure.

### Phase 2: `spec`

Implement constructs and measurement contracts:

- types;
- strict loader;
- validation;
- semantic and byte digests;
- fixture examples;
- malformed-input tests.

Definition of done: a CoinVault-style faithfulness contract loads and has stable identity.

### Phase 3: `eval`

Implement artifacts, evidence, required facts, instances, validation, and digests.

Definition of done: knowledge and SQL evidence can coexist without losing type or provenance.

### Phase 4: `protocol`

Implement model, decoding, retry, parser, and aggregation identity.

Definition of done: changing any semantically relevant field changes the protocol digest.

### Phase 5: `assessment`

Implement claims, spans, three-way support, dimensions, reports, cross-reference validation, and sealing.

Definition of done: unknown evidence, duplicate claims, invalid spans, and non-finite values fail closed.

### Phase 6: `judging`

Implement provider-neutral interfaces, strict parsing, repair abstraction, cache interface, and two-stage claim judge.

Use a deterministic fake generator for all unit tests. Provider integration is not required for this phase.

### Phase 7: CoinVault pilot

Port only the claim extraction and support-verdict path:

1. CoinVault renders existing prompts.
2. CoinVault adapts its generator.
3. Judgekit orchestrates and validates.
4. CoinVault stores the report and projects existing metrics.
5. Compare characterization fixtures.
6. Delete replaced generic local types and helpers.

Do not refactor CoinVault treatments in the same change.

### Phase 8: calibration

Implement gold records, extraction recall, confusion metrics, slices, Brier score, and ECE. Integrate with ragopt blinded review output through an application adapter.

### Phase 9: audit and suite

Implement evaluator suites, repeated fresh-call probes, order transformations, disagreement reports, and panel preservation.

### Phase 10: optional provider adapters

Add adapters only after two applications use the core interface. Keep credentials, profiles, and product defaults outside judgekit.

## 11. Test strategy

### 11.1 Table-driven validation tests

Every public value type needs:

- valid minimal fixture;
- valid complete fixture;
- missing required field;
- unknown field;
- duplicate identity;
- whitespace and normalization edge cases;
- NaN/Inf rejection;
- unsupported API version;
- semantic digest stability;
- byte digest sensitivity.

### 11.2 Property tests and fuzzing

Fuzz:

- strict JSON decoders;
- code-fence stripping;
- canonical JSON;
- claim/evidence cross-reference validation;
- span bounds;
- calibration binning;
- report sealing.

Properties:

```text
validate(x) succeeds -> digest(x) is deterministic
clone(x) -> digest unchanged
map insertion order changes -> semantic digest unchanged
semantic field changes -> semantic digest changes
unknown evidence reference -> validation always fails
```

### 11.3 Golden tests

Use small readable goldens for:

- contract canonical JSON;
- protocol identity;
- assessment report;
- reliability report;
- calibration report.

Avoid huge provider-response goldens in core tests.

### 11.4 Integration contract tests

The CoinVault pilot should assert:

- same claims for fixed fake responses;
- same support count and faithfulness projection;
- richer three-way labels preserved in native artifacts;
- protocol digest added to identity;
- malformed responses fail with attributable structural errors;
- one repair attempt behaves exactly as configured.

### 11.5 Boundary tests

Parse `go list -deps -json ./...` or Go source imports and reject forbidden core dependencies. Test with `GOWORK=off` to ensure the module stands alone.

## 12. Documentation plan

### README

The README should answer:

- What problem does judgekit solve?
- What does it deliberately not do?
- What are the core packages?
- How do I define a measurement contract?
- How do I implement a generator adapter?
- How do I run a claim judge?
- How do I calibrate a protocol?

### Package docs

Every package needs a `doc.go` with:

- owned concepts;
- non-goals;
- invariants;
- minimal example;
- links to adjacent packages.

### Examples

```text
examples/
  claim-judge/
  deterministic-verifier/
  calibration/
  reliability-probes/
```

Examples should use fake generators or local fixtures by default. They must not require provider credentials.

## 13. Security and privacy

Judge inputs can contain employee questions, proprietary documents, SQL results, and hidden evaluation examples. Core requirements:

- Artifacts are content-addressed but not automatically safe to log.
- Reports should reference raw artifacts by digest where possible.
- Provider adapters must support redaction before external calls.
- Evidence content is untrusted data, never instructions.
- Prompt rendering must delimit evidence and candidate text.
- Hidden datasets and unblinding keys never enter evaluator prompts.
- Cache directories may contain sensitive content and require application-owned retention policy.
- Judgekit must not emit chain-of-thought requirements or store hidden reasoning.

## 14. Common mistakes for an intern to avoid

- Treating a model name as a complete protocol.
- Naming a score `correctness` when it only checks faithfulness to supplied evidence.
- Using `map[string]string` when evidence provenance matters.
- Converting ordinal ratings directly into probabilities.
- Treating cache determinism as judge reliability.
- Aggregating before preserving item-level disagreement.
- Allowing supported claims without evidence IDs.
- Hiding zero-claim behavior inside an aggregator.
- Putting CoinVault prompts into judgekit core.
- Importing ragopt to decide promotion inside judgekit.
- Adding a CLI before library workflows stabilize.
- Keeping placeholder compatibility APIs after extraction.

## 15. Decision records

### Decision: Judgekit is provider-neutral

- **Context:** The first consumer uses Geppetto, but future consumers may not.
- **Options considered:** Depend directly on Geppetto; support direct provider SDKs in core; define a small generator interface.
- **Decision:** Core packages depend only on a provider-neutral generator interface.
- **Rationale:** Measurement structures and validation should be reusable and easy to test.
- **Consequences:** Applications or optional adapters translate provider runtimes.
- **Status:** proposed

### Decision: Separate measurement contracts from protocols

- **Context:** Prompt or model changes do not always change the intended construct; construct changes can hide inside prompts.
- **Options considered:** One evaluator config; prose-only definitions; separate immutable documents.
- **Decision:** Measurement and protocol are separate documents with separate digests.
- **Rationale:** This distinguishes what should be measured from how one instrument attempts to measure it.
- **Consequences:** Reports carry both identities indirectly through the protocol.
- **Status:** proposed

### Decision: Use three-way claim support

- **Context:** Boolean support conflates contradiction and absent evidence.
- **Options considered:** Boolean; three-way; open-ended taxonomy.
- **Decision:** Core support labels are entailed, contradicted, and insufficient.
- **Rationale:** The distinction is broadly useful and maps to different failure interventions.
- **Consequences:** Applications may project to booleans but native reports retain three-way labels.
- **Status:** proposed

### Decision: Library-first, no initial CLI

- **Context:** The template includes a placeholder binary.
- **Options considered:** Build a judgekit CLI immediately; retain placeholder command; focus on library APIs.
- **Decision:** Remove the placeholder CLI and implement the library first.
- **Rationale:** No stable command workflow exists yet, and CLI dependencies would contaminate the core.
- **Consequences:** Operational commands remain in consuming applications until reusable workflows emerge.
- **Status:** proposed

### Decision: No optimization or deployment decisions in judgekit

- **Context:** Judge reports will feed ragopt and product gates.
- **Options considered:** Add thresholds and promotion helpers; expose only measurements and audits.
- **Decision:** Judgekit does not approve candidates or deployments.
- **Rationale:** Evaluation evidence and decision policy are separate responsibilities.
- **Consequences:** Ragopt and products must interpret reports explicitly.
- **Status:** accepted

### Decision: Incubate through a CoinVault pilot

- **Context:** A new generic API can become speculative without a real consumer.
- **Options considered:** Design the complete library before integration; extract everything at once; implement core slices and pilot them incrementally.
- **Decision:** Implement core value packages, then pilot the two-stage CoinVault judge before expanding.
- **Rationale:** The pilot supplies concrete pressure while characterization tests protect behavior.
- **Consequences:** Later APIs may evolve based on integration evidence; no stability promise is made before the pilot.
- **Status:** proposed

## 16. Intern implementation checklist

Before opening the first code PR:

- Read this guide and the CoinVault structured-refactor guide.
- Inspect `coinvault/internal/knowledge/judge.go` end to end.
- Run standalone judgekit tests with `GOWORK=off`.
- Confirm the repository branch and template state.
- Separate repository normalization from domain implementation.

For every implementation PR:

- State which package owns the concept.
- Add strict validation tests.
- Add semantic identity tests.
- Run `GOWORK=off go test ./...`.
- Run boundary tests.
- Update package docs.
- Avoid unrelated CoinVault changes.
- Remove replaced code rather than adding permanent adapters.

## 17. Definition of done

Judgekit v0 is ready for an initial tagged release when:

- the repository is normalized;
- core packages have no forbidden dependencies;
- measurement and protocol documents load strictly and have stable identities;
- instances and evidence preserve provenance;
- reports validate claim/evidence cross-references;
- the two-stage claim judge works with a fake generator;
- calibration computes extraction recall and claim confusion metrics;
- one CoinVault pilot uses judgekit without duplicate local generic structures;
- README, package docs, and examples are complete;
- `GOWORK=off go test ./...`, lint, logcopter checks, and security scans pass;
- no CLI or provider credential is required for tests;
- public APIs are reviewed before the first stability commitment.

## Conclusion

Judgekit should make machine evaluation explicit without pretending to make it infallible. Its value is structural: a reader can see the construct, operational definition, evidence, protocol, assessment, uncertainty, and calibration record as separate artifacts. An application can change its prompts without silently changing its target, or change its target while making that semantic shift visible.

The project should begin small. Normalize the repository, establish dependency boundaries, implement strict value types, and prove them against CoinVault's existing decomposed judge. Reliability suites, panels, and provider adapters should follow observed needs. Optimization and promotion remain outside the package.

That boundary is the central design principle: judgekit produces auditable measurement evidence; it does not grant that evidence authority it has not earned.

## References

- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/go.mod`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/README.md`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/AGENT.md`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/coinvault/internal/knowledge/judge.go`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/coinvault/cmd/coinvault/cmds/knowledge_ragopt.go`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/coinvault/cmd/coinvault/cmds/knowledge_ragopt_trace.go`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/ragopt/pkg/eval/types.go`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/ragopt/pkg/review/review.go`
- `/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/coinvault/ttmp/2026/08/17/COINVAULT-045--study-self-optimization-and-exploitable-evaluator-errors/design-doc/09-structured-evaluation-and-optimization-refactor-for-coinvault.md`
