package strictdecode

import (
	"errors"
	"testing"
)

type statementsPayload struct {
	Statements []string `json:"statements"`
}

func TestDecodeJSONObjectStrictAccepts(t *testing.T) {
	raw := `{"statements":["a","b"]}`
	got, err := DecodeJSONObjectStrict[statementsPayload](raw)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got.Statements) != 2 || got.Statements[0] != "a" {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestDecodeJSONObjectStrictStripsSingleFence(t *testing.T) {
	raw := "```json\n{\"statements\":[\"a\"]}\n```"
	got, err := DecodeJSONObjectStrict[statementsPayload](raw)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}
	if len(got.Statements) != 1 {
		t.Errorf("unexpected result: %+v", got)
	}
}

func TestDecodeJSONObjectStrictRejectsProse(t *testing.T) {
	raw := `Here is the answer: {"statements":["a"]} thanks.`
	if _, err := DecodeJSONObjectStrict[statementsPayload](raw); err == nil {
		t.Errorf("Decode accepted prose-wrapped JSON, want structural error")
	}
}

func TestDecodeJSONObjectStrictRejectsArray(t *testing.T) {
	raw := `["a","b"]`
	if _, err := DecodeJSONObjectStrict[statementsPayload](raw); err == nil {
		t.Errorf("Decode accepted an array, want structural error")
	}
}

func TestDecodeJSONObjectStrictRejectsUnknownField(t *testing.T) {
	raw := `{"statements":["a"],"extra":1}`
	if _, err := DecodeJSONObjectStrict[statementsPayload](raw); err == nil {
		t.Errorf("Decode accepted unknown field, want structural error")
	}
}

func TestDecodeJSONObjectStrictRejectsTrailingData(t *testing.T) {
	raw := `{"statements":["a"]}{"statements":["b"]}`
	if _, err := DecodeJSONObjectStrict[statementsPayload](raw); err == nil {
		t.Errorf("Decode accepted trailing JSON, want structural error")
	}
}

func TestDecodeJSONObjectStrictRejectsEmpty(t *testing.T) {
	if _, err := DecodeJSONObjectStrict[statementsPayload]("   "); err == nil {
		t.Errorf("Decode accepted empty input, want structural error")
	}
}

func TestIsStructural(t *testing.T) {
	_, err := DecodeJSONObjectStrict[statementsPayload]("nope")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !IsStructural(err) {
		t.Errorf("IsStructural = false, want true")
	}
	if !errors.As(err, new(*StructuralError)) {
		t.Errorf("errors.As(*StructuralError) = false, want true")
	}
}

func TestStripSingleCodeFenceLeavesEmbeddedFence(t *testing.T) {
	// A fence that does not wrap the whole response must be left intact so the
	// strict decoder rejects the surrounding prose.
	raw := "prose ```json\n{}\n``` more prose"
	if got := StripSingleCodeFence(raw); got != raw {
		t.Errorf("StripSingleCodeFence modified non-wrapping fence: got %q", got)
	}
}

func TestStripSingleCodeFenceStripsWrappingFence(t *testing.T) {
	raw := "```\n{\"a\":1}\n```"
	if got := StripSingleCodeFence(raw); got != `{"a":1}` {
		t.Errorf("StripSingleCodeFence = %q, want %q", got, `{"a":1}`)
	}
}

func TestStripSingleCodeFenceHandlesLanguageTag(t *testing.T) {
	raw := "```json\n{\"a\":1}\n```"
	if got := StripSingleCodeFence(raw); got != `{"a":1}` {
		t.Errorf("StripSingleCodeFence = %q, want %q", got, `{"a":1}`)
	}
}

func TestValidateSingleJSONValue(t *testing.T) {
	if err := ValidateSingleJSONValue(`{"a":1}`); err != nil {
		t.Errorf("ValidateSingleJSONValue(valid) = %v, want nil", err)
	}
	if err := ValidateSingleJSONValue(`{"a":1}{"b":2}`); err == nil {
		t.Errorf("ValidateSingleJSONValue(trailing) = nil, want error")
	}
	if err := ValidateSingleJSONValue(``); err == nil {
		t.Errorf("ValidateSingleJSONValue(empty) = nil, want error")
	}
	if err := ValidateSingleJSONValue(`123`); err != nil {
		t.Errorf("ValidateSingleJSONValue(number) = %v, want nil", err)
	}
}
