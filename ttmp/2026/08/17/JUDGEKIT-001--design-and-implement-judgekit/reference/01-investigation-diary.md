---
Title: Investigation Diary
Ticket: JUDGEKIT-001
Status: active
Topics:
    - research
    - llm
    - evaluation
    - optimization
    - safety
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: repo://.ttmp.yaml
      Note: Repository-local docmgr root configuration
    - Path: repo://ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/01-judgekit-architecture-and-implementation-guide.md
      Note: Primary judgekit design deliverable
ExternalSources: []
Summary: Chronological record of initializing the judgekit ticket and designing the library architecture.
LastUpdated: 2026-08-17T18:58:00-04:00
WhatFor: Preserve design evidence, repository findings, failures, decisions, validation, and delivery details for judgekit implementation.
WhenToUse: Before resuming JUDGEKIT-001 or implementing any phase of the judgekit design.
---


# Diary

## Goal

Create a local judgekit docmgr ticket and produce an intern-ready architecture and implementation guide for building judgekit as a provider-neutral evaluation library.

## Step 1: Initialize and Design Judgekit

I inspected the existing judgekit repository and found an untouched Go template: placeholder module and command names, a template README, and no local docmgr workspace. I then created a repository-local ticket and designed the library from the concrete evaluation structures currently embedded in CoinVault and the experiment boundaries already present in ragopt and ragkit.

The design keeps measurement and judging in judgekit, optimization in ragopt, RAG primitives in ragkit, and product semantics in consuming applications. It provides package structure, APIs, invariants, migration phases, tests, security requirements, decision records, and a CoinVault pilot path.

### Prompt Context

**User prompt (verbatim):** "Create a new ticket in ./judgekit/ to create judgekit itself. Create  a detailed analysis / design / implementation guide that is for a new intern, explaining all the parts of the system needed to understand what it is, with prose paragraphs and bullet point sand pseudocode and diagrams and api references and file references. It should be very clear and technical. Store in the ticket and the nupload to remarkable.

[REMINDER] Output a <summary>...</summary> block at the VERY END of your response. This is mandatory."

**Assistant interpretation:** Initialize docmgr locally in the existing judgekit repository, create JUDGEKIT-001, write a detailed architecture and implementation guide for the new library, validate the ticket, and upload the deliverable to reMarkable.

**Inferred user intent:** Establish an implementation-ready foundation for extracting reusable machine-judging and measurement structures from CoinVault without mixing product semantics, optimization, or RAG primitives into the new library.

### What I did
- Inspected judgekit's template repository, module, README, branch, remotes, and baseline test state.
- Initialized a repository-local `ttmp/` and created `JUDGEKIT-001`.
- Added the primary architecture guide and this diary.
- Designed provider-neutral spec, eval, protocol, assessment, judging, suite, audit, and calibration packages.
- Defined public Go APIs, strict validation, semantic identities, integration boundaries, and ten implementation phases.
- Added testing, documentation, security, review, and release-readiness guidance.

### Why
- Judgekit needs a concrete consumer-driven design before template normalization or code extraction begins.
- Explicit ownership prevents generic packages from absorbing CoinVault-specific prompts, authorization, promotion policy, or RAG internals.

### What worked
- `GOWORK=off go test ./...` passed on the template baseline.
- The second docmgr initialization used a repository-local config and created the ticket under `judgekit/ttmp`.
- The primary guide provides a library-first path with a bounded CoinVault pilot.
- `docmgr doctor --ticket JUDGEKIT-001 --stale-after 30` passed after adding repository-local vocabulary entries.
- Remarquee reported: `OK: uploaded JUDGEKIT-001 Architecture and Implementation Guide.pdf -> /ai/2026/08/17/JUDGEKIT-001`.

### What didn't work
- The first `docmgr init --root ttmp` inherited the workspace-level CoinVault root and wrote `judgekit/.ttmp.yaml` pointing to `../coinvault/ttmp`. It created a temporary `JUDGEKIT-001` under CoinVault.
- I removed only that newly created temporary ticket and config, wrote a local `.ttmp.yaml` with `root: ttmp`, reran initialization, and verified `root=/home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp` before recreating the ticket.
- The first real reMarkable upload failed because the verbatim prompt contained literal `\\n` sequences that Pandoc interpreted as an undefined LaTeX command: `! Undefined control sequence. l.2011 and the nupload to remarkable.\\n`.
- I replaced the escaped sequences with actual Markdown line breaks and reran the same upload successfully.

### What I learned
- A nested repository under a workspace with a parent `.ttmp.yaml` requires an explicit local config before docmgr initialization.
- Judgekit's first implementation should remove the placeholder CLI rather than normalize it, because the project is library-first.
- The measurement-contract/protocol distinction is the central API boundary.

### What was tricky to build
- The design must be generic enough for non-RAG tasks while still supporting CoinVault's claim/evidence workflow. Evidence therefore uses generic kinds and provenance rather than ragkit types.
- Judgekit must support reports consumed by ragopt without becoming an optimization or promotion package. Rich reports remain application artifacts; ragopt receives metric projections and identities.

### What warrants a second pair of eyes
- Review whether the package count is appropriate for v0 or whether `suite` and `audit` should wait until after the CoinVault pilot.
- Review strict parsing defaults, especially the decision not to search arbitrary surrounding prose for a JSON object.
- Review whether `provider/geppetto` belongs in this repository or in a separate adapter repository after the core stabilizes.

### What should be done in the future
- Implement Phase 0 repository normalization as a focused commit, followed by internal canonical primitives and the `spec` package.

