package audit

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/internal/canonicaljson"
	"github.com/go-go-golems/judgekit/judging"
	"github.com/go-go-golems/judgekit/spec"
)

// ReportAPIVersion is the only reliability report API version judgekit accepts.
const ReportAPIVersion = "judgekit.reliability/v1"

// ReliabilityReport summarizes agreement between base and variant reports
// across a probe set. It carries per-construct breakdowns, never one
// "reliability score", because a single number hides which construct and which
// perturbation are unstable.
type ReliabilityReport struct {
	APIVersion          string                       `json:"api_version" yaml:"api_version"`
	ProtocolDigest      string                       `json:"protocol_digest" yaml:"protocol_digest"`
	ProbeSetDigest      string                       `json:"probe_set_digest" yaml:"probe_set_digest"`
	TotalPairs          int                          `json:"total_pairs" yaml:"total_pairs"`
	ClaimLabelAgreement float64                      `json:"claim_label_agreement" yaml:"claim_label_agreement"`
	DimensionAgreement  map[spec.ConstructID]float64 `json:"dimension_agreement" yaml:"dimension_agreement"`
	MeanAbsoluteDelta   map[spec.ConstructID]float64 `json:"mean_absolute_delta" yaml:"mean_absolute_delta"`
	Disagreements       []Disagreement               `json:"disagreements" yaml:"disagreements"`
	Digest              string                       `json:"digest" yaml:"digest"`
}

// RunProbe runs judge over the probe's base and variant instances and returns
// the two reports. It does not compare them; call CompareReports for that.
func RunProbe(ctx context.Context, judge judging.Judge, probe Probe) (assessment.Report, assessment.Report, error) {
	if err := ValidateProbe(&probe); err != nil {
		return assessment.Report{}, assessment.Report{}, fmt.Errorf("run probe: %w", err)
	}
	if judge == nil {
		return assessment.Report{}, assessment.Report{}, fmt.Errorf("run probe: judge is nil")
	}
	evaluate := judge.Evaluate
	if configurable, ok := judge.(judging.ConfigurableJudge); ok {
		evaluate = func(ctx context.Context, inst eval.Instance) (assessment.Report, error) {
			return configurable.EvaluateWithOptions(ctx, inst, judging.EvaluationOptions{CacheMode: judging.CacheBypass})
		}
	}
	base, err := evaluate(ctx, probe.BaseInstance)
	if err != nil {
		return assessment.Report{}, assessment.Report{}, fmt.Errorf("run probe %q: base: %w", probe.ID, err)
	}
	variant, err := evaluate(ctx, probe.VariantInstance)
	if err != nil {
		return assessment.Report{}, assessment.Report{}, fmt.Errorf("run probe %q: variant: %w", probe.ID, err)
	}
	return base, variant, nil
}

// Reliability runs judge over every probe in the set, compares base and variant
// reports, and aggregates agreement. protocolDigest is the digest of the
// protocol the judge was run under.
func Reliability(ctx context.Context, judge judging.Judge, set ProbeSet, protocolDigest string) (ReliabilityReport, error) {
	if err := ValidateProbeSet(&set); err != nil {
		return ReliabilityReport{}, fmt.Errorf("reliability: %w", err)
	}
	if judge == nil {
		return ReliabilityReport{}, fmt.Errorf("reliability: judge is nil")
	}
	if !strings.HasPrefix(protocolDigest, "sha256:") {
		return ReliabilityReport{}, fmt.Errorf("reliability: protocol_digest must be a sha256: digest")
	}

	report := ReliabilityReport{
		APIVersion:         ReportAPIVersion,
		ProtocolDigest:     protocolDigest,
		ProbeSetDigest:     set.Digest,
		DimensionAgreement: map[spec.ConstructID]float64{},
		MeanAbsoluteDelta:  map[spec.ConstructID]float64{},
	}

	claimMatches := 0
	claimTotal := 0
	dimMatches := map[spec.ConstructID]int{}
	dimTotal := map[spec.ConstructID]int{}
	dimDeltaSum := map[spec.ConstructID]float64{}
	dimDeltaCount := map[spec.ConstructID]int{}

	for i := range set.Probes {
		probe := set.Probes[i]
		base, variant, err := RunProbe(ctx, judge, probe)
		if err != nil {
			return ReliabilityReport{}, err
		}
		report.TotalPairs++
		disags := CompareReports(probe.ID, probe.Kind, base, variant)
		report.Disagreements = append(report.Disagreements, disags...)

		// Claim-label agreement over the union. Missing claims count in the
		// denominator and cannot be agreements.
		bc := indexClaims(base)
		vc := indexClaims(variant)
		for _, cid := range unionKeys(bc, vc) {
			b, baseOK := bc[cid]
			v, variantOK := vc[cid]
			claimTotal++
			if baseOK && variantOK && b.Label == v.Label {
				claimMatches++
			}
		}

		// Per-construct agreement also uses the union. Numeric delta is only
		// defined when both sides emitted numeric values.
		bd := indexDims(base)
		vd := indexDims(variant)
		for _, cid := range unionKeys(bd, vd) {
			b, baseOK := bd[cid]
			v, variantOK := vd[cid]
			cidID := spec.ConstructID(cid)
			dimTotal[cidID]++
			if baseOK && variantOK && !dimDiffers(b, v) {
				dimMatches[cidID]++
			}
			if baseOK && variantOK && b.Value != nil && v.Value != nil {
				dimDeltaSum[cidID] += math.Abs(*b.Value - *v.Value)
				dimDeltaCount[cidID]++
			}
		}
	}

	if claimTotal > 0 {
		report.ClaimLabelAgreement = float64(claimMatches) / float64(claimTotal)
	}
	for cid, total := range dimTotal {
		if total > 0 {
			report.DimensionAgreement[cid] = float64(dimMatches[cid]) / float64(total)
		}
		if dimDeltaCount[cid] > 0 {
			report.MeanAbsoluteDelta[cid] = dimDeltaSum[cid] / float64(dimDeltaCount[cid])
		}
	}

	if err := sealReliability(&report); err != nil {
		return ReliabilityReport{}, fmt.Errorf("reliability: %w", err)
	}
	return report, nil
}

