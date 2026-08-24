package suite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/judging"
)

func ptr(f float64) *float64 { return &f }

func makeInstance(t *testing.T, id string) eval.Instance {
	t.Helper()
	set, err := eval.NewEvidenceSet([]eval.EvidenceItem{
		{ID: "e1", Kind: "knowledge", Content: eval.NewTextArtifact("text/plain", "evidence"), SourceID: "s"},
	}, "sha256:policy")
	if err != nil {
		t.Fatalf("evidence: %v", err)
	}
	inst, err := eval.NewInstance(id,
		eval.NewTextArtifact("text/plain", "q"),
		eval.NewTextArtifact("text/plain", "a"),
		set, nil, nil, nil)
	if err != nil {
		t.Fatalf("instance: %v", err)
	}
	return inst
}

func reportFor(name string, value float64) assessment.Report {
	now := time.Date(2026, 8, 17, 12, 0, 0, 0, time.UTC)
	return assessment.Report{
		APIVersion:     assessment.ReportAPIVersion,
		InstanceID:     "inst",
		InstanceDigest: "sha256:inst",
		ProtocolDigest: "sha256:" + name,
		Dimensions:     []assessment.DimensionResult{{ConstructID: "faithfulness", Applicable: true, Value: ptr(value)}},
		StartedAt:      now,
		FinishedAt:     now.Add(time.Second),
	}
}

// stubJudge returns a fixed report for any instance.
type stubJudge struct {
	name  string
	value float64
}

func (s *stubJudge) Evaluate(_ context.Context, _ eval.Instance) (assessment.Report, error) {
	return reportFor(s.name, s.value), nil
}

var _ judging.Judge = (*stubJudge)(nil)

// dependEvaluator depends on another evaluator and records its name; it reads
// the dependency's report from Results to prove wiring.
type dependEvaluator struct {
	name   string
	deps   []string
	value  float64
	sawDep bool
}

func (d *dependEvaluator) Name() string        { return d.name }
func (d *dependEvaluator) DependsOn() []string { return d.deps }
func (d *dependEvaluator) Evaluate(_ context.Context, _ eval.Instance, results Results) (assessment.Report, error) {
	if _, ok := results.Report(d.deps[0]); ok {
		d.sawDep = true
	}
	return reportFor(d.name, d.value), nil
}

func TestNewSuiteRejectsCycle(t *testing.T) {
	a := &dependEvaluator{name: "a", deps: []string{"b"}, value: 0.1}
	b := &dependEvaluator{name: "b", deps: []string{"a"}, value: 0.2}
	if _, err := NewSuite("cyclic", []Evaluator{a, b}); err == nil {
		t.Errorf("accepted a cyclic suite")
	}
}

func TestNewSuiteRejectsUnknownDep(t *testing.T) {
	a := &dependEvaluator{name: "a", deps: []string{"missing"}, value: 0.1}
	if _, err := NewSuite("bad", []Evaluator{a}); err == nil {
		t.Errorf("accepted a dependency on an unknown evaluator")
	}
}

func TestNewSuiteRejectsDuplicateName(t *testing.T) {
	a := &dependEvaluator{name: "a", value: 0.1}
	b := &dependEvaluator{name: "a", value: 0.2}
	if _, err := NewSuite("dup", []Evaluator{a, b}); err == nil {
		t.Errorf("accepted duplicate evaluator names")
	}
}

func TestRunIndependentEvaluators(t *testing.T) {
	a := &JudgeEvaluator{EvaluatorName: "a", Judge: &stubJudge{name: "a", value: 0.5}}
	b := &JudgeEvaluator{EvaluatorName: "b", Judge: &stubJudge{name: "b", value: 0.9}}
	s, err := NewSuite("indep", []Evaluator{a, b})
	if err != nil {
		t.Fatalf("NewSuite: %v", err)
	}
	results, err := s.Run(context.Background(), makeInstance(t, "inst"))
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(results.Reports) != 2 {
		t.Fatalf("expected 2 reports, got %d", len(results.Reports))
	}
	if ra, ok := results.Report("a"); !ok || *ra.Dimensions[0].Value != 0.5 {
		t.Errorf("report a = %v", ra)
	}
	if rb, ok := results.Report("b"); !ok || *rb.Dimensions[0].Value != 0.9 {
		t.Errorf("report b = %v", rb)
	}
}

func TestRunDependentEvaluatorSeesDependency(t *testing.T) {
	// b depends on a and must see a's report in Results.
	a := &JudgeEvaluator{EvaluatorName: "a", Judge: &stubJudge{name: "a", value: 0.5}}
	b := &dependEvaluator{name: "b", deps: []string{"a"}, value: 0.9}
	s, err := NewSuite("dep", []Evaluator{a, b})
	if err != nil {
		t.Fatalf("NewSuite: %v", err)
	}
	if _, err := s.Run(context.Background(), makeInstance(t, "inst")); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !b.sawDep {
		t.Errorf("dependent evaluator b did not see dependency a's report")
	}
}

