// Package suite combines multiple evaluators over one instance without
// collapsing their reports into a single score.
//
// A real evaluation often needs more than one judge: a claim extractor feeding a
// support judge, a required-fact verifier, a citation resolver, an abstention
// judge, a style judge. Collapsing them into one output loses the per-evaluator
// protocol identity and the per-evaluator disagreements. A suite runs them in
// dependency order, lets one evaluator consume another's results, and retains
// every report keyed by evaluator name.
//
// Non-goals:
//
//   - suite does not define constructs, prompts, or model identity; each
//     evaluator carries its own protocol.
//   - suite does not decide deployments; it produces a set of reports.
//
// Invariants:
//
//   - The dependency graph must be acyclic. A support judge may consume a claim
//     extractor's output only when that dependency is declared; a cycle is
//     rejected before any evaluator runs.
//   - Independent evaluators run concurrently. Each report retains its own
//     protocol identity; the suite does not merge them.
//   - A suite digest identifies the evaluator graph so a set of reports can be
//     pinned to the suite that produced them.
package suite
