package identifier

import (
	"errors"
	"testing"
)

func TestValidateAccepts(t *testing.T) {
	cases := []string{
		"a",
		"faithfulness",
		"evidence-faithfulness",
		"claim_1",
		"e1",
		"sql1",
		"ab-_cd",
	}
	for _, id := range cases {
		if err := Validate(id); err != nil {
			t.Errorf("Validate(%q) = %v, want nil", id, err)
		}
	}
}

func TestValidateRejects(t *testing.T) {
	cases := []string{
		"",
		"-leading",
		"trailing-",
		"_leading",
		"trailing_",
		"Upper",
		"with space",
		"with.dot",
		"with/slash",
	}
	for _, id := range cases {
		if err := Validate(id); err == nil {
			t.Errorf("Validate(%q) = nil, want error", id)
		}
	}
}

func TestValidateRejectsTooLong(t *testing.T) {
	id := make([]byte, MaxLen+1)
	for i := range id {
		id[i] = 'a'
	}
	if err := Validate(string(id)); err == nil {
		t.Errorf("Validate(%d bytes) = nil, want error", len(id))
	}
	if err := Validate(string(id[:MaxLen])); err != nil {
		t.Errorf("Validate(%d bytes) = %v, want nil", MaxLen, err)
	}
}

func TestMustValidatePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustValidate did not panic on invalid id")
		}
	}()
	MustValidate("-bad")
}

func TestValidateErrorIsComparable(t *testing.T) {
	err := Validate("bad id")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, err) {
		t.Errorf("errors.Is identity check failed")
	}
}
