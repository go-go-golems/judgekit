---
Title: "judgekit Developer Reference"
Slug: developer-reference
Short: "Public API, invariants, and integration boundaries for developers building on judgekit."
Topics:
- evaluation
- llm
- reference
- development
Commands: []
Flags: []
IsTopLevel: false
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This reference is for developers building on judgekit: implementing a generator
adapter, writing a contract and protocol, wiring a custom judge, or integrating
judgekit reports into a product gate. It assumes you have read the user guide
and the getting-started tutorial.

## Public surface

Only the top-level packages (`spec`, `eval`, `protocol`, `assessment`,
`judging`, `audit`, `calibration`, `suite`) are public. Everything under
`internal/` (`canonicaljson`, `strictdecode`, `identifier`) is not importable
outside the module; if you need that behavior, it will be promoted
deliberately.

## spec — what an evaluator measures

```go
type Construct struct {
    ID         ConstructID
    Name       string
    Definition string
    Unit       string
    Direction  Direction   // maximize | minimize | descriptive
    Range      *Range      // optional, inclusive bounds
}

type MeasurementContract struct {
    APIVersion     string                      // must be spec.ContractAPIVersion
    Name           string
    Constructs     []Construct
    EvidencePolicy EvidencePolicy
    Labels         map[ConstructID][]string
    Aggregations   map[ConstructID]Aggregation
    Exclusions     map[ConstructID][]string
}

type ContractDocument struct {
    Contract   MeasurementContract
    Digest     string  // semantic, order-independent
    ByteDigest string  // raw bytes
    Path       string
}
```

Entry points: `spec.LoadContract(path)`, `spec.LoadContractFromBytes(path, raw)`,
`spec.ValidateContract(&c)`, `spec.SemanticDigest(&c)`, `spec.ByteDigest(raw)`.

Invariants:
- `APIVersion` must equal `spec.ContractAPIVersion`; unknown versions fail
  closed.
- Every construct has an aggregation. Fraction aggregations reference declared
  labels via comma-separated `numerator`/`denominator`.
- Evidence-policy allowed and forbidden kinds must not overlap.

## eval — what an evaluator observes

```go
type Artifact struct {
    MediaType string
    Text      string  // exactly one of Text/URI
    URI       string
    Digest    string  // "sha256:..."
    SizeBytes int64
}

type EvidenceSet struct {
    Items        []EvidenceItem
    PolicyDigest string
    Digest       string
}

type Instance struct {
    ID            string
    Input         Artifact
    Candidate     Artifact
    Evidence      EvidenceSet
    Reference     *Artifact
    RequiredFacts []RequiredFact
    Metadata      map[string]string
    Digest        string
}
```

Entry points: `eval.NewTextArtifact(mediaType, text)`,
`eval.NewEvidenceSet(items, policyDigest)`, `eval.NewInstance(...)`,
`eval.ValidateInstance(&inst)`.

Invariants:
- Artifacts are content-addressed; `ValidateArtifact` rejects stale text
  digests.
- `RequiredFact.EvidenceIDs` must resolve to evidence items in the instance.
- Instance and evidence-set digests are non-circular functions of content.

## protocol — how an evaluator measures

```go
type Protocol struct {
    APIVersion        string  // must be protocol.ProtocolAPIVersion
    Name              string  // bounded identifier
    MeasurementDigest string  // pins the contract
    Model             ModelIdentity
    PromptDigests     map[string]string  // step -> template/renderer digest
    Decoding          DecodingPolicy
    EvidenceOrder     string  // as_given | sorted | shuffled
    ParserVersion     string
    AggregatorVersion string
    Retry             RetryPolicy
}
```

Entry points: `protocol.LoadProtocol(path)`, `protocol.Validate(&p)`,
`protocol.SemanticDigest(&p)`.

Invariants:
- `MeasurementDigest` must be a `sha256:` digest (the contract's semantic
  digest).
- `EvidenceOrderShuffled` requires a decoding seed (reproducible
  randomization).
- The digest changes on any semantically relevant field; it is stable across
  prompt-map insertion order.

## assessment — what an evaluator produces

