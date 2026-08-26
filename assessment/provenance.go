package assessment

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/judgekit/protocol"
)

// PromptExecution records one rendered generation attempt. Template identities
// live on RunProvenance; RenderedPromptDigest identifies the exact text used
// for this instance and attempt. ObservedModel is provider-reported research
// attribution, not a cryptographic attestation.
type PromptExecution struct {
	Step                 string                 `json:"step" yaml:"step"`
	Attempt              int                    `json:"attempt" yaml:"attempt"`
	RenderedPromptDigest string                 `json:"rendered_prompt_digest" yaml:"rendered_prompt_digest"`
	ObservedModel        protocol.ModelIdentity `json:"observed_model" yaml:"observed_model"`
	CacheHit             bool                   `json:"cache_hit" yaml:"cache_hit"`
	InputTokens          int                    `json:"input_tokens,omitempty" yaml:"input_tokens,omitempty"`
	OutputTokens         int                    `json:"output_tokens,omitempty" yaml:"output_tokens,omitempty"`
	DurationNanos        int64                  `json:"duration_nanos,omitempty" yaml:"duration_nanos,omitempty"`
}

// RunProvenance binds a report to the contract, protocol, current instance,
// prompt templates, rendered prompts, expected/observed model, and run-scoped
// cache behavior that produced it.
type RunProvenance struct {
	ContractDigest        string                 `json:"contract_digest" yaml:"contract_digest"`
	ProtocolDigest        string                 `json:"protocol_digest" yaml:"protocol_digest"`
	InstanceDigest        string                 `json:"instance_digest" yaml:"instance_digest"`
	PromptTemplateDigests map[string]string      `json:"prompt_template_digests" yaml:"prompt_template_digests"`
	ExpectedModel         protocol.ModelIdentity `json:"expected_model" yaml:"expected_model"`
	CacheMode             string                 `json:"cache_mode" yaml:"cache_mode"`
	Generations           []PromptExecution      `json:"generations" yaml:"generations"`
}

func validateRunProvenance(p *RunProvenance, reportProtocol, reportInstance string) error {
	for name, digest := range map[string]string{
		"contract_digest": p.ContractDigest,
		"protocol_digest": p.ProtocolDigest,
		"instance_digest": p.InstanceDigest,
	} {
		if !strings.HasPrefix(digest, "sha256:") {
			return fmt.Errorf("report provenance: %s must be a sha256: digest", name)
		}
	}
	if p.ProtocolDigest != reportProtocol {
		return fmt.Errorf("report provenance: protocol digest %q does not match report %q", p.ProtocolDigest, reportProtocol)
	}
	if p.InstanceDigest != reportInstance {
		return fmt.Errorf("report provenance: instance digest %q does not match report %q", p.InstanceDigest, reportInstance)
	}
	if strings.TrimSpace(p.ExpectedModel.Provider) == "" || strings.TrimSpace(p.ExpectedModel.Model) == "" {
		return fmt.Errorf("report provenance: expected model provider and model are required")
	}
	if p.CacheMode != "use" && p.CacheMode != "bypass" {
		return fmt.Errorf("report provenance: cache mode %q is not recognized", p.CacheMode)
	}
	if len(p.PromptTemplateDigests) == 0 {
		return fmt.Errorf("report provenance: prompt template digests are required")
	}
	for step, digest := range p.PromptTemplateDigests {
		if strings.TrimSpace(step) == "" || !strings.HasPrefix(digest, "sha256:") {
			return fmt.Errorf("report provenance: prompt template %q must have a sha256: digest", step)
		}
	}
	if len(p.Generations) == 0 {
		return fmt.Errorf("report provenance: at least one generation is required")
	}
	for index := range p.Generations {
		generation := &p.Generations[index]
		if strings.TrimSpace(generation.Step) == "" {
			return fmt.Errorf("report provenance: generation %d step is required", index)
		}
		if _, ok := p.PromptTemplateDigests[generation.Step]; !ok {
			return fmt.Errorf("report provenance: generation %d references unknown prompt step %q", index, generation.Step)
		}
		if generation.Attempt < 1 {
			return fmt.Errorf("report provenance: generation %d attempt must be positive", index)
		}
		if !strings.HasPrefix(generation.RenderedPromptDigest, "sha256:") {
			return fmt.Errorf("report provenance: generation %d rendered prompt must be a sha256: digest", index)
		}
		if generation.InputTokens < 0 || generation.OutputTokens < 0 || generation.DurationNanos < 0 {
			return fmt.Errorf("report provenance: generation %d usage and duration must be non-negative", index)
		}
		if err := protocol.ValidateObservedModel(p.ExpectedModel, generation.ObservedModel); err != nil {
			return fmt.Errorf("report provenance: generation %d: %w", index, err)
		}
	}
	return nil
}
