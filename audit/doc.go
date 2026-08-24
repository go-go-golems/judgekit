// Package audit measures whether a judge is consistent (reliability) and
// whether it responds to features outside the intended construct (bias).
//
// audit runs a judge over a base instance and a variant that changes only
// something that should not affect the construct, then compares the two
// reports. Reliability is consistency, not correctness: a judge can be wrong
// but reliable, or right on average but unreliable.
//
// Non-goals:
//
//   - audit does not call providers directly. It runs a judging.Judge, which
//     is the provider-neutral interface; it never imports a provider SDK.
//   - audit does not decide deployments. A reliability report is evidence
//     about a judge's stability, not a verdict.
//
// Invariants:
//
//   - A Probe states what remains semantically invariant (Invariants). Removing
//     headings, reversing pair order, or shuffling evidence should not change
//     the intended construct; a probe that finds a change localizes the
//     sensitivity.
//   - A ReliabilityReport carries per-construct and per-probe breakdowns, never
//     one "reliability score". A single number hides which construct and which
//     perturbation are unstable.
//   - Panel output preserves every member report and disagreement. Majority
//     vote is an aggregation, not independent truth; judgekit can compute
//     pairwise agreement but cannot infer error independence without external
//     labels.
//
// See GLOSSARY.md for the measurement-theory definitions of reliability, bias,
// and the nuisance-variable concept.
package audit
