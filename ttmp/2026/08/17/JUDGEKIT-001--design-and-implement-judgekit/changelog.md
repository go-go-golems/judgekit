# Changelog

## 2026-08-17

- Initial workspace created


## 2026-08-17

Initialized the repository-local ticket and authored the judgekit architecture and implementation guide with package APIs, invariants, integration boundaries, ten implementation phases, tests, security controls, and decision records.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/go.mod — Template baseline analyzed
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/01-judgekit-architecture-and-implementation-guide.md — Primary design and implementation guide
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Investigation and initialization record


## 2026-08-17

Validated and uploaded the Judgekit Architecture and Implementation Guide bundle to reMarkable.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/01-judgekit-architecture-and-implementation-guide.md — Completed and delivered architecture guide
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Recorded initialization, validation, upload failure, and successful delivery


## 2026-08-17

Implemented the judgekit core: normalized the repo, added internal canonical primitives, and built spec/eval/protocol/assessment/judging with strict validation, dual digests, three-way support, a provider-neutral two-stage claim judge, and a boundary test. Added GLOSSARY.md (textbook-grounded), README, a user guide, a getting-started Glazed help entry, a developer reference, a thin help-host CLI (cmd/judgekit), and a runnable claim-judge example. All tests and the boundary test pass.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/GLOSSARY.md — Construct glossary
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/cmd/judgekit/main.go — Help-host CLI
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/judging/claimjudge.go — Two-stage claim judge


## 2026-08-17

Implemented the remaining core packages: calibration (gold records, extraction recall, confusion, Brier, ECE), audit (reliability probes, disagreement reports, panels), and suite (acyclic evaluator graph, concurrent execution). All pass tests and lint; boundary test extended to cover all eight core packages.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/audit/reliability.go — Reliability aggregation
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/calibration/report.go — Calibrate entry point
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/suite/suite.go — Acyclic evaluator suites


## 2026-08-24

Added three intern-ready guides separating PR 2 stabilization, post-merge lightweight research guarantees, and the optional fully hardened architecture.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/02-pr-2-code-review-stabilization-guide.md — Maps review findings to fixes and tests
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/03-lightweight-research-guarantees-after-merge.md — Defines recommended research-stage guarantees
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/04-fully-hardened-judgekit-architecture.md — Preserves a deferred cross-trust hardening design
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Records analysis, decisions, and delivery work


## 2026-08-24

Validated and uploaded the three-guide JUDGEKIT-001 bundle to reMarkable at /ai/2026/08/24/JUDGEKIT-001.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/02-pr-2-code-review-stabilization-guide.md — Uploaded immediate stabilization guide
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/03-lightweight-research-guarantees-after-merge.md — Uploaded recommended post-merge guide
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/04-fully-hardened-judgekit-architecture.md — Uploaded deferred hardening reference
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Recorded validation and delivery evidence


## 2026-08-24

Implemented PR 2 stabilization across judging, calibration, audit, suite, CI, tests, and docs (commits 4126d73, 52869e3, 7a69477, 3e7e258, 2739bac).

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/audit/panel.go — Stable concurrent panel member identity
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/audit/reliability.go — Fresh calls and union-based reliability
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/calibration/report.go — Explicit target probability and protocol-homogeneous calibration
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/judging/claimjudge.go — Contract, protocol, prompt, evidence, direct-result, and cache-mode enforcement
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/suite/suite.go — Fresh context per dependency wave
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Implementation evidence and validation results


## 2026-08-24

Enabled dependency review and restored github.com/buger/jsonparser v1.2.0 after CI exposed the vulnerable 1.1.1 transitive downgrade (commit 6f8acb9).

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/go.mod — Pins the patched transitive dependency
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/go.sum — Records patched module checksums
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Records CI failure and correction


## 2026-08-24

PR 2 stabilization complete: all 11 review threads resolved and all 8 GitHub checks pass.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Final PR and CI completion evidence


## 2026-08-24

Added the PR 2 second-review addendum covering restricted extraction input, required field presence, observed model binding, exact integer canonicalization, reliability protocol binding, and durable snapshot verification.

### Related Files

- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/01-judgekit-architecture-and-implementation-guide.md — Corrected foundational extraction and canonical identity requirements
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/02-pr-2-code-review-stabilization-guide.md — Added findings 12-18 and expanded merge gate
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/design-doc/03-lightweight-research-guarantees-after-merge.md — Clarified lightweight attribution and snapshot boundaries
- /home/manuel/workspaces/2026-08-12/deploy-dev-indexer/judgekit/ttmp/2026/08/17/JUDGEKIT-001--design-and-implement-judgekit/reference/01-investigation-diary.md — Recorded second-review design reasoning

