package canonicaljson

import (
	"math"
	"strings"
	"testing"
)

func TestMarshalSortsKeys(t *testing.T) {
	in := map[string]any{
		"zeta":  1,
		"alpha": 2,
		"beta":  3,
	}
	got, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"alpha":2,"beta":3,"zeta":1}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

func TestMarshalIsCompact(t *testing.T) {
	in := map[string]any{"a": map[string]any{"b": []any{1, 2, 3}}}
	got, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if strings.ContainsAny(string(got), " \n\t") {
		t.Errorf("Marshal produced whitespace: %s", got)
	}
	want := `{"a":{"b":[1,2,3]}}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

func TestMarshalNoHTMLEscape(t *testing.T) {
	in := map[string]any{"html": "<b>&</b>"}
	got, err := Marshal(in)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"html":"<b>&</b>"}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s (HTML chars must not be escaped)", got, want)
	}
}

func TestMarshalRejectsNonFinite(t *testing.T) {
	for _, v := range []any{
		map[string]any{"x": math.NaN()},
		map[string]any{"y": math.Inf(1)},
		map[string]any{"z": math.Inf(-1)},
	} {
		if _, err := Marshal(v); err == nil {
			t.Errorf("Marshal(non-finite) = nil, want error")
		}
	}
}

func TestMarshalStructFieldOrderIrrelevant(t *testing.T) {
	// Two structs with the same fields in different declaration order must
	// produce the same canonical bytes after the generic round-trip.
	type a struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	type b struct {
		B int `json:"b"`
		A int `json:"a"`
	}
	v1 := a{A: 1, B: 2}
	v2 := b{A: 1, B: 2}
	s1, err := Marshal(v1)
	if err != nil {
		t.Fatalf("Marshal v1: %v", err)
	}
	s2, err := Marshal(v2)
	if err != nil {
		t.Fatalf("Marshal v2: %v", err)
	}
	if string(s1) != string(s2) {
		t.Errorf("struct field order affected canonical form: %s vs %s", s1, s2)
	}
}

func TestSumIsDeterministic(t *testing.T) {
	m1 := map[string]any{"a": 1, "b": 2}
	m2 := map[string]any{"b": 2, "a": 1}
	s1, err := Sum(m1)
	if err != nil {
		t.Fatalf("Sum m1: %v", err)
	}
	s2, err := Sum(m2)
	if err != nil {
		t.Fatalf("Sum m2: %v", err)
	}
	if s1 != s2 {
		t.Errorf("Sum not deterministic across key order: %s vs %s", s1, s2)
	}
	if !strings.HasPrefix(s1, "sha256:") {
		t.Errorf("Sum missing sha256 prefix: %s", s1)
	}
}

func TestSumChangesOnSemanticChange(t *testing.T) {
	s1, err := Sum(map[string]any{"a": 1, "b": 2})
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	s2, err := Sum(map[string]any{"a": 1, "b": 3})
	if err != nil {
		t.Fatalf("Sum: %v", err)
	}
	if s1 == s2 {
		t.Errorf("Sum did not change when a value changed")
	}
}

func TestNumberFormat(t *testing.T) {
	cases := []struct {
		in   any
		want string
	}{
		{5.0, `5`},
		{0.5, `0.5`},
		{0.86, `0.86`},
		{1.5, `1.5`},
		{0.0, `0`},
	}
	for _, c := range cases {
		got, err := Marshal(map[string]any{"n": c.in})
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		want := `{"n":` + c.want + `}`
		if string(got) != want {
			t.Errorf("Marshal(%v) = %s, want %s", c.in, got, want)
		}
	}
}