### Code review instructions
- Start with Sections 2–5 for the conceptual and API model.
- Review Section 9 for integration boundaries and Section 10 for implementation ordering.
- Confirm no core package proposal requires CoinVault, ragopt, ragkit, or provider frameworks.

### Technical details
- Current branch: `task/deploy-dev-indexer`.
- Upstream remote: `git@github.com:go-go-golems/judgekit`.
- Current module placeholder: `github.com/go-go-golems/XXX`.
- Intended module: `github.com/go-go-golems/judgekit`.
- Design guide size: 1,218 lines and 4,812 words.
- reMarkable destination: `/ai/2026/08/17/JUDGEKIT-001/JUDGEKIT-001 Architecture and Implementation Guide.pdf`.

## Step 2: Normalize the repository (Phase 0)

The repo was an untouched Go template: module `github.com/go-go-golems/XXX`, a placeholder `cmd/XXX` binary, a `pkg/` placeholder, and logcopter wiring that pulled glazed/cobra/bubbletea into the dependency graph. I renamed the module, removed the placeholder binary and packages, and simplified the build tooling so the core can stay stdlib-only. A thin `cmd/judgekit` CLI will be added later only to host the Glazed help system; the design's "no CLI in the first phase" decision is respected for domain logic, while the help entry the user explicitly requested gets a real host.

### Prompt Context

**User prompt (verbatim):** "Implement JUDGEKIT-001, commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill). Create a GLOSSARY.md file that uses the textbook we imported previously to properly define the constructs. Also update the README.md and then write a user-guide and a getting-started glazed help entry (see skill) and a reference document, for developers using the package."

**Assistant interpretation:** Implement the judgekit design end to end inside the judgekit repo, committing per phase, keeping a detailed diary, and producing GLOSSARY.md, an updated README, a user guide, a getting-started Glazed help entry, and a developer reference.

**Inferred user intent:** Turn the JUDGEKIT-001 design doc into a real, tested, provider-neutral Go library with intern-facing documentation, without touching the sibling repositories a colleague is working on.

**Commit (code):** 05a6560 — "chore: normalize judgekit repository"

### What I did
- Rewrote `go.mod` to `module github.com/go-go-golems/judgekit` with no requires.
- `git rm`-ed `cmd/XXX/`, `pkg/`, and `logcopter_generate.go`; deleted `go.sum`.
- Rewrote `Makefile`: dropped logcopter-generate/logcopter-check, fixed module paths, added `cmd/judgekit` build/install targets, dropped `gifs`.
- Simplified `.goreleaser.yaml` to a single CGO-free `judgekit` build for linux/darwin amd64+arm64.
- Updated `AGENT.md` to describe the library-first layout and the core dependency boundary.

### Why
- logcopter transitively imports glazed, cobra, and bubbletea; keeping it would force those into core, violating the design's "core depends only on stdlib + internal helpers" and "no accidental Cobra, Glazed, Bubble Tea dependency in core" rules.
- A library-first repo with top-level domain packages matches the design doc's package map.

### What worked
- `GOWORK=off go mod tidy` produced a minimal `go.mod` with no dependencies.
- After removing all Go files, the module built cleanly once packages were re-added in Phase 1.

### What didn't work
- `go test ./...` returned exit code 1 with "no packages to test" while the repo had no Go files; this was expected and resolved once Phase 1 added `internal/` packages.

### What I learned
- The workspace `go.work` already includes `./judgekit`, so sibling modules are available locally, but I must keep judgekit self-contained and `GOWORK=off`-buildable per the Makefile convention.
- Removing logcopter is itself an alignment with the design boundary, not just cleanup.

### What was tricky to build
- Choosing whether to keep a CLI at all. The design says "no initial CLI", but the user explicitly asked for a Glazed help entry, which the help skill wants wired into a Cobra root. Resolution: keep core provider-neutral and stdlib-only, and add a thin `cmd/judgekit` CLI that is the only place allowed to import Glazed. The boundary test will enforce that core never imports it.

### What warrants a second pair of eyes
- The decision to drop logcopter entirely vs. keeping structured logging. For pure value types, logging is an anti-pattern in core; confirm that operational logging belongs only in adapters/CLI.
- The simplified goreleaser dropped Homebrew/nfpm/fury publishing. Confirm whether source + binary release is enough for v0.

### What should be done in the future
- Add the `cmd/judgekit` CLI and wire Glazed help when the help entries are written.
- Add a boundary test that rejects forbidden imports in core packages.

### Code review instructions
- Inspect `go.mod` (no deps), `Makefile` (no logcopter), `.goreleaser.yaml` (CGO off, `cmd/judgekit`).
- Verify `git ls-files` no longer contains `cmd/XXX` or `pkg/`.

### Technical details
- Commands: `GOWORK=off go mod tidy`, `GOWORK=off go build ./...`, `GOWORK=off go test ./internal/...`.
- Branch unchanged: `task/deploy-dev-indexer`.

## Step 3: Implement internal canonical primitives (Phase 1)

With the repo normalized, I added the three internal helper packages that every core value type depends on: identifier validation, canonical JSON for stable digests, and strict structured-output parsing. These stay under `internal/` so they cannot be imported outside the module, keeping the public surface minimal until repeated use justifies promotion.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Assistant interpretation:** Build the foundational helpers (identifier, canonicaljson, strictdecode) with tests so the spec/eval/protocol/assessment packages can rely on stable identities and strict parsing.

**Inferred user intent:** Establish deterministic, testable primitives before building domain value types on top of them.

**Commit (code):** 08a92dd — "feat(internal): add canonicaljson, strictdecode, identifier primitives"

