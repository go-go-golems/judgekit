package suite

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/go-go-golems/judgekit/assessment"
	"github.com/go-go-golems/judgekit/eval"
	"github.com/go-go-golems/judgekit/internal/canonicaljson"
	"github.com/go-go-golems/judgekit/internal/identifier"
	"github.com/go-go-golems/judgekit/judging"
	"golang.org/x/sync/errgroup"
)

// SuiteAPIVersion is the only suite API version judgekit accepts.
const SuiteAPIVersion = "judgekit.suite/v1"

// Results holds the reports produced by a suite run, keyed by evaluator name.
// One evaluator may read another's report from Results only when it declared
// that dependency and the dependency has already run.
type Results struct {
	Reports map[string]assessment.Report `json:"reports" yaml:"reports"`
}

// Report returns the report produced by the named evaluator, or false if it has
// not run yet or does not exist.
func (r Results) Report(name string) (assessment.Report, bool) {
	rep, ok := r.Reports[name]
	return rep, ok
}

// Evaluator is one judge in a suite. DependsOn names the evaluators whose
// results this evaluator may consume; the suite runs them first. Evaluate
// receives the instance and the partial Results available so far.
type Evaluator interface {
	Name() string
	DependsOn() []string
	Evaluate(ctx context.Context, inst eval.Instance, results Results) (assessment.Report, error)
}

// Suite is an acyclic graph of evaluators run in dependency order.
type Suite struct {
	APIVersion string      `json:"api_version" yaml:"api_version"`
	Name       string      `json:"name" yaml:"name"`
	Evaluators []Evaluator `json:"-"` // not serializable; carry identity via Name
	Digest     string      `json:"digest" yaml:"digest"`
}

// Validate returns nil when s is a well-formed suite: supported API version,
// valid name, unique evaluator names, and an acyclic dependency graph whose
// dependencies reference declared evaluators.
func (s Suite) Validate() error {
	if err := s.validateBody(); err != nil {
		return err
	}
	if !strings.HasPrefix(s.Digest, "sha256:") {
		return fmt.Errorf("suite %q: digest must be a sha256: digest", s.Name)
	}
	return nil
}

// validateBody performs all suite checks except the digest-presence check, so
// NewSuite can validate structure before the digest is computed.
func (s Suite) validateBody() error {
	if s.APIVersion != SuiteAPIVersion {
		return fmt.Errorf("suite: api_version %q is not supported (want %s)", s.APIVersion, SuiteAPIVersion)
	}
	if err := identifier.Validate(s.Name); err != nil {
		return fmt.Errorf("suite name: %w", err)
	}
	if len(s.Evaluators) == 0 {
		return fmt.Errorf("suite %q: at least one evaluator is required", s.Name)
	}
	names := make(map[string]struct{}, len(s.Evaluators))
	for _, e := range s.Evaluators {
		if e == nil {
			return fmt.Errorf("suite %q: nil evaluator", s.Name)
		}
		if err := identifier.Validate(e.Name()); err != nil {
			return fmt.Errorf("suite %q: evaluator name: %w", s.Name, err)
		}
		if _, dup := names[e.Name()]; dup {
			return fmt.Errorf("suite %q: duplicate evaluator %q", s.Name, e.Name())
		}
		names[e.Name()] = struct{}{}
	}
	for _, e := range s.Evaluators {
		for _, dep := range e.DependsOn() {
			if _, ok := names[dep]; !ok {
				return fmt.Errorf("suite %q: evaluator %q depends on unknown evaluator %q", s.Name, e.Name(), dep)
			}
		}
	}
	if err := checkAcyclic(s.Evaluators); err != nil {
		return err
	}
	return nil
}

// checkAcyclic rejects a dependency cycle. It does a DFS over the declared
// dependency edges and reports the first cycle it finds.
func checkAcyclic(evaluators []Evaluator) error {
	names := make(map[string]Evaluator, len(evaluators))
	for _, e := range evaluators {
		names[e.Name()] = e
	}
	const (
		white = 0 // unvisited
		gray  = 1 // on the current path
		black = 2 // fully explored
	)
	color := make(map[string]int, len(evaluators))
	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		if color[name] == gray {
			return fmt.Errorf("suite: dependency cycle: %s -> %s", strings.Join(path, " -> "), name)
		}
		if color[name] == black {
			return nil
		}
		color[name] = gray
		for _, dep := range names[name].DependsOn() {
			if err := visit(dep, append(path, name)); err != nil {
				return err
			}
		}
		color[name] = black
		return nil
	}
	for _, e := range evaluators {
		if err := visit(e.Name(), nil); err != nil {
			return err
		}
	}
	return nil
}

