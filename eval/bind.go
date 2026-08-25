package eval

import "fmt"

// BindCurrentIdentity recomputes the evidence-set and instance digests from
// current content at an execution boundary, then validates the resulting
// snapshot. It prevents a stale caller-supplied digest from being trusted by a
// cache or report while keeping ordinary input values mutable during assembly.
func BindCurrentIdentity(instance Instance) (Instance, error) {
	evidenceDigest, err := EvidenceSetDigest(instance.Evidence.Items, instance.Evidence.PolicyDigest)
	if err != nil {
		return Instance{}, fmt.Errorf("bind instance evidence: %w", err)
	}
	instance.Evidence.Digest = evidenceDigest
	instanceDigest, err := InstanceDigest(&instance)
	if err != nil {
		return Instance{}, fmt.Errorf("bind instance: %w", err)
	}
	instance.Digest = instanceDigest
	if err := ValidateInstance(&instance); err != nil {
		return Instance{}, err
	}
	return instance, nil
}