### What I did
- Wrote `internal/identifier/identifier.go` + tests: bounded portable identifiers (lowercase, 1..128 bytes, no leading/trailing separators).
- Wrote `internal/canonicaljson/canonicaljson.go` + tests: deterministic JSON (sorted keys, compact, no HTML escaping, NaN/Inf rejected), `Marshal`, `Sum` ("sha256:..."), `MustSum`.
- Wrote `internal/strictdecode/strictdecode.go` + tests: `StripSingleCodeFence`, `DecodeJSONObjectStrict[T]`, `ValidateSingleJSONValue`, typed `StructuralError` + `IsStructural`.

### Why
- Stable semantic digests require canonical encoding independent of YAML/struct key order; the generic round-trip through `any` makes struct field declaration order irrelevant.
- Strict parsing (reject prose, unknown fields, trailing data) prevents the silent tolerance that hides model failures; CoinVault's `parseJudgeJSON` searches prose for `{...}`, which judgekit deliberately does not do by default.

### What worked
- All internal tests pass: `GOWORK=off go test ./internal/...` green.
- `go vet` clean; `gofmt -l` clean after `go fmt`.

### What didn't work
- First build failed: `encodeObject` was missing a trailing `return nil` after the loop (`missing return` at line 138). Fixed by adding the return.
- First test compile failed: `TestMarshalRejectsNonFinite` used `1.0/0.0` and `0.0/0.0`, which are compile-time division-by-zero errors. Fixed by using `math.NaN()`/`math.Inf`.

### What I learned
- Go's `encoding/json` already sorts `map[string]any` keys, but to make struct field order irrelevant I round-trip through `any` and re-encode with sorted keys and no HTML escaping.
- `json.Decoder.Decode` called a second time returns `io.EOF` exactly when there is no trailing data, which is the clean way to enforce "exactly one JSON value".

### What was tricky to build
- Canonical float formatting: I delegate numbers to `json.Marshal` for shortest representation (5, 0.5, 0.86) but reject NaN/Inf before encoding, so digests stay stable without hand-rolling float formatting.
- Code-fence stripping must only strip a fence that wraps the *whole* response; an embedded fence inside prose must be left intact so strict decode rejects the surrounding prose. The function returns the input unchanged unless trimmed input starts with ``` and ends with ```.

### What warrants a second pair of eyes
- `canonicaljson` re-encodes numbers as float64 after the `any` round-trip, which loses int64 precision for very large numbers. Acceptable for judgekit value ranges (scores 0..1, counts), but confirm no construct uses >2^53 integer identity.
- `strictdecode` classifies errors by substring matching on the json error message (`"unknown field"`, `"cannot find"`). This is fragile across Go versions; consider mapping via typed errors if it breaks.

### What should be done in the future
- Add fuzz tests for `StripSingleCodeFence`, `canonicaljson.Marshal`, and `DecodeJSONObjectStrict` per the design's Phase 11.2.

### Code review instructions
- Start at `internal/canonicaljson/canonicaljson.go` (`Marshal` -> `encode` -> `encodeObject`/`encodeString`).
- Validate `internal/strictdecode/strictdecode_test.go` covers prose rejection, array rejection, unknown field, trailing data, and fence stripping.
- Run `GOWORK=off go test ./internal/...`.

### Technical details
- Identifier regex: `^[a-z0-9]([a-z0-9_-]*[a-z0-9])?$`.
- Digest format: `sha256:` + lowercase hex of SHA-256 over canonical bytes.
- No external dependencies; only stdlib (`crypto/sha256`, `encoding/json`, `math`, `sort`, `regexp`).

## Step 4: Implement the spec package (Phase 2)

spec owns the first link of the inference chain: what an evaluator measures. I added constructs and measurement contracts with strict validation, dual digests, and a strict YAML/JSON loader, using a CoinVault-style faithfulness fixture as the proving target.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** 62c54a1 — "feat(spec): add constructs and measurement contracts with strict loading"

### What I did
- Added `Construct`/`ConstructID`/`Direction`/`Range` and `MeasurementContract` (evidence policy, labels, aggregations, exclusions, empty policy).
- Strict validation: supported `api_version`, duplicate construct IDs, dangling label/aggregation/exclusion references, evidence-kind overlap, fraction aggregations referencing declared labels (comma-separated label lists so faithfulness = entailed / (entailed+contradicted+insufficient)).
- Dual digests: `SemanticDigest` (canonical JSON, order-independent) and `ByteDigest` (raw bytes).
- Strict `LoadContract`/`LoadContractFromBytes` for `.json`/`.yaml`/`.yml` with unknown-field rejection in both formats.
- Fixture `spec/testdata/faithfulness.yaml` mirroring a CoinVault faithfulness contract.

### What worked
- `gopkg.in/yaml.v3` round-trips `map[ConstructID][]string` keys correctly, so typed map keys survive YAML.
- `go test ./spec` green; semantic digest is stable across YAML key order but changes on semantic change.

### What didn't work
- First fixture used a single-label denominator for fraction; faithfulness needs a label SET, so I extended the validator to accept comma-separated label lists.

### What was tricky to build
- Keeping `spec` value types stdlib-only while supporting strict YAML. Resolution: only `load.go` imports `yaml.v3`; the boundary test forbids frameworks, not data-format libraries.

### What warrants a second pair of eyes
- `canonicaljson` re-encodes numbers as float64, losing int64 precision >2^53. Acceptable for judgekit value ranges; confirm no construct uses large integer identity.
- The supported `api_version` is a single constant (`judgekit.measurement/v1`); a future incompatible schema will need a migration path rather than silent acceptance.

