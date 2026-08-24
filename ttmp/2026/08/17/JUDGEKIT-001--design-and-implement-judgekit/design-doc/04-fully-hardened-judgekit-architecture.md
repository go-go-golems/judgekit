---
Title: Fully Hardened Judgekit Architecture
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
    - Path: repo://assessment/digest.go
      Note: Current report sealing foundation
    - Path: repo://eval/digest.go
      Note: Instance and evidence identity foundation
    - Path: repo://internal/canonicaljson/canonicaljson.go
      Note: Current canonical identity primitive
    - Path: repo://judging/cache.go
      Note: Future content-addressed execution cache
    - Path: repo://judging/claimjudge.go
      Note: Future compiled-instrument execution boundary
    - Path: repo://protocol/load.go
      Note: Protocol loading boundary
    - Path: repo://spec/load.go
      Note: Contract loading boundary
ExternalSources: []
Summary: Reference architecture for a production-grade, cross-trust judgekit with immutable verified values, compiled protocols, attestations, and controlled artifact custody.
LastUpdated: 2026-08-24T19:25:00-04:00
WhatFor: Document the maximum-integrity design so future maintainers can evaluate it without accidentally implementing pieces incoherently.
WhenToUse: Only if judgekit must support regulated decisions, untrusted plugins, cross-organization artifact exchange, or durable audit evidence.
---


# Fully Hardened Judgekit Architecture

## 1. Executive summary

This document describes a possible fully hardened judgekit. It is a reference design, not the recommended next implementation. The current project is internal research software, and the lighter-guarantee design is more proportionate. Hardening becomes justified only when reports cross trust boundaries, drive high-impact automated decisions, must remain verifiable for years, or are produced by independently operated components.

A hardened judgekit would treat an evaluation report as an attestation over a closed set of verified inputs and executable protocol components. Values would move through explicit lifecycle states, digests would be recomputed and checked recursively, runtime prompts and provider responses would be content-addressed, and reports could be signed by an execution service. Application policy would still decide what the measurements mean; hardening would prove identity and custody, not truth.

The central flow would become:

```text
untrusted bytes
    |
    v
strict decode + schema validation
    |
    v
verified content-addressed documents
    |
    v
compiled measurement instrument
    |
    v
isolated execution + complete run manifest
    |
    v
sealed and signed report bundle
    |
    v
independent verification + calibration + decision
```

## 2. When hardening is warranted

Consider this architecture only if one or more conditions are true:

- reports approve financial, medical, legal, safety, or access-control actions;
- separate organizations exchange and rely on reports;
- evaluator plugins or adapters are not fully trusted;
- reports must survive long-term archival and later independent verification;
- hidden promotion suites require strict access and query budgets;
- regulatory or contractual audit requires custody evidence;
- a compromised application process must not rewrite evaluation history;
- multiple execution services need interoperable attestations.

Do not implement hardening because cryptographic digests look rigorous. Every guarantee must answer a concrete threat.

## 3. Threat model

### 3.1 Protected properties

The hardened design protects:

- **integrity:** recorded content was not changed after identity was established;
- **binding:** protocol, contract, evidence, prompt, model, and output belong to one run;
- **provenance:** sources and transformations are recorded;
- **replayability:** enough deterministic configuration is retained to attempt reproduction;
- **isolation:** untrusted evidence cannot modify evaluator instructions;
- **custody:** an execution service can attest that it produced the result;
- **population validity:** calibration/audit reports consume homogeneous verified populations.

### 3.2 Explicit limitations

Hardening does not prove:

- the judge is correct;
- evidence is factually true;
- a provider served the claimed model unless it supplies a verifiable attestation;
- prompt rationale reflects hidden model reasoning;
- calibration transfers to a new population;
- a majority panel is independent;
- an optimizer cannot exploit residual evaluator errors.

## 4. Lifecycle and type states

The hardened design separates untrusted transport objects from verified domain values.

