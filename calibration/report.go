package calibration

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/internal/canonicaljson"
)

// ReportAPIVersion is the only calibration report API version judgekit accepts.
const ReportAPIVersion = "judgekit.calibration/v1"

// SliceReport is a per-group breakdown of the calibration metrics, used by
// Report.ByGroup so a reader can inspect calibration across strata (for
// example by topic, difficulty, or evidence kind) instead of one aggregate.
type SliceReport struct {
	Count            int      `json:"count" yaml:"count"`
	ExtractionRecall float64  `json:"extraction_recall" yaml:"extraction_recall"`
	Sensitivity      float64  `json:"sensitivity" yaml:"sensitivity"`
	Specificity      float64  `json:"specificity" yaml:"specificity"`
	FalseSupportRate float64  `json:"false_support_rate" yaml:"false_support_rate"`
	BrierScore       *float64 `json:"brier_score,omitempty" yaml:"brier_score,omitempty"`
	ECE              *float64 `json:"ece,omitempty" yaml:"ece,omitempty"`
}

// Report is the calibration evidence for one protocol against one gold
// dataset. BrierScore and ECE are nil when the protocol emits no confidence
// probabilities; calibration metrics apply only when confidence is present.
type Report struct {
	APIVersion       string                 `json:"api_version" yaml:"api_version"`
	ProtocolDigest   string                 `json:"protocol_digest" yaml:"protocol_digest"`
	DatasetDigest    string                 `json:"dataset_digest" yaml:"dataset_digest"`
	ExtractionRecall float64                `json:"extraction_recall" yaml:"extraction_recall"`
	Sensitivity      float64                `json:"sensitivity" yaml:"sensitivity"`
	Specificity      float64                `json:"specificity" yaml:"specificity"`
	FalseSupportRate float64                `json:"false_support_rate" yaml:"false_support_rate"`
	BrierScore       *float64               `json:"brier_score,omitempty" yaml:"brier_score,omitempty"`
	ECE              *float64               `json:"ece,omitempty" yaml:"ece,omitempty"`
	ByGroup          map[string]SliceReport `json:"by_group,omitempty" yaml:"by_group,omitempty"`
	Digest           string                 `json:"digest" yaml:"digest"`
}

// CalibrateInput is the data needed to compute a calibration report: the gold
// set, the model's reports keyed by instance ID, the protocol digest the
// reports were produced under, the number of ECE bins, and an optional matcher.
type CalibrateInput struct {
	Gold           GoldSet
	Reports        map[string]assessment.Report
	ProtocolDigest string
	Bins           int
	Matcher        ClaimMatcher
}

// Calibrate computes a calibration report by matching gold claims against the
// model's verdicts per instance. It builds a confusion matrix over matched
// claims, extraction recall over all gold claims, and Brier/ECE over claims
// with confidence. The report is sealed with a content-addressed digest.
func Calibrate(in CalibrateInput) (Report, error) {
	if err := ValidateGoldSet(&in.Gold); err != nil {
		return Report{}, fmt.Errorf("calibrate: %w", err)
	}
	if !strings.HasPrefix(in.ProtocolDigest, "sha256:") {
		return Report{}, fmt.Errorf("calibrate: protocol_digest must be a sha256: digest")
	}
	if in.Bins < 1 {
		in.Bins = 10
	}
	if in.Matcher == nil {
		in.Matcher = MatchByID
	}

	var allConf, allOutc []float64
	var confusion Confusion
	matched := 0
	total := 0
	for i := range in.Gold.Claims {
		g := &in.Gold.Claims[i]
		total++
		report, ok := in.Reports[g.InstanceID]
		if !ok {
			continue
		}
		verdict, found := findPredicted(report.ClaimResults, g.Claim, in.Matcher)
		if !found {
			continue
		}
		matched++
		confusion.Add(g.Label, verdict.Label)
		if verdict.Confidence != nil {
			allConf = append(allConf, *verdict.Confidence)
			if IsEntailed(g.Label) {
				allOutc = append(allOutc, 1)
			} else {
				allOutc = append(allOutc, 0)
			}
		}
	}

	report := Report{
		APIVersion:       ReportAPIVersion,
		ProtocolDigest:   in.ProtocolDigest,
		DatasetDigest:    in.Gold.Digest,
		ExtractionRecall: float64(matched) / float64(max1(total)),
		Sensitivity:      confusion.Sensitivity(),
		Specificity:      confusion.Specificity(),
		FalseSupportRate: confusion.FalseSupportRate(),
	}
	if len(allConf) > 0 {
		brier, err := BrierScore(allConf, allOutc)
		if err != nil {
			return Report{}, fmt.Errorf("calibrate: %w", err)
		}
		ece, err := ExpectedCalibrationError(allConf, allOutc, in.Bins)
		if err != nil {
			return Report{}, fmt.Errorf("calibrate: %w", err)
		}
		report.BrierScore = &brier
		report.ECE = &ece
	}
	if err := Seal(&report); err != nil {
		return Report{}, fmt.Errorf("calibrate: %w", err)
	}
	return report, nil
}

