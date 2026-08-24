// Package judging runs evaluators over instances and produces assessments.
//
// judging is the only core package that depends on all the value packages
// (spec, eval, protocol, assessment). It exposes provider-neutral interfaces
// so that core never imports a provider SDK: applications implement
// Generator to adapt their model runtime, and judgekit orchestrates the rest.
//
// Non-goals:
//
//   - judging does not own prompts or product rubrics; applications supply
//     prompt rendering via ClaimProtocol and similar.
//   - judging does not decide deployments or promotions.
//   - judging does not store hidden chain-of-thought.
//
// Invariants:
//
//   - A Judge produces a sealed assessment.Report whose evidence references
//     resolve to the instance's evidence set.
//   - The two-stage claim judge hides evidence from the extractor and gives
//     evidence to the support judge; the judge never invents evidence.
//   - Only structural failures are repaired by default; semantic failures
//     fail closed.
package judging