// Run executes the evaluators in dependency order. Evaluators with no remaining
// unmet dependencies run concurrently via errgroup; an evaluator that declares
// dependencies runs after they have completed and may read their reports from
// the shared Results.
func (s Suite) Run(ctx context.Context, inst eval.Instance) (Results, error) {
	if err := s.Validate(); err != nil {
		return Results{}, err
	}
	if err := eval.ValidateInstance(&inst); err != nil {
		return Results{}, fmt.Errorf("suite: invalid instance: %w", err)
	}

	results := Results{Reports: make(map[string]assessment.Report, len(s.Evaluators))}

	// remaining tracks, per evaluator, the dependencies not yet satisfied.
	remaining := make(map[string]map[string]struct{}, len(s.Evaluators))
	for _, e := range s.Evaluators {
		deps := make(map[string]struct{}, len(e.DependsOn()))
		for _, d := range e.DependsOn() {
			deps[d] = struct{}{}
		}
		remaining[e.Name()] = deps
	}
	pending := make(map[string]Evaluator, len(s.Evaluators))
	for _, e := range s.Evaluators {
		pending[e.Name()] = e
	}

	for len(pending) > 0 {
		// Find one dependency wave. Scheduling state stays single-threaded;
		// workers only write to distinct result slots.
		ready := make([]Evaluator, 0)
		for name, e := range pending {
			if len(remaining[name]) == 0 {
				ready = append(ready, e)
			}
		}
		if len(ready) == 0 {
			return Results{}, fmt.Errorf("suite %q: no ready evaluators but %d pending (deadlock)", s.Name, len(pending))
		}
		sort.Slice(ready, func(i, j int) bool { return ready[i].Name() < ready[j].Name() })
		snapshot := snapshotResults(results)
		waveReports := make([]assessment.Report, len(ready))
		g, waveCtx := errgroup.WithContext(ctx)
		for i, evaluator := range ready {
			i, evaluator := i, evaluator
			g.Go(func() error {
				report, err := evaluator.Evaluate(waveCtx, inst, snapshot)
				if err != nil {
					return fmt.Errorf("suite: evaluator %q: %w", evaluator.Name(), err)
				}
				waveReports[i] = report
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			return Results{}, err
		}
		// errgroup cancels waveCtx after Wait, so the next iteration creates a
		// fresh group/context. Publish and satisfy the completed wave now.
		for i, evaluator := range ready {
			name := evaluator.Name()
			results.Reports[name] = waveReports[i]
			delete(pending, name)
			for dependent := range remaining {
				delete(remaining[dependent], name)
			}
		}
	}
	return results, nil
}

// snapshotResults copies the current reports so a dispatched evaluator sees a
// stable view independent of later concurrent writes.
func snapshotResults(r Results) Results {
	out := Results{Reports: make(map[string]assessment.Report, len(r.Reports))}
	for k, v := range r.Reports {
		out.Reports[k] = v
	}
	return out
}

// NewSuite validates the evaluators and returns a content-addressed Suite.
func NewSuite(name string, evaluators []Evaluator) (Suite, error) {
	s := Suite{APIVersion: SuiteAPIVersion, Name: name, Evaluators: evaluators}
	if err := s.validateBody(); err != nil {
		return Suite{}, err
	}
	digest, err := canonicaljson.Sum(suiteDigestInput{
		APIVersion:   s.APIVersion,
		Name:         s.Name,
		Dependencies: evaluatorDeps(evaluators),
	})
	if err != nil {
		return Suite{}, fmt.Errorf("suite digest: %w", err)
	}
	s.Digest = digest
	return s, nil
}

func evaluatorDeps(evaluators []Evaluator) map[string][]string {
	deps := make(map[string][]string, len(evaluators))
	for _, e := range evaluators {
		d := append([]string{}, e.DependsOn()...)
		sort.Strings(d)
		deps[e.Name()] = d
	}
	return deps
}

type suiteDigestInput struct {
	APIVersion   string              `json:"api_version"`
	Name         string              `json:"name"`
	Dependencies map[string][]string `json:"dependencies"`
}

// JudgeEvaluator adapts a judging.Judge to a suite.Evaluator with no
// dependencies. It lets a suite run a plain judge without the caller
// implementing the Evaluator interface by hand.
type JudgeEvaluator struct {
	EvaluatorName string
	Judge         judging.Judge
}

// Name returns the evaluator's name.
func (j JudgeEvaluator) Name() string { return j.EvaluatorName }

// DependsOn returns no dependencies.
func (j JudgeEvaluator) DependsOn() []string { return nil }

// Evaluate runs the wrapped judge over the instance.
func (j JudgeEvaluator) Evaluate(ctx context.Context, inst eval.Instance, _ Results) (assessment.Report, error) {
	if j.Judge == nil {
		return assessment.Report{}, fmt.Errorf("evaluator %q: wrapped judge is nil", j.EvaluatorName)
	}
	return j.Judge.Evaluate(ctx, inst)
}

var _ Evaluator = JudgeEvaluator{}