```text
RawContractDTO ----parse----> ValidContract ----verify digest----> VerifiedContract
RawProtocolDTO ----parse----> ValidProtocol ----bind-----------> BoundProtocol
RawInstanceDTO ----parse----> ValidInstance ----verify---------> VerifiedInstance

VerifiedContract + BoundProtocol + Verified Renderer + Provider Adapter
                                  |
                                  v
                         CompiledInstrument
                                  |
                                  v
                            ExecutionBundle
                                  |
                                  v
                              SignedReport
```

### 4.1 Untrusted DTOs

DTOs are serialization structures only:

```go
type ContractDTO struct {
    APIVersion string          `json:"api_version"`
    Body       json.RawMessage `json:"body"`
    Digest     string          `json:"digest"`
    Signature  *SignatureDTO   `json:"signature,omitempty"`
}
```

They are never accepted by judging functions.

### 4.2 Verified values

Verified values expose read-only accessors and keep canonical bytes privately:

```go
type VerifiedContract struct {
    contract   MeasurementContract
    canonical  []byte
    digest     Digest
    source     SourceDescriptor
}

func (c VerifiedContract) Digest() Digest
func (c VerifiedContract) Contract() MeasurementContract // defensive copy
```

Constructors clone slices/maps so later caller mutation cannot alter verified content.

### 4.3 Binding values

A bound protocol proves compatibility with one verified contract and prompt renderer:

```go
type BoundProtocol struct {
    protocol      VerifiedProtocol
    contract      VerifiedContract
    renderer      VerifiedRenderer
    parser        VerifiedParser
    aggregator    VerifiedAggregator
    bindingDigest Digest
}
```

## 5. Canonical identity architecture

### 5.1 Typed digests

Avoid unstructured strings:

```go
type Algorithm string
const SHA256 Algorithm = "sha256"

type Digest struct {
    Algorithm Algorithm
    Bytes     [32]byte
}
```

Parsing rejects truncated values such as `sha256:1`. Formatting is canonical.

### 5.2 Domain-separated hashes

The same canonical bytes used in different domains should not accidentally share identity:

```text
H("judgekit.contract.v1\x00" || canonical_contract)
H("judgekit.protocol.v1\x00" || canonical_protocol)
H("judgekit.instance.v1\x00" || canonical_instance)
H("judgekit.report.v1\x00" || canonical_report)
```

Domain separation makes cross-type substitution detectable.

### 5.3 Merkle-linked execution manifest

A run manifest links every component:

```go
type ExecutionManifest struct {
    ContractDigest       Digest
    ProtocolDigest       Digest
    BindingDigest        Digest
    InstanceDigest       Digest
    EvidenceDigest       Digest
    PromptDigests        map[string]Digest
    ProviderRequestIDs   map[string]string
    ProviderResponseDigests map[string]Digest
    ParserDigest         Digest
    AggregatorDigest     Digest
    RuntimeDigest        Digest
    ParentManifest       *Digest
}
```

Manifest digest becomes the root identity carried by the report.

```text
                      Manifest Root
                    /       |       \
             protocol   instance   execution
              /   \        |       /   |   \
        contract renderer evidence request response runtime
```

### 5.4 Canonicalization registry

Every API version must identify a canonicalization algorithm. Changing number normalization, omitted-field behavior, or map ordering creates a new canonicalization version. Historical verifiers retain old algorithms.

## 6. Compiled measurement instrument

A hardened evaluator cannot be assembled from unrelated public fields. Construction produces a compiled instrument:

```go
type CompileRequest struct {
    Contract   VerifiedContract
    Protocol   VerifiedProtocol
    Renderer   VerifiedRenderer
    Parser     VerifiedParser
    Aggregator VerifiedAggregator
    Provider   ProviderExecutor
}

func Compile(req CompileRequest) (CompiledInstrument, error)
```

Compilation verifies:

