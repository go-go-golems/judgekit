package judging

import (
	"context"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/protocol"
)

// GenerationRequest is a provider-neutral generation request. Step names the
// judge stage (for example "extract" or "support") so a generator can route or
// observe by stage. The prompt is fully rendered by the application.
type GenerationRequest struct {
	Prompt     string
	MediaType  string
	ProtocolID string
	Step       string
}

// GenerationResult is a provider-neutral generation result. Model records the
// identity that actually served the request so a report can prove which model
// produced its raw text.
type GenerationResult struct {
	Text         string
	InputTokens  int
	OutputTokens int
	Model        protocol.ModelIdentity
	Duration     time.Duration
}

// Generator runs one rendered prompt on the evaluator model and returns its
// text. Implementations translate provider configuration into this stable
// interface; core never imports a provider SDK.
type Generator interface {
	Generate(ctx context.Context, req GenerationRequest) (GenerationResult, error)
}

// Judge evaluates one instance and returns a sealed report.
type Judge interface {
	Evaluate(ctx context.Context, inst eval.Instance) (assessment.Report, error)
}

// Critique is diagnostic feedback on an instance/report pair, intended to
// guide a revision rather than rank candidates.
type Critique struct {
	InstanceID string
	Summary    string
	Actions    []string
	Digest     string
}

// Critic produces diagnostic feedback. It need not rank or score.
type Critic interface {
	Critique(ctx context.Context, inst eval.Instance, report assessment.Report) (Critique, error)
}

// VerificationRequest asks a verifier to check one proposition against an
// instance. Verifiers are usually narrower than a general judge and often
// have access to tools or formal evidence.
type VerificationRequest struct {
	Instance    eval.Instance
	Proposition string
}

// VerificationResult is a narrow verdict on one proposition.
type VerificationResult struct {
	Holds      bool
	Confidence *float64
	Reason     string
}

// Verifier checks a proposition, intermediate step, constraint, or result.
type Verifier interface {
	Verify(ctx context.Context, req VerificationRequest) (VerificationResult, error)
}