### Code review instructions
- Start at `spec/validate.go` (`ValidateContract`), then `spec/digest.go`, then `spec/load.go`.
- Run `GOWORK=off go test ./spec`.

## Step 5: Implement the eval package (Phase 3)

eval defines one concrete item being judged. It depends only on stdlib + internal helpers (parallel to spec, not on it).

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** 85d832f — "feat(eval): add artifacts, evidence, required facts, and instances"

### What I did
- `Artifact` (content-addressed text/URI; `NewTextArtifact` computes SHA-256 + size; validation rejects stale digests), `EvidenceItem`/`EvidenceSet` (ordered, unique IDs, provenance), `RequiredFact`, `Instance`.
- Cross-reference validation: `RequiredFact.EvidenceIDs` must resolve to real evidence item IDs.
- Non-circular content digests for evidence sets and instances; `NewEvidenceSet`/`NewInstance` compute and verify digests.

### What worked
- All eval tests green; instance digest is deterministic and changes with metadata.

### What was tricky to build
- The instance digest must exclude its own `Digest` field to avoid circularity; I used a separate `instanceDigestInput` struct.

### Code review instructions
- Start at `eval/validate.go` (`ValidateInstance` cross-references), `eval/digest.go`.

## Step 6: Implement the protocol package (Phase 4)

protocol identifies how an evaluator measures. It references the contract by digest only, so it does not import spec.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** 55d6788 — "feat(protocol): add complete evaluator protocol identity"

### What I did
- `ModelIdentity`, `DecodingPolicy`, `RetryPolicy`, `Protocol` (api_version, name, measurement_digest, prompt digests, decoding, evidence_order, parser/aggregator versions, retry).
- Validation: supported version, identifier name, sha256 digests, shuffled-evidence-order requires a seed, positive max_tokens, retry attempts >= 1.
- Dual digests + strict loader; digest stable across prompt-map insertion order, changes on any semantic field.

### What didn't work
- First `ByteDigest` used `canonicaljson.MustSum(raw)` which base64-encodes `[]byte`; fixed to raw-byte SHA-256 (same bug pattern as spec, caught here first).
- `&baseProtocol()` in a test cannot take the address of a function-call result; assigned to a variable first.

### Code review instructions
- `protocol/validate.go`, `protocol/load.go`, `protocol/protocol_test.go`.

## Step 7: Implement the assessment package (Phase 5)

assessment is the convergence point: it imports spec + eval but not judging.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** d98b80d — "feat(assessment): add claims, three-way support, dimensions, and sealed reports"

### What I did
- `Claim`/`Span`, three-way `SupportLabel` (entailed/contradicted/insufficient), `ClaimAssessment` (entailed+contradicted require cited evidence), `DimensionResult`, sealed `Report`.
- Fail-closed validation: duplicate claims, unknown claim results, missing results, invalid spans, non-finite values, not-applicable-with-value, evidence references outside an allowed set supplied by the judging layer.
- `Seal` computes a non-circular digest; `ValidateReport` checks a sealed report. `EvidenceIDSet` builds the allowed set from an instance's evidence.

### What was tricky to build
- Sealing vs. the digest-presence check: `ValidateReport` requires a digest, but `Seal` validates before the digest exists. Split into `validateReportBody` (all checks except digest presence) + `ValidateReport` (body + digest check); `Seal` calls the body, computes the digest, sets it.

### What didn't work
- A test passed an empty allowed-evidence set while another verdict still cited `e1`, causing a false "unknown evidence" failure; fixed the test to allow `e1`.

### Code review instructions
- `assessment/validate.go` (`ValidateReport`, `ValidateClaimAssessment`), `assessment/digest.go` (`Seal`).

## Step 8: Implement the judging package and boundary test (Phase 6)

judging runs evaluators and depends on all value packages. It is the only place where the model is called, through a provider-neutral `Generator`.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** 96a72ef — "feat(judging): add provider-neutral judge interfaces and two-stage claim judge"

### What I did
- `Generator`/`Judge`/`Critic`/`Verifier` interfaces, `CacheKey`+`Cache` (`NoopCache`, JSON-backed `MemoryCache`), strict-parsing re-exports, `DefaultRepairer`.
- `FakeGenerator` (no credentials) and the two-stage `ClaimJudge`: extract claims (evidence hidden), judge support against evidence, aggregate per contract (fraction counts claim labels; direct takes model-emitted dimensions), seal the report.
- Repair: structural failures retry with a repaired prompt; semantic failures fail closed at seal.
- `spec.MethodDirect` added for non-claim-aggregated constructs (relevance, abstention).
- `boundary_test.go` at the module root: runs `go list -json` over core packages and rejects forbidden imports (glazed, cobra, bubbletea, geppetto, pinocchio, coinvault, ragopt, ragkit, provider SDKs). `cmd/judgekit` is intentionally excluded so it can host Glazed help.

### What worked
- End-to-end claim-judge test: two claims → faithfulness 0.5, relevance 0.9, abstention "attempted"; extractor prompt verified to not leak evidence; caching test confirms no regeneration on a second run; repair test confirms a malformed extract is retried once; bad-label test confirms fail-closed.

### What didn't work
- Used `spec.MethodDirect` before declaring it; added it to `spec/contract.go` and `validMethods`.
- `errExhausted` referenced in the sequence generator test before being defined; added it in the test file.

### What was tricky to build
- Generic `generateAndDecode[T]` must be a free function (Go methods cannot have type parameters); it takes the `*ClaimJudge` to access cache/repair/protocol.
- Caching + repair interaction: a repaired prompt changes the prompt digest and thus the cache key, so the repaired generation is a new cache entry rather than a stale hit.

