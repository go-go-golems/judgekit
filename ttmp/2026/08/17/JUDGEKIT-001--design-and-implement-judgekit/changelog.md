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

