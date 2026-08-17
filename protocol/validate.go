package protocol

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/judgekit/internal/identifier"
)

// ValidateModel returns nil when m is a well-formed model identity.
func ValidateModel(m *ModelIdentity) error {
	if strings.TrimSpace(m.Provider) == "" {
		return fmt.Errorf("model: provider is required")
	}
	if strings.TrimSpace(m.Model) == "" {
		return fmt.Errorf("model: model is required")
	}
	for k, v := range m.Settings {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("model settings: blank key")
		}
		_ = v
	}
	return nil
}

// ValidateDecoding returns nil when d is a well-formed decoding policy.
func ValidateDecoding(d *DecodingPolicy) error {
	if d.MaxTokens <= 0 {
		return fmt.Errorf("decoding: max_tokens must be positive, got %d", d.MaxTokens)
	}
	if d.Temperature != nil {
		t := *d.Temperature
		if math.IsNaN(t) || math.IsInf(t, 0) || t < 0 || t > 2 {
			return fmt.Errorf("decoding: temperature %g must be in [0,2]", t)
		}
	}
	if d.TopP != nil {
		p := *d.TopP
		if math.IsNaN(p) || math.IsInf(p, 0) || p <= 0 || p > 1 {
			return fmt.Errorf("decoding: top_p %g must be in (0,1]", p)
		}
	}
	return nil
}

// ValidateRetry returns nil when r is a well-formed retry policy.
func ValidateRetry(r *RetryPolicy) error {
	if r.MaximumAttempts < 1 {
		return fmt.Errorf("retry: maximum_attempts must be >= 1, got %d", r.MaximumAttempts)
	}
	seen := make(map[string]struct{}, len(r.RepairKinds))
	for _, k := range r.RepairKinds {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("retry: blank repair kind")
		}
		if _, dup := seen[k]; dup {
			return fmt.Errorf("retry: duplicate repair kind %q", k)
		}
		seen[k] = struct{}{}
	}
	return nil
}

// Validate returns nil when p is a well-formed protocol.
func Validate(p *Protocol) error {
	if p.APIVersion != ProtocolAPIVersion {
		return fmt.Errorf("protocol %q: api_version %q is not supported (want %s)", p.Name, p.APIVersion, ProtocolAPIVersion)
	}
	if err := identifier.Validate(p.Name); err != nil {
		return fmt.Errorf("protocol name: %w", err)
	}
	if !strings.HasPrefix(p.MeasurementDigest, "sha256:") {
		return fmt.Errorf("protocol %q: measurement_digest must be a sha256: digest", p.Name)
	}
	if err := ValidateModel(&p.Model); err != nil {
		return fmt.Errorf("protocol %q: %w", p.Name, err)
	}
	if len(p.PromptDigests) == 0 {
		return fmt.Errorf("protocol %q: at least one prompt digest is required", p.Name)
	}
	for step, d := range p.PromptDigests {
		if strings.TrimSpace(step) == "" {
			return fmt.Errorf("protocol %q: prompt digest has a blank step name", p.Name)
		}
		if !strings.HasPrefix(d, "sha256:") {
			return fmt.Errorf("protocol %q: prompt digest for step %q must be a sha256: digest", p.Name, step)
		}
	}
	if err := ValidateDecoding(&p.Decoding); err != nil {
		return fmt.Errorf("protocol %q: %w", p.Name, err)
	}
	if !validEvidenceOrders[p.EvidenceOrder] {
		return fmt.Errorf("protocol %q: evidence_order %q is not recognized", p.Name, p.EvidenceOrder)
	}
	if p.EvidenceOrder == EvidenceOrderShuffled && p.Decoding.Seed == nil {
		return fmt.Errorf("protocol %q: evidence_order shuffled requires a decoding seed", p.Name)
	}
	if strings.TrimSpace(p.ParserVersion) == "" {
		return fmt.Errorf("protocol %q: parser_version is required", p.Name)
	}
	if strings.TrimSpace(p.AggregatorVersion) == "" {
		return fmt.Errorf("protocol %q: aggregator_version is required", p.Name)
	}
	if err := ValidateRetry(&p.Retry); err != nil {
		return fmt.Errorf("protocol %q: %w", p.Name, err)
	}
	return nil
}
