package judging

import (
	"context"
	"fmt"

	"github.com/go-go-golems/judgekit/protocol"
)

// FakeGenerator returns canned responses keyed by step and records every
// request. It lets the claim judge and examples run with no provider
// credentials, and lets tests assert prompt construction (for example, that
// the extractor never received evidence).
type FakeGenerator struct {
	Responses map[string]string
	Calls     []GenerationRequest
	Model     protocol.ModelIdentity
	Err       error
}

// Generate implements Generator by returning the canned response for req.Step.
func (f *FakeGenerator) Generate(_ context.Context, req GenerationRequest) (GenerationResult, error) {
	f.Calls = append(f.Calls, req)
	if f.Err != nil {
		return GenerationResult{}, f.Err
	}
	raw, ok := f.Responses[req.Step]
	if !ok {
		return GenerationResult{}, fmt.Errorf("fake generator: no response for step %q", req.Step)
	}
	m := f.Model
	if m.Model == "" {
		m = protocol.ModelIdentity{Provider: "fake", Model: "fake-1"}
	}
	return GenerationResult{Text: raw, Model: m}, nil
}

// Reset clears recorded calls.
func (f *FakeGenerator) Reset() { f.Calls = nil }
