// Package protocol identifies how an evaluator measures, as opposed to what
// it measures (spec) or what it observes (eval).
//
// A protocol is the complete reproducible instrument configuration: the
// measurement-contract digest it targets, the model identity, the prompt
// digests, the decoding policy, the evidence ordering, the parser and
// aggregator versions, and the retry policy. A model name alone is never a
// protocol identity: a one-token prompt change, a different decoding seed, or
// a new parser version is a different protocol with a different digest.
//
// Non-goals:
//
//   - protocol does not own prompts themselves (applications own prompt text;
//     protocol stores only their digests).
//   - protocol does not run the evaluator (see judging).
//
// Invariants:
//
//   - The protocol digest covers every semantically relevant field, so any
//     change to model, prompts, decoding, ordering, parser, aggregator, or
//     retry changes the digest.
//   - MeasurementDigest pins the contract the protocol targets; reports carry
//     the protocol digest and thereby the contract indirectly.
package protocol