```go
type SupportLabel string  // entailed | contradicted | insufficient

type ClaimAssessment struct {
    ClaimID     string
    Label       SupportLabel
    EvidenceIDs []string
    VerdictConfidence   *float64  // confidence in the emitted label
    EntailedProbability *float64  // P(entailed), used by binary calibration
    Reason              string
}

type DimensionResult struct {
    ConstructID spec.ConstructID
    Applicable  bool
    Value       *float64
    Label       string
    Confidence  *float64
    EvidenceIDs []string
    Diagnostics []string
}

type Report struct {
    APIVersion     string
    InstanceID     string
    InstanceDigest string
    ProtocolDigest string
    Claims         []Claim
    ClaimResults   []ClaimAssessment
    Dimensions     []DimensionResult
    RawArtifacts   []eval.Artifact
    Provenance     RunProvenance  // contract, prompts, model, cache mode
    StartedAt      time.Time
    FinishedAt     time.Time
    Digest         string
}
```

Entry points: `assessment.Seal(&report, allowedEvidence)`,
`assessment.ValidateReport(&report, allowedEvidence)`,
`assessment.EvidenceIDSet(items)`.

Invariants:
- `Seal` validates structure (without requiring a digest), computes the digest,
  and sets it. `ValidateReport` then checks the sealed report.