### What warrants a second pair of eyes
- The `MemoryCache.Load` JSON-round-trips values; for very large raw responses this is wasteful but correct. A production cache should store raw bytes.
- `generateAndDecode` retries on every structural error up to `MaximumAttempts`; confirm that semantic-but-decoded-as-structural cases (e.g., a valid JSON object with the wrong shape) are caught at seal, not retried indefinitely.

### Code review instructions
- `judging/claimjudge.go` (`Evaluate`, `generateAndDecode`, `aggregate`), `judging/cache.go`, `judging/repair.go`, `boundary_test.go`.
- Run `GOWORK=off go test ./...` (all packages, including the boundary test).

## Step 9: Add GLOSSARY, README, help entries, help-host CLI, and example (docs)

With the core implemented and tested, I wrote the intern-facing documentation the user asked for and a thin CLI so the Glazed help entry is real and queryable rather than just a markdown file.

### Prompt Context

**User prompt (verbatim):** (see Step 2)

**Commit (code):** a8011c0 — "docs: add GLOSSARY, README, user guide, getting-started help, developer reference, and help-host CLI"

### What I did
- `GLOSSARY.md`: measurement-theory terms (construct, latent utility, proxy, protocol, judge/critic/verifier, reliability, validity, calibration, bias, construct shift) grounded in the imported textbook Chapter 1 reader edition (Definitions 1.1–1.19), each mapped to a judgekit type/package.
- `README.md`: problem, non-goals, package map, quick start (contract, generator adapter, claim judge), status.
- `pkg/doc/`: three Glazed help entries — `tutorials/01-getting-started.md` (Tutorial, slug `getting-started`), `topics/01-user-guide.md` (GeneralTopic, slug `user-guide`), `reference/01-developer-reference.md` (GeneralTopic, slug `developer-reference`), each with frontmatter per the glazed-help-authoring skill.
- `pkg/doc/doc.go`: `embed` of the help entries + `AddDocToHelpSystem`.
- `cmd/judgekit/main.go`: thin CLI wiring `help.NewHelpSystem()` + `doc.AddDocToHelpSystem` + `help_cmd.SetupCobraRootCommand` + `logging.AddLoggingSectionToRootCommand`, so `judgekit help <slug>` works.
- `examples/claim-judge/`: runnable end-to-end example (`contract.yaml` + `example_test.go`) with a fake generator.

### Why
- The user explicitly asked for a GLOSSARY using the textbook, an updated README, a user guide, a getting-started Glazed help entry, and a developer reference. The glazed-help-authoring skill's definition of done requires help entries to be wired into a Cobra root, not left as loose markdown — hence the thin CLI.

### What worked
- `judgekit help getting-started|user-guide|developer-reference` all render via the embedded FS.
- Full test suite and the boundary test still pass after adding Glazed/Cobra as dependencies for the CLI only.

### What was tricky to build
- Reconciling the design's "no initial CLI" decision with the user's request for a Glazed help entry (which the help skill wants wired into Cobra). Resolution: keep core provider-neutral and stdlib-only; add `cmd/judgekit` as the only place allowed to import Glazed/Cobra, hosting just help. The boundary test excludes `cmd/` and `pkg/doc`, so core stays clean.
- `go:embed` can only embed within a package's subtree, so the embed lives in `pkg/doc/doc.go` (next to the markdown) and the CLI imports it, rather than embedding from `cmd/judgekit`.

### What warrants a second pair of eyes
- The CLI depends on Glazed v1.4.1 (a large transitive tree). That is acceptable because it is confined to `cmd/` and `pkg/doc`; confirm no core package transitively pulls it (the boundary test covers direct imports; a transitive check via `go list -deps` of core packages would be stronger).

### Code review instructions
- `GLOSSARY.md`, `README.md`, `pkg/doc/{topics,tutorials,reference}/*.md`, `pkg/doc/doc.go`, `cmd/judgekit/main.go`, `examples/claim-judge/`.
- Run `GOWORK=off go run ./cmd/judgekit help getting-started` and `GOWORK=off go test ./examples/claim-judge -v`.

## Step 10: Ticket bookkeeping and validation

### Prompt Context

**User prompt (verbatim):** (see Step 2)

### What I did
- Related 12 key implementation/doc files to the design doc via `docmgr doc relate`.
- Added and checked implementation tasks (std1, qoyn); updated the changelog.
- `docmgr doctor --ticket JUDGEKIT-001 --stale-after 30` passes; all tasks complete.

### What worked
- All JUDGEKIT-001 tasks complete; doctor clean; `GOWORK=off go test ./...` green including the boundary test.

### What should be done in the future
- Implement the `audit` (reliability/bias probes, panels) and `calibration` (gold, extraction recall, confusion, Brier, ECE) packages per the design's Phases 8–9.
- Add fuzz tests for `canonicaljson`, `strictdecode`, and report sealing per the design's Phase 11.2.
- Run a CoinVault pilot that ports the claim extraction + support path onto judgekit and removes the duplicated local generic structures (design Phase 7), gated by characterization fixtures.

## Step 11: Implement the calibration package (Phase 8)

The user asked to implement the three remaining core packages. I started with `calibration` because it sits at the bottom of the dependency graph (it consumes `assessment.Report` and `spec.ConstructID`; it does not run judges), so it could be built and tested in isolation before the packages that depend on `judging`.

### Prompt Context

**User prompt (verbatim):** "Ok, implement. commit at appropriate intervals and keep a detailed diary as you work (using the diary format from the skill) (make sure to use the skill format)"

