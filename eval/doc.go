// Package eval defines one concrete item being judged.
//
// An evaluation instance is the unit of evaluation: the task input, the
// candidate artifact, the evidence set the evaluator may inspect, an optional
// reference, optional required facts, and metadata. Everything an evaluator
// observes is captured here so that two runs of the same protocol over the
// same instance produce comparable assessments.
//
// Non-goals:
//
//   - eval does not define constructs or measurement contracts (see spec).
//   - eval does not define model identity or prompts (see protocol).
//   - eval does not run evaluators (see judging).
//
// Invariants:
//
//   - Artifact content is content-addressed: a text artifact carries the
//     SHA-256 of its bytes; a URI artifact carries a caller-provided digest.
//   - Evidence item IDs are unique within a set and valid identifiers.
//   - RequiredFact.EvidenceIDs reference real evidence item IDs at the
//     instance level (cross-reference validation fails closed).
//   - Instance, EvidenceSet, and Artifact digests make inputs content
//     addressable so reports and caches can pin exact inputs.
package eval