- Every claim has exactly one claim result; claim results reference real claims.
- Evidence references are cross-checked against `allowedEvidence` (built from
  the instance's evidence by the judging layer).

## judging — running evaluators

```go
type Generator interface {
    Generate(ctx context.Context, req GenerationRequest) (GenerationResult, error)
}

type Judge interface {
    Evaluate(ctx context.Context, inst eval.Instance) (assessment.Report, error)
}

type ClaimProtocol interface {
    TemplateDigest(step string) (string, error)
    ExtractPrompt(input ClaimExtractionInput) (string, error)
    SupportPrompt(inst eval.Instance, claims []assessment.Claim) (string, error)
}

type ClaimJudge struct {
    Contract spec.ContractDocument
    Protocol protocol.Document
    Prompts  ClaimProtocol
    Generate Generator
    Cache    Cache          // optional; nil -> NoopCache
    Repairer Repairer       // optional; nil -> DefaultRepairer
    Clock    func() time.Time  // optional; for deterministic tests
}
```

Entry points: `(*ClaimJudge).Evaluate(ctx, inst)`,
`(*ClaimJudge).EvaluateWithOptions(ctx, inst, opts)`,
`judging.NewMemoryCache()`, `judging.FakeGenerator`,
`judging.DecodeJSONObjectStrict[T](raw)`, `judging.IsStructural(err)`.

Invariants:
- The extractor receives a restricted value with no evidence, reference, or
  required facts; the support judge may only cite evidence in the instance.
- Instance and evidence-set identities are recomputed from current content at
  the execution boundary instead of trusting a stale caller digest.
- Every generation must report the exact model identity pinned by the protocol.
- Reports retain template and rendered prompt digests, model attribution, cache
  mode, cache hits, usage, and duration.
- Only structural failures are repaired (up to `Retry.MaximumAttempts`); semantic
  failures fail closed at seal.
- Protocol prompt digests bind stable template identities; cache keys hash the fully rendered prompt.
- Reliability audits call configurable judges with `CacheBypass`.

## Implementing a generator adapter

Adapters live in applications or optional `provider/` packages, never in core.

```go
type myGen struct{ client *http.Client }

func (g *myGen) Generate(ctx context.Context, req judging.GenerationRequest) (judging.GenerationResult, error) {
    start := time.Now()
    raw, model, in, out, err := g.call(ctx, req.Prompt) // your model call
    if err != nil {
        return judging.GenerationResult{}, err
    }
    return judging.GenerationResult{Text: raw, Model: model, InputTokens: in, OutputTokens: out, Duration: time.Since(start)}, nil
}
```

`req.Step` ("extract", "support") lets you route by stage or observe usage. The
prompt is fully rendered by your `ClaimProtocol`.

## audit — reliability and bias

```go
type Probe struct {
    ID              string
    Kind            ProbeKind   // repeat | evidence_order | candidate_order | ...
    BaseInstance    eval.Instance
    VariantInstance eval.Instance
    Invariants      []string    // required: what must not change
}

type ReliabilityReport struct {
    TotalPairs          int
    ClaimLabelAgreement float64
    DimensionAgreement  map[spec.ConstructID]float64
    MeanAbsoluteDelta   map[spec.ConstructID]float64
    Disagreements       []Disagreement
    Digest              string
}

type PanelMember struct {
    ID    string
    Judge judging.Judge
}

type Panel struct {
    Members []PanelMember
    Policy  AggregationPolicy
}
```

Entry points: `audit.NewProbeSet`, `audit.RunProbe`, `audit.Reliability`,
`audit.CompareReports`, `(*Panel).Evaluate`, `audit.ValidateReport`.

Invariants: a probe must state its invariants; reliability reports carry
per-construct breakdowns; panels preserve every member report. `audit` runs a
`judging.Judge` (provider-neutral) over base/variant instances; it never calls
providers directly.

## calibration — agreement with labels

```go
type GoldSet struct {
    Claims []GoldClaim    // retain ReviewerIDs; Adjudicated does not erase them
    Dimensions []GoldDimension
    Digest string
}

type Report struct {
    ExtractionRecall float64
    Sensitivity      float64
    Specificity       float64
    FalseSupportRate float64
    BrierScore       *float64  // nil when no entailed probability emitted
    ECE              *float64
    Digest           string
}
```

Entry points: `calibration.Calibrate`, `calibration.ConfusionFromClaims`,
`calibration.ExtractionRecall`, `calibration.BrierScore`,
`calibration.ExpectedCalibrationError`, `calibration.ValidateReport`.

Invariants: gold records retain reviewer identity; extraction recall is over
all gold claims while confusion is over matched claims; Brier/ECE apply only
when an explicit entailed probability is present. Reports from another protocol are rejected. `calibration` consumes reports and gold records; it
does not run judges.

## suite — combining evaluators

```go
type Evaluator interface {
    Name() string
    DependsOn() []string
    Evaluate(ctx context.Context, inst eval.Instance, results Results) (assessment.Report, error)
}

type Suite struct {
    APIVersion string
    Name       string
    Evaluators []Evaluator
    Digest     string
}
```

Entry points: `suite.NewSuite`, `(*Suite).Validate`, `(*Suite).Run`,
`suite.JudgeEvaluator` (adapts a `judging.Judge` to a dependency-free
`Evaluator`).

Invariants: the dependency graph must be acyclic; independent evaluators run
concurrently via `errgroup`; each report retains its own protocol identity; the
suite digest is a function of the graph structure, not declaration order.

## Integration boundaries

- **Your application** owns prompts, rubrics, evidence adapters, authorization,
  and judge model resolution. Map your evidence into `eval.EvidenceItem` and
  your prompts into `ClaimProtocol`.
- **ragopt** should not call judge providers. It receives protocol and
  measurement digests, metric projections, and full reports through native
  artifacts.
- **ragkit** stays independent. Adapt `rag.Evidence` into judgekit evidence;
  judgekit must not depend on RAG-specific fields.
- **Provider adapters** implement `judging.Generator` and define nothing else.

## Testing

Tests use only fake generators and local fixtures. No provider credentials are
required. The boundary test at the module root (`boundary_test.go`) rejects
forbidden imports in core packages.

```bash
GOWORK=off go test ./...   # includes the boundary test
GOWORK=off go build ./...
make lint
```

## Common mistakes

- Treating a model name as a complete protocol.
- Naming a score `correctness` when it only checks faithfulness to supplied
  evidence.
- Converting an ordinal rating directly into a probability.
- Allowing supported claims without evidence ids.
- Putting product prompts into judgekit core.
- Importing ragopt to decide promotion inside judgekit.
- Keeping placeholder compatibility APIs after extraction.

## See Also

- `user-guide` for the conceptual model and invariants.
- `getting-started` for a runnable tutorial.
- `GLOSSARY.md` for measurement-theory definitions.
- The design doc in `ttmp/.../JUDGEKIT-001.../design-doc/` for decision records
  and the full implementation plan.