**Assistant interpretation:** Implement the audit/calibration/suite packages specified in the design doc (Sections 6-8, Phases 8-9), commit per package, and record each as a diary step in the strict skill format.

**Inferred user intent:** Complete the judgekit core so the planned reliability, calibration, and evaluator-suite capabilities exist as tested, provider-neutral code, not just design prose.

**Commit (code):** ffdf863 — "feat(calibration): add gold records, extraction recall, confusion, Brier, ECE"

### What I did
- Wrote `calibration/doc.go`, `gold.go`, `confusion.go`, `recall.go`, `scoring.go`, `report.go`, and `calibration_test.go`.
- `GoldClaim`/`GoldDimension`/`GoldSet` retain `ReviewerIDs`; `Adjudicated` does not erase them. Gold sets are content-addressed.
- `Confusion` is a 2x2 over matched claims (entailed vs non-entailed) with `Sensitivity`, `Specificity`, `FalseSupportRate`.
- `ExtractionRecall` = matched / len(gold); `MatchByID` default, `MatchByText` brittle fallback.
- `BrierScore` and `ExpectedCalibrationError` over (confidence, outcome) pairs; `*float64` and nil when no confidence.
- `Calibrate` matches gold to model verdicts per instance, builds confusion over matched claims, recall over all gold, Brier/ECE over claims with confidence, seals the report.
- Exported `assessment.ValidSupportLabel` so calibration can validate labels without re-declaring the set.
- Removed a stray 43 MB `judgekit` build binary from the repo root and added `/judgekit` to `.gitignore`.

### Why
- Calibration is the link to human/objective labels and the place where "a judge can look accurate by extracting fewer claims" is exposed. It belongs in judgekit, not the application, because the metrics are generic over reports and gold records.
- Keeping `ReviewerIDs` even after adjudication is a measurement-theory invariant (Chapter 1 §1.5): agreement must be measurable, and adjudication must not erase the disagreement that produced it.

### What worked
- `GOWORK=off go test ./calibration/` green; full suite still green. The `TestCalibrateMissedClaimReducesRecallNotConfusion` test pins the key semantic: a missed claim lowers recall but does not enter the confusion matrix, so the two metrics measure different things.

### What didn't work
- `calibration/scoring.go` used `assessment.ClaimAssessment` in `ConfidenceOutcomePairs` without importing `assessment` -> `undefined: assessment`. Fixed by adding the import.
- `&goldClaim(...)` in tests: cannot take the address of a function-call result. Fixed by assigning to a variable first.
- `TestBrierScore` compared `bs != 0.01` exactly; 0.01 is not exactly representable in float64, so the test failed with `0.009999999999999998`. Fixed with a tolerance band (`> 0.0101 || < 0.0099`).

### What I learned
- The clean separation between extraction recall and confusion: recall is over all gold claims (a miss is a recall failure), while confusion is over matched claims only (a miss gives no predicted label, so it cannot enter a 2x2). Conflating them would double-count misses.
- Brier/ECE must be `*float64` and nil when no confidence is emitted, because forcing a number would imply the protocol is calibrated when it emitted no probabilities at all.

### What was tricky to build
- Sealing vs. digest presence, again (the same pattern as `assessment`): `Seal` must validate before the digest exists, so I split `validateReportBody` (all checks except digest presence) from `ValidateReport` (body + digest check), and `Seal` calls the body, computes the digest, sets it. This keeps a single source of truth for the range/format checks.
- `MatchByText` needs the predicted claim text, which lives in `report.Claims`, not in `ClaimAssessment`. I made `MatchByText(report)` return a `ClaimMatcher` that closes over a claim-ID-to-text map built from the report, so the matcher has access to the text without the caller threading it through.

### What warrants a second pair of eyes
- `Calibrate` matches gold claims to model verdicts per instance and silently skips gold claims whose instance has no report. Confirm that is the desired behavior (an unjudged instance counts as a recall miss for all its gold claims, which lowers recall) versus an explicit error.
- ECE uses equal-width binning and can hide within-bin structure; the design doc warns about this. A reliability-diagram helper would make the within-bin structure inspectable; deferred.

### What should be done in the future
- Add per-group `ByGroup` slicing in `Calibrate` (the field exists on the report but `Calibrate` does not populate it yet); requires a grouping key on gold records.
- Add a reliability-diagram helper alongside ECE.

### Code review instructions
- Start at `calibration/report.go` (`Calibrate`, `Seal`), then `calibration/confusion.go` (`ConfusionFromClaims`, the metric methods), then `calibration/recall.go` (`ExtractionRecall`, matchers).
- Run `GOWORK=off go test ./calibration/ -v`.

### Technical details
- Commands: `GOWORK=off go build ./calibration/`, `GOWORK=off go test ./calibration/`, `GOWORK=off go test ./...`.
- Brier: `(1/n) * sum (p - z)^2`. ECE: `sum_b (|I_b|/n) * |acc(I_b) - conf(I_b)|` with equal-width bins; a confidence of exactly 1.0 is clamped into the last bin.
- No new external dependencies; calibration imports only `assessment`, `spec`, and `internal/{canonicaljson,identifier}`.

## Step 12: Implement the audit package (Phase 9, part 1)

`audit` measures consistency (reliability) and sensitivity to features outside the construct (bias). It depends on `eval`, `assessment`, `spec`, and `judging` (panels use `judging.Judge`); it never calls providers directly. I built it after `calibration` because it sits above `calibration` in the dependency graph and needs the `Judge` interface to run probes.

### Prompt Context

**User prompt (verbatim):** (see Step 11)