func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}

// ValidateReport returns nil when r is a well-formed, sealed calibration report.
func ValidateReport(r *Report) error {
	if err := validateReportBody(r); err != nil {
		return err
	}
	if !strings.HasPrefix(r.Digest, "sha256:") {
		return fmt.Errorf("calibration report: digest must be a sha256: digest")
	}
	return nil
}

// Seal validates the report body, computes its digest, and sets r.Digest.
func Seal(r *Report) error {
	if err := validateReportBody(r); err != nil {
		return err
	}
	digest, err := canonicaljson.Sum(reportDigestInput{
		APIVersion:       r.APIVersion,
		ProtocolDigest:   r.ProtocolDigest,
		DatasetDigest:    r.DatasetDigest,
		ExtractionRecall: r.ExtractionRecall,
		Sensitivity:      r.Sensitivity,
		Specificity:      r.Specificity,
		FalseSupportRate: r.FalseSupportRate,
		BrierScore:       r.BrierScore,
		ECE:              r.ECE,
		ByGroup:          r.ByGroup,
	})
	if err != nil {
		return fmt.Errorf("seal calibration report: %w", err)
	}
	r.Digest = digest
	return nil
}

// reportDigestInput is Report without the Digest field, so the digest is a
// non-circular function of report content.
type reportDigestInput struct {
	APIVersion       string                 `json:"api_version"`
	ProtocolDigest   string                 `json:"protocol_digest"`
	DatasetDigest    string                 `json:"dataset_digest"`
	ExtractionRecall float64                `json:"extraction_recall"`
	Sensitivity      float64                `json:"sensitivity"`
	Specificity      float64                `json:"specificity"`
	FalseSupportRate float64                `json:"false_support_rate"`
	BrierScore       *float64               `json:"brier_score,omitempty"`
	ECE              *float64               `json:"ece,omitempty"`
	ByGroup          map[string]SliceReport `json:"by_group,omitempty"`
}

func validateReportBody(r *Report) error {
	if r.APIVersion != ReportAPIVersion {
		return fmt.Errorf("calibration report: api_version %q is not supported (want %s)", r.APIVersion, ReportAPIVersion)
	}
	if !strings.HasPrefix(r.ProtocolDigest, "sha256:") {
		return fmt.Errorf("calibration report: protocol_digest must be a sha256: digest")
	}
	if !strings.HasPrefix(r.DatasetDigest, "sha256:") {
		return fmt.Errorf("calibration report: dataset_digest must be a sha256: digest")
	}
	if r.ExtractionRecall < 0 || r.ExtractionRecall > 1 {
		return fmt.Errorf("calibration report: extraction_recall %g must be in [0,1]", r.ExtractionRecall)
	}
	if r.Sensitivity < 0 || r.Sensitivity > 1 {
		return fmt.Errorf("calibration report: sensitivity %g must be in [0,1]", r.Sensitivity)
	}
	if r.Specificity < 0 || r.Specificity > 1 {
		return fmt.Errorf("calibration report: specificity %g must be in [0,1]", r.Specificity)
	}
	if r.FalseSupportRate < 0 || r.FalseSupportRate > 1 {
		return fmt.Errorf("calibration report: false_support_rate %g must be in [0,1]", r.FalseSupportRate)
	}
	if r.BrierScore != nil {
		b := *r.BrierScore
		if b < 0 || b > 1 {
			return fmt.Errorf("calibration report: brier_score %g must be in [0,1]", b)
		}
	}
	if r.ECE != nil {
		e := *r.ECE
		if e < 0 || e > 1 {
			return fmt.Errorf("calibration report: ece %g must be in [0,1]", e)
		}
	}
	return nil
}
