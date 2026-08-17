// Package calibration measures whether a judge's stated confidence agrees with
// empirical frequency, and whether the judge extracts the claims a human would
// extract.
//
// calibration is the link between judgekit reports and human or objective
// labels. It owns gold records (which retain reviewer identity so inter-rater
// agreement can be measured), extraction recall (a judge can look accurate by
// failing to extract difficult claims), confusion matrices, sensitivity and
// specificity, the false-support rate, the Brier score, and expected
// calibration error.
//
// Non-goals:
//
//   - calibration does not run judges or call providers. It consumes
//     assessment.Report values and GoldSet values.
//   - calibration does not decide deployments; it produces evidence about a
//     judge's calibration, not a verdict.
//
// Invariants:
//
//   - Gold records retain ReviewerIDs. Adjudication (GoldClaim.Adjudicated)
//     does not erase original disagreement; the reviewer list is kept so
//     agreement can be measured even after adjudication.
//   - Extraction recall compares model-extracted claims against
//     human-enumerated claims. A judge that extracts fewer claims can appear
//     accurate on the claims it does extract; recall exposes that.
//   - Brier score and ECE apply only when the protocol emits confidence
//     probabilities. A 1-5 ordinal score is not a probability. These metrics are
//     *float64 in the report and are nil when no confidence is available.
//
// See GLOSSARY.md for the measurement-theory definitions of calibration,
// sensitivity, specificity, and reliability.
package calibration