**Commit (code):** c27ced9 — "feat(audit): add reliability probes, disagreement reports, and panels"

### What I did
- Wrote `audit/doc.go`, `probes.go`, `disagreement.go`, `reliability.go`, `panel.go`, `audit_test.go`.
- `Probe`/`ProbeKind` (repeat, evidence_order, candidate_order, prompt_paraphrase, format_transform, cross_judge). A probe MUST state `Invariants` so a found sensitivity can be localized. `NewProbeSet` content-addresses a probe set.
- `CompareReports` returns per-construct value/label differences and per-claim support-label differences as `Disagreement`s.
- `Reliability` runs a judge over a probe set, aggregates per-construct agreement and mean absolute delta, and seals the report. Never one "reliability score".
- `Panel` runs several judges over one instance, preserves every member report, and computes a pairwise claim-label agreement matrix. Never collapses reports.
- Extended `boundary_test.go` to `go list` `./audit` and `./calibration` so the forbidden-import guard covers the new packages.
- A `scriptedJudge` test double returns canned reports keyed by instance ID so a probe's base and variant produce different reports deterministically.

### Why
- Reliability is consistency, not correctness; it must be separate from caching (a cached wrong verdict is stable but unreliable) and from the judging path. A probe that states its invariants localizes a sensitivity instead of just reporting "unstable".
- Panels must preserve every member report because majority vote is an aggregation, not independent truth; collapsing loses the disagreement that reveals correlated errors (the "five judges, one error" counterexample).

### What worked
- `GOWORK=off go test ./audit/` green; full suite green; boundary test passes with audit+calibration included. `TestReliabilityDetectsDisagreement` pins faithfulness disagreement (0.5 vs 0.9 -> mean abs delta 0.4) and claim-label agreement 0 (entailed vs contradicted). `TestPanelPreservesAllReports` confirms two reports survive.

### What didn't work
- First `CompareReports` test failed: I had defined `Disagreement.BaseLabel`/`VariantLabel` as `assessment.SupportLabel`, but abstention dimensions carry free-form labels ("attempted"/"abstained") that `labelToSupport` discarded, so the disagreement was found but the label was empty. Fixed by changing the fields to `string` and removing `labelToSupport`; the fields now carry both support labels (claim disagreements) and free-form labels (dimension disagreements).
- A `makeProbe` test helper carried an invalid `_Reports: nil` field; removed the helper since tests construct probes inline.

### What I learned
- A disagreement struct that conflates claim labels and dimension labels must use `string`, not `SupportLabel`, because dimension labels are an open set. The construct-vs-claim distinction is already carried by which ID field is set (`ConstructID` vs `ClaimID`).
- Pairwise agreement over no shared claims must be 0, not 1, or the matrix over-reports agreement for judges that extracted disjoint claim sets.

### What was tricky to build
- `scriptedJudge.Evaluate` re-stamps `InstanceID`/`InstanceDigest` from the passed instance so the report matches the instance the reliability runner used, while the canned label/value comes from the script. This keeps the report internally consistent without the script needing to know each instance's digest.
- `dimDiffers` treats a nil-vs-non-nil value as a difference but compares present values with a 1e-9 tolerance so float formatting noise does not create false disagreements.

### What warrants a second pair of eyes
- `Panel.Evaluate` runs judges sequentially, not concurrently, to keep the test double deterministic. A production panel should run concurrently with errgroup; confirm the sequential version is acceptable for v0 or switch before any real use.
- `Reliability` skips probes whose judge evaluation errors out, returning the error; confirm fail-fast is preferred over partial reliability over the probes that succeeded.

### What should be done in the future
- Add bias probes (position, verbosity, style) as concrete probe builders, not just the `ProbeKind` enum.
- Add cross-judge agreement as a first-class metric distinct from reliability.

### Code review instructions
- Start at `audit/reliability.go` (`Reliability`, `RunProbe`, `ValidateReport`), `audit/disagreement.go` (`CompareReports`, `dimDiffers`), `audit/panel.go` (`Panel.Evaluate`, `pairwiseAgreement`).
- Run `GOWORK=off go test ./audit/ -v`.

### Technical details
- Commands: `GOWORK=off go build ./audit/`, `GOWORK=off go test ./audit/`, `GOWORK=off go test -run TestCorePackageBoundaries .`.
- audit imports: `eval`, `assessment`, `spec`, `judging`, `internal/canonicaljson`, `internal/identifier`. No provider SDKs; the boundary test confirms `judging` is the only execution dependency.

## Step 13: Implement the suite package (Phase 9, part 2)

`suite` combines multiple evaluators over one instance without collapsing their reports. It depends on `eval`, `assessment`, and `judging`, and adds `golang.org/x/sync/errgroup` for concurrent execution. I built it last because it sits beside `audit` and is the top of the dependency graph among the three new packages.

### Prompt Context

**User prompt (verbatim):** (see Step 11)

**Commit (code):** ca7a698 — "feat(suite): add acyclic evaluator suites with concurrent execution"

### What I did
- Wrote `suite/doc.go`, `suite.go`, `suite_test.go`.
- `Evaluator` interface: `Name`, `DependsOn`, `Evaluate(ctx, inst, Results)`. `Results.Report(name)` lets one evaluator read another's report only after that dependency ran.
- `Suite.Validate` (via `validateBody`): supported api_version, valid name, unique evaluator names, dependencies reference declared evaluators, and an acyclic graph (DFS with white/gray/black coloring reports the first cycle).
- `Suite.Run`: dispatches evaluators in dependency waves; independent evaluators run concurrently via `errgroup`, dependents run after their dependencies complete and see a snapshotted `Results`. Each report retains its own protocol identity.
- `NewSuite`: content-addressed; the digest is a function of the graph structure only (sorted dependency lists, no declaration order), so reordering evaluators does not change identity.
- `JudgeEvaluator` adapts a `judging.Judge` to a dependency-free `Evaluator`.
- Extended `boundary_test.go` to cover `./suite`.
- Lint fixes across the two earlier packages: dropped a named return and an unused `sortConfidences` in calibration; switched `MatchByText` to `strings.EqualFold`; removed an unused `makeProbe` test helper in audit; gofmt'd several files.

