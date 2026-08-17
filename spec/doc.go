// Package spec defines what an evaluator measures.
//
// spec owns the first link of judgekit's inference chain:
//
//	construct -> measurement contract -> protocol -> instance -> assessment
//
// A Construct is the abstract property an evaluation intends to measure
// (faithfulness, completeness, appropriate abstention). A MeasurementContract
// is its operational definition: the constructs, allowed evidence kinds,
// labels, aggregations, exclusions, and empty-case behavior.
//
// Non-goals:
//
//   - spec does not run evaluators or call providers (see package judging).
//   - spec does not define prompts, model identity, or decoding (see package
//     protocol).
//   - spec does not decide deployments or promotions.
//
// Invariants:
//
//   - Construct and MeasurementContract values are validated before they are
//     used; invalid values fail closed.
//   - Semantic identity is the SHA-256 of the canonical JSON of a validated
//     contract, so harmless YAML key ordering does not change identity.
//   - Byte identity is the SHA-256 of the raw source bytes, so the exact
//     reviewed file is provable.
//
// See the GLOSSARY.md at the repository root for the measurement-theory
// definitions of construct, measurement, proxy, reliability, validity, and
// calibration that this package operationalizes.
package spec
