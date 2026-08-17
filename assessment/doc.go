// Package assessment represents the structured observation an evaluator
// produces: extracted claims, per-claim support verdicts against evidence,
// per-construct dimension results, and a sealed, content-addressed report.
//
// assessment is the convergence point of the core value packages: it imports
// spec (for ConstructID) and eval (for raw artifacts) but not judging, so a
// report can be validated without running an evaluator.
//
// Non-goals:
//
//   - assessment does not decide deployments or promotions.
//   - assessment does not call providers or render prompts.
//
// Invariants:
//
//   - Support is three-way (entailed, contradicted, insufficient) so
//     contradiction is never conflated with absent evidence.
//   - Entailed and contradicted verdicts require cited evidence; an empty
//     evidence list is only valid for insufficient.
//   - Every claim has exactly one claim result; claim results reference real
//     claims; dimensions have unique construct IDs.
//   - Evidence cross-references are checked against an allowed set supplied by
//     the judging layer, so a verdict cannot invent evidence.
//   - Reports are sealed: their digest is a function of content, computed
//     once, after which the report is treated as immutable.
package assessment
