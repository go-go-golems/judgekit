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
