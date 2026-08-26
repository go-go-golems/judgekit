package protocol

import "testing"

func TestValidateObservedModelRequiresExactResearchIdentity(t *testing.T) {
	expected := ModelIdentity{Provider: "fake", Model: "judge-1", Revision: "r1", Settings: map[string]string{"region": "test"}}
	if err := ValidateObservedModel(expected, expected); err != nil {
		t.Fatalf("matching model: %v", err)
	}
	cases := []ModelIdentity{
		{Provider: "other", Model: "judge-1", Revision: "r1", Settings: map[string]string{"region": "test"}},
		{Provider: "fake", Model: "judge-2", Revision: "r1", Settings: map[string]string{"region": "test"}},
		{Provider: "fake", Model: "judge-1", Revision: "r2", Settings: map[string]string{"region": "test"}},
		{Provider: "fake", Model: "judge-1", Revision: "r1", Settings: map[string]string{"region": "other"}},
		{},
	}
	for _, observed := range cases {
		if err := ValidateObservedModel(expected, observed); err == nil {
			t.Errorf("accepted mismatched observed model: %+v", observed)
		}
	}
}