- protocol measurement digest equals contract digest;
- every prompt step has a renderer/template identity;
- parser and aggregator versions match protocol declarations;
- evidence-order policy is supported;
- retry behavior is bounded;
- constructs have executable validators and aggregators;
- provider identity can satisfy the protocol model constraints;
- forbidden combinations are rejected before execution.

The compiled instrument has one identity:

```go
func (i CompiledInstrument) Digest() Digest
```

## 7. Evidence trust and provenance

### 7.1 Evidence source descriptors

```go
type EvidenceSource struct {
    SourceID       string
    Authority      string
    RetrievalTime  time.Time
    EffectiveTime  *time.Time
    ContentDigest  Digest
    TransportURI   string
    TransformationChain []Transformation
    Signature      *Signature
}
```

### 7.2 Transformation chain

Each transformation is content-addressed:

```text
source document
  -> decoder version
  -> text normalization version
  -> chunker version/config
  -> selected span
  -> evidence item
```

A transformed item records parent digest and transformation digest. This enables later reconstruction of how quoted evidence was produced.

### 7.3 Evidence admission service

A hardened policy may require independently verified evidence admission:

```go
type AdmissionDecision struct {
    PolicyDigest Digest
    EvidenceDigest Digest
    Admitted bool
    Reasons []string
    Signature Signature
}
```

The judge consumes only admitted evidence sets.

### 7.4 Prompt-injection isolation

Evidence and candidate content are untrusted byte strings. Renderer interfaces use typed channels rather than string concatenation:

```go
type PromptEnvelope struct {
    SystemInstructions TrustedText
    Rubric              TrustedText
    Candidate           UntrustedText
    Evidence            []UntrustedEvidence
}
```

Provider adapters must preserve role boundaries. An adapter that flattens everything into one string is incompatible with a protocol requiring role isolation.

## 8. Provider execution and attestations

### 8.1 Provider executor

```go
type ProviderExecutor interface {
    Identity() ProviderIdentity
    Execute(context.Context, VerifiedRequest) (VerifiedResponse, error)
}
```

The request records exact model identity, decoding settings, tool/schema definitions, and rendered messages.

### 8.2 Response custody

Responses are hashed immediately and stored in an append-only artifact store. Parser input refers to the stored response digest. Repair attempts are separate child requests, not overwritten retries.

```text
attempt 1 response -> structural failure
       |
       +--> repair request -> attempt 2 response -> parsed output
```

The manifest retains both attempts.

### 8.3 Model identity limitations

When a hosted provider cannot cryptographically attest a model revision, the manifest records the provider's claimed model ID and request ID. Verification can prove what was requested and recorded, not which internal weights actually served the call.

## 9. Hardened cache and repeatability

### 9.1 Content-addressed cache

Cache entries are immutable and keyed by the full execution request digest. Cache reads return a verified response bundle. Mutable replacement is prohibited; a refreshed generation receives a new response digest and manifest.

### 9.2 Audit nonces

Repeatability probes add a run nonce excluded from the semantic instance but included in execution identity:

```go
type AuditExecution struct {
    SemanticInstanceDigest Digest
    RunNonce [32]byte
    CachePolicy CacheForbidden
}
```

This makes accidental cache reuse impossible at the type/API level.

### 9.3 Determinism claims

The manifest distinguishes:

- requested deterministic settings;
- provider-reported seed support;
- observed repeatability;
- cache replay.

A cached response cannot be labeled a repeat measurement.

## 10. Hardened report model

### 10.1 Complete report

```go
type SignedReport struct {
    Report          VerifiedReportBody
    ManifestDigest  Digest
    CalibrationRefs []Digest
    AuditRefs       []Digest
    ProducedAt      time.Time
    Producer        ServiceIdentity
    Signature       Signature
}
```

### 10.2 Signatures

Use standard, externally reviewed signing libraries. Keys belong to an execution service, not application source code. Support rotation and revocation records. Never invent custom cryptography.

### 10.3 Independent verification

```go
func VerifySignedReport(
    report SignedReport,
    resolver ArtifactResolver,
    trust TrustPolicy,
) (VerifiedReport, error)
```