func TestRunDependentEvaluatorReceivesLiveContext(t *testing.T) {
	a := &JudgeEvaluator{EvaluatorName: "a", Judge: &stubJudge{name: "a", value: 0.5}}
	b := &contextCheckingEvaluator{name: "b", deps: []string{"a"}}
	s, err := NewSuite("live-context", []Evaluator{a, b})
	if err != nil {
		t.Fatalf("NewSuite: %v", err)
	}
	if _, err := s.Run(context.Background(), makeInstance(t, "inst")); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !b.ran {
		t.Errorf("dependent evaluator did not run")
	}
}

type contextCheckingEvaluator struct {
	name string
	deps []string
	ran  bool
}

func (e *contextCheckingEvaluator) Name() string        { return e.name }
func (e *contextCheckingEvaluator) DependsOn() []string { return e.deps }
func (e *contextCheckingEvaluator) Evaluate(ctx context.Context, _ eval.Instance, _ Results) (assessment.Report, error) {
	if err := ctx.Err(); err != nil {
		return assessment.Report{}, err
	}
	e.ran = true
	return reportFor(e.name, 1), nil
}

func TestRunPreservesEachReportIdentity(t *testing.T) {
	a := &JudgeEvaluator{EvaluatorName: "a", Judge: &stubJudge{name: "a", value: 0.5}}
	b := &JudgeEvaluator{EvaluatorName: "b", Judge: &stubJudge{name: "b", value: 0.9}}
	s, _ := NewSuite("identity", []Evaluator{a, b})
	results, _ := s.Run(context.Background(), makeInstance(t, "inst"))
	ra, _ := results.Report("a")
	rb, _ := results.Report("b")
	if ra.ProtocolDigest == rb.ProtocolDigest {
		t.Errorf("suite collapsed protocol identities: both %s", ra.ProtocolDigest)
	}
}

func TestRunErrorPropagates(t *testing.T) {
	fail := &failEvaluator{name: "fail"}
	s, err := NewSuite("fail", []Evaluator{fail})
	if err != nil {
		t.Fatalf("NewSuite: %v", err)
	}
	if _, err := s.Run(context.Background(), makeInstance(t, "inst")); err == nil {
		t.Errorf("expected evaluator error to propagate")
	}
}

type failEvaluator struct{ name string }

func (f *failEvaluator) Name() string        { return f.name }
func (f *failEvaluator) DependsOn() []string { return nil }
func (f *failEvaluator) Evaluate(context.Context, eval.Instance, Results) (assessment.Report, error) {
	return assessment.Report{}, errors.New("boom")
}

func TestSuiteDigestStable(t *testing.T) {
	a := &JudgeEvaluator{EvaluatorName: "a", Judge: &stubJudge{name: "a", value: 0.5}}
	b := &JudgeEvaluator{EvaluatorName: "b", Judge: &stubJudge{name: "b", value: 0.9}}
	s1, _ := NewSuite("same", []Evaluator{a, b})
	s2, _ := NewSuite("same", []Evaluator{a, b})
	if s1.Digest != s2.Digest {
		t.Errorf("suite digest not stable: %s vs %s", s1.Digest, s2.Digest)
	}
	// Reordering evaluators changes the digest only if names/dependencies
	// change; same names in different order should still be stable because
	// canonical JSON sorts map keys. Confirm order does not matter.
	s3, _ := NewSuite("same", []Evaluator{b, a})
	if s1.Digest != s3.Digest {
		t.Errorf("suite digest changed with evaluator order: %s vs %s", s1.Digest, s3.Digest)
	}
}

// TestRunConcurrent confirms independent evaluators actually run concurrently
// (the suite uses errgroup). Two evaluators that each sleep should overlap.
func TestRunConcurrent(t *testing.T) {
	const delay = 80 * time.Millisecond
	a := &sleepEvaluator{name: "a", delay: delay}
	b := &sleepEvaluator{name: "b", delay: delay}
	s, _ := NewSuite("concurrent", []Evaluator{a, b})
	start := time.Now()
	if _, err := s.Run(context.Background(), makeInstance(t, "inst")); err != nil {
		t.Fatalf("Run: %v", err)
	}
	elapsed := time.Since(start)
	// Sequential would take >= 2*delay; concurrent takes ~ delay. Allow slack.
	if elapsed >= 2*delay {
		t.Errorf("evaluators appear to have run sequentially: elapsed=%v, delay=%v", elapsed, delay)
	}
}

type sleepEvaluator struct {
	name  string
	delay time.Duration
}

func (s *sleepEvaluator) Name() string        { return s.name }
func (s *sleepEvaluator) DependsOn() []string { return nil }
func (s *sleepEvaluator) Evaluate(ctx context.Context, _ eval.Instance, _ Results) (assessment.Report, error) {
	select {
	case <-time.After(s.delay):
	case <-ctx.Done():
		return assessment.Report{}, ctx.Err()
	}
	return reportFor(s.name, 1.0), nil
}