// ValidateReport returns nil when r is a well-formed, sealed reliability report.
func ValidateReport(r *ReliabilityReport) error {
	if r.APIVersion != ReportAPIVersion {
		return fmt.Errorf("reliability report: api_version %q is not supported (want %s)", r.APIVersion, ReportAPIVersion)
	}
	if !strings.HasPrefix(r.ProtocolDigest, "sha256:") {
		return fmt.Errorf("reliability report: protocol_digest must be a sha256: digest")
	}
	if !strings.HasPrefix(r.ProbeSetDigest, "sha256:") {
		return fmt.Errorf("reliability report: probe_set_digest must be a sha256: digest")
	}
	if r.ClaimLabelAgreement < 0 || r.ClaimLabelAgreement > 1 {
		return fmt.Errorf("reliability report: claim_label_agreement %g must be in [0,1]", r.ClaimLabelAgreement)
	}
	for cid, a := range r.DimensionAgreement {
		if a < 0 || a > 1 {
			return fmt.Errorf("reliability report: dimension_agreement[%s] %g must be in [0,1]", cid, a)
		}
	}
	for cid, d := range r.MeanAbsoluteDelta {
		if d < 0 {
			return fmt.Errorf("reliability report: mean_absolute_delta[%s] %g must be non-negative", cid, d)
		}
	}
	if !strings.HasPrefix(r.Digest, "sha256:") {
		return fmt.Errorf("reliability report: digest must be a sha256: digest")
	}
	return nil
}

// sealReliability validates the body, computes the digest, and sets r.Digest.
func sealReliability(r *ReliabilityReport) error {
	digest, err := canonicaljson.Sum(reliabilityDigestInput{
		APIVersion:          r.APIVersion,
		ProtocolDigest:      r.ProtocolDigest,
		ProbeSetDigest:      r.ProbeSetDigest,
		TotalPairs:          r.TotalPairs,
		ClaimLabelAgreement: r.ClaimLabelAgreement,
		DimensionAgreement:  r.DimensionAgreement,
		MeanAbsoluteDelta:   r.MeanAbsoluteDelta,
		Disagreements:       r.Disagreements,
	})
	if err != nil {
		return fmt.Errorf("seal reliability report: %w", err)
	}
	r.Digest = digest
	return nil
}

// reliabilityDigestInput is ReliabilityReport without the Digest field.
type reliabilityDigestInput struct {
	APIVersion          string                       `json:"api_version"`
	ProtocolDigest      string                       `json:"protocol_digest"`
	ProbeSetDigest      string                       `json:"probe_set_digest"`
	TotalPairs          int                          `json:"total_pairs"`
	ClaimLabelAgreement float64                      `json:"claim_label_agreement"`
	DimensionAgreement  map[spec.ConstructID]float64 `json:"dimension_agreement"`
	MeanAbsoluteDelta   map[spec.ConstructID]float64 `json:"mean_absolute_delta"`
	Disagreements       []Disagreement               `json:"disagreements"`
}