Verification checks:

1. signature and producer trust;
2. report canonical digest;
3. execution manifest root;
4. all linked component digests;
5. contract/protocol/renderer binding;
6. report dimensions against the contract;
7. claim/evidence cross-references;
8. allowed calibration and audit references.

## 11. Hardened calibration and audit populations

Calibration input is a manifest, not a loose map:

```go
type CalibrationPopulation struct {
    ProtocolDigest Digest
    ContractDigest Digest
    ReportDigests  []Digest
    GoldSetDigest  Digest
    InclusionRule  Digest
}
```

The inclusion rule records sampling, exclusions, time window, and group definitions. Calibration refuses reports outside the declared population.

Audit reports similarly record:

- probe transformations;
- semantic invariants;
- fresh-execution nonces;
- judge identities;
- missing-output accounting;
- error correlation estimates where labels exist.

## 12. Execution-service architecture

A production deployment could separate roles:

```text
                         +------------------+
client ---------------->| intake validator |
                         +--------+---------+
                                  |
                                  v
                         +------------------+
                         | instrument store |
                         +--------+---------+
                                  |
                                  v
+----------------+       +------------------+       +----------------+
| evidence store |------>| execution worker |------>| provider API   |
+----------------+       +--------+---------+       +----------------+
                                  |
                                  v
                         +------------------+
                         | append-only store|
                         +--------+---------+
                                  |
                                  v
                         +------------------+
                         | signing service  |
                         +------------------+
```

Security boundaries:

- intake cannot sign;
- execution worker cannot mutate stored contracts;
- artifact store is append-only;
- signing service signs only verified manifests;
- hidden promotion data is available through a limited service API, not raw files;
- all access is auditable.

## 13. API sketch

```go
// Decode and verify durable inputs.
contract, err := registry.LoadVerifiedContract(contractDigest)
protocol, err := registry.LoadVerifiedProtocol(protocolDigest)
renderer, err := registry.LoadVerifiedRenderer(rendererDigest)

// Bind one executable instrument.
instrument, err := judgekit.Compile(judgekit.CompileRequest{
    Contract: contract,
    Protocol: protocol,
    Renderer: renderer,
    Parser: parser,
    Aggregator: aggregator,
    Provider: provider,
})

// Verify the concrete evaluation item.
instance, err := judgekit.VerifyInstance(rawInstance, contract)

// Execute under an explicit custody policy.
bundle, err := instrument.Execute(ctx, instance, judgekit.ExecutionPolicy{
    Cache: judgekit.CacheForbidden,
    ArtifactRetention: judgekit.RetainEncrypted,
    Sign: true,
})

// Independent consumer verifies before use.
report, err := verifier.VerifySignedReport(bundle.Report)
```

## 14. Storage and privacy

Hardening increases stored provenance, which increases privacy risk. Requirements include:

- encryption at rest and in transit;
- tenant-scoped artifact namespaces;
- content-sensitive retention classes;
- redacted public manifests with private artifact references;
- deletion workflows that preserve tombstone identity without retaining content;
- no chain-of-thought requirement;
- separation of hidden datasets from optimizer-visible artifacts;
- access logs for contracts, prompts, evidence, and reports.

A digest of sensitive content can itself enable confirmation attacks. Do not publish private-content digests without a threat review.

## 15. Reliability, validity, and hardening

The hardened system improves integrity and reproducibility. It does not automatically improve validity.

```text
hardening answers:  Did these exact components produce this exact report?
calibration answers: How does this evaluator behave on a labeled population?
validity asks:      Does this measurement support the intended interpretation?
optimization asks:  Does it remain valid under selection pressure?
```

A perfectly signed invalid judge remains invalid. Promotion must still use independent evidence, constraints, and adversarial tests.

## 16. Migration path if ever needed

Do not jump directly from the lightweight library to the complete system. Use gates:

### Stage 1: verified loaders

- typed digests;
- full digest recomputation on persisted documents;
- defensive copies;
- independent report verifier.