### Why
- A real evaluation needs more than one judge (extractor -> support, fact verifier, abstention judge, style judge); collapsing them loses per-evaluator protocol identity and disagreement. The suite runs them in dependency order and retains every report.
- The acyclic check is the invariant that makes `DependsOn` safe: a support judge may consume a claim extractor's output only when that edge is declared and cycle-free, otherwise the run would deadlock.

### What worked
- `GOWORK=off go test ./suite/` green; `TestRunConcurrent` confirms independent evaluators overlap (elapsed < 2*delay), proving errgroup actually runs them concurrently; `TestRunDependentEvaluatorSeesDependency` confirms a dependent sees its dependency's report; `TestSuiteDigestStable` confirms reordering evaluators does not change the digest.
- `golangci-lint run` reports 0 issues after the cleanup.

### What didn't work
- First `TestSuiteDigestStable` failed: the digest included an ordered `EvaluatorNames` slice, so reordering evaluators changed the digest. Fixed by dropping `EvaluatorNames` (derivable from the dependencies map keys) and sorting each dependency list, so the digest is a function of the graph structure only.
- `NewSuite` first used a placeholder-digest hack to pass `Validate` (which checked the digest before it was set). Fixed by splitting `validateBody` (no digest check) from `Validate` (body + digest check), the same pattern used in assessment and calibration.
- Lint flagged: a named return in `ConfidenceOutcomePairs` (nonamedreturns), an unused `sortConfidences`, `strings.ToLower == strings.ToLower` instead of `strings.EqualFold` (staticcheck SA6005), and an unused `makeProbe` test helper (unused). All fixed.

### What I learned
- A suite digest must be a function of the graph, not the declaration order, or the same evaluation described two ways gets two identities. Canonical JSON sorts map keys, so a digest over the dependencies map (with sorted dep slices) is order-independent.
- `errgroup` per wave plus `Wait()` between waves is the simple way to give dependents their dependencies' results: dispatch all ready evaluators concurrently, wait, then the next wave's readiness is recomputed.

### What was tricky to build
- The wave dispatch loop: each iteration finds evaluators whose remaining-dependencies set is empty, dispatches them concurrently, waits, then removes them from every other evaluator's remaining set. A snapshot of `Results` is taken at dispatch time so a concurrent evaluator sees a stable view. The `checkAcyclic` guard makes the "no ready evaluators but pending remain" branch unreachable, but it is kept as a deadlock guard.
- `errgroup.WithContext` cancels remaining work on first error, so an evaluator failure fails the whole run fast; `TestRunErrorPropagates` pins that.

### What warrants a second pair of eyes
- The wave model runs each ready-set concurrently but waits between waves, so a long chain of dependents is sequential across waves. For deep chains, a per-evaluator future/condition would overlap more; confirm the wave model is acceptable for v0.
- `Suite.Evaluators` is `[]Evaluator` (not serializable), and the digest carries only names + deps. Confirm that pinning a suite by digest is sufficient for reproducibility (the evaluators themselves are not content-addressed, only their graph).

### What should be done in the future
- Add a per-evaluator future/condition model for deeper concurrency if real suites have long dependency chains.
- Content-address individual evaluators if suite reproducibility needs to pin the exact judges, not just the graph.

### Code review instructions
- Start at `suite/suite.go` (`Validate`/`validateBody`, `checkAcyclic`, `Run`, `NewSuite`), then `suite/suite_test.go`.
- Run `GOWORK=off go test ./suite/ -v` and `GOWORK=off go test -run TestCorePackageBoundaries .`.

### Technical details
- Commands: `GOWORK=off go get golang.org/x/sync`, `GOWORK=off go mod tidy`, `GOWORK=off go test ./suite/`, `GOWORK=off golangci-lint run`.
- New dependency: `golang.org/x/sync/errgroup` (the AGENT.md convention for goroutines; the design doc specifies errgroup). It is not on the boundary-test forbidden list.
- suite imports: `eval`, `assessment`, `judging`, `internal/canonicaljson`, `internal/identifier`, `golang.org/x/sync/errgroup`. No provider SDKs.

## Step 14: Full verification and ticket bookkeeping

### Prompt Context

**User prompt (verbatim):** (see Step 11)

### What I did
- Ran `GOWORK=off go test -count=1 ./...` (15 packages, all green including the boundary test over spec/eval/protocol/assessment/judging/audit/calibration/suite/internal).
- Ran `GOWORK=off golangci-lint run` -> 0 issues.
- Confirmed `go vet ./...` and `gofmt -l` clean.

### What worked
- All three new packages pass tests and lint; the boundary test enforces the stdlib-only-core invariant across all eight core packages.

### What should be done in the future
- Update GLOSSARY.md and the developer reference to mark audit/calibration/suite as implemented (currently documented as "future").
- Add the `audit` and `calibration` packages to the README package map and the user-guide invariants.
- Add fuzz tests for the new metric functions (Brier/ECE, extraction recall, disagreement comparison) per the design's Phase 11.2.