### Stage 2: compiled instruments

- bind contract, protocol, renderer, parser, aggregator;
- produce execution manifests;
- retain raw attempts content-addressably.

### Stage 3: custody service

- append-only artifact store;
- execution worker isolation;
- signing service;
- key rotation.

### Stage 4: cross-trust operation

- trust policies;
- signed evidence admission;
- interoperable registries;
- external verifier tooling.

At each stage, require a consumer and threat that justify the added complexity.

## 17. Cost and complexity assessment

Expected costs include:

- more types and constructors;
- serialization migration machinery;
- canonicalization compatibility obligations;
- artifact-store operations;
- key management and incident response;
- provider-response retention costs;
- privacy review;
- slower API experimentation;
- significantly larger integration-test surface.

For internal research, these costs likely exceed the benefit. The architecture should remain documentation until the threat model changes.

## 18. Decision records

### Decision: Do not implement hardening without a cross-trust requirement

- **Context:** The design is technically possible but expensive.
- **Options considered:** Implement immediately; ignore future integrity needs; preserve a reference architecture.
- **Decision:** Preserve this design but require an explicit consumer, threat model, and owner before implementation.
- **Rationale:** Internal research needs semantic correctness more than adversarial custody.
- **Consequences:** The current system remains unsuitable as a signed attestation platform.
- **Status:** accepted

### Decision: Hardening proves identity, not truth

- **Context:** Cryptographic mechanisms can create false confidence in evaluator validity.
- **Options considered:** Treat signed reports as authoritative; separate integrity from measurement validity.
- **Decision:** Signatures and manifests attest provenance only; calibration and decision policy remain separate.
- **Rationale:** A traceable error is still an error.
- **Consequences:** Consumers must define independent acceptance criteria.
- **Status:** proposed

### Decision: Compile executable instruments

- **Context:** Independently valid components can be incompatible.
- **Options considered:** Runtime checks scattered across evaluation; public mutable assembly; compiled binding.
- **Decision:** A hardened system accepts only compiled instruments that bind verified components.
- **Rationale:** It makes invalid combinations unrepresentable at execution.
- **Consequences:** Dynamic configuration changes require recompilation and a new instrument identity.
- **Status:** proposed

## 19. Intern review questions

Before implementing any hardened feature, answer:

1. Which concrete attacker or accidental failure does it address?
2. Which trust boundary changes?
3. What does the feature prove, and what does it not prove?
4. Who owns keys, migrations, storage, and incident response?
5. Can a lighter boundary check solve the problem?
6. Does the feature increase sensitive-data retention?
7. How will old reports remain verifiable?
8. What consumer will test the feature?
9. What operational failure occurs when verification infrastructure is unavailable?
10. How will the API avoid implying that integrity equals validity?

## 20. Definition of done for a hypothetical hardened release

A hardened release would require:

- documented threat model and security review;
- typed, domain-separated digests;
- strict canonicalization versions;
- verified immutable domain values;
- compiled instrument binding;
- prompt/evidence role isolation;
- content-addressed raw attempts and manifests;
- append-only encrypted artifact storage;
- signing and independent verification;
- key rotation and revocation;
- homogeneous calibration/audit population manifests;
- privacy and retention controls;
- migration tests across schema versions;
- external penetration and cryptographic implementation review;
- evidence that a real consumer requires all of the above.

## 21. References

Current implementation anchors:

- `internal/canonicaljson/canonicaljson.go`
- `spec/load.go`
- `protocol/load.go`
- `eval/digest.go`
- `assessment/digest.go`
- `judging/claimjudge.go`
- `judging/cache.go`
- `calibration/report.go`
- `audit/reliability.go`
- `suite/suite.go`

Companion documents:

- `design-doc/02-pr-2-code-review-stabilization-guide.md`
- `design-doc/03-lightweight-research-guarantees-after-merge.md`
- `design-doc/01-judgekit-architecture-and-implementation-guide.md`
