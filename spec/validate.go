package spec

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-go-golems/judgekit/internal/identifier"
)

// ValidateConstruct returns nil when c is a well-formed construct.
func ValidateConstruct(c *Construct) error {
	if err := identifier.Validate(string(c.ID)); err != nil {
		return fmt.Errorf("construct %q: %w", c.ID, err)
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("construct %q: name is required", c.ID)
	}
	if strings.TrimSpace(c.Definition) == "" {
		return fmt.Errorf("construct %q: definition is required", c.ID)
	}
	if strings.TrimSpace(c.Unit) == "" {
		return fmt.Errorf("construct %q: unit is required", c.ID)
	}
	if !validDirections[c.Direction] {
		return fmt.Errorf("construct %q: direction %q is not one of maximize, minimize, descriptive", c.ID, c.Direction)
	}
	if c.Range != nil {
		if err := validateRange(c.Range); err != nil {
			return fmt.Errorf("construct %q: %w", c.ID, err)
		}
	}
	return nil
}

func validateRange(r *Range) error {
	if math.IsNaN(r.Minimum) || math.IsInf(r.Minimum, 0) {
		return fmt.Errorf("range minimum is not finite")
	}
	if math.IsNaN(r.Maximum) || math.IsInf(r.Maximum, 0) {
		return fmt.Errorf("range maximum is not finite")
	}
	if r.Minimum > r.Maximum {
		return fmt.Errorf("range minimum %g is greater than maximum %g", r.Minimum, r.Maximum)
	}
	return nil
}

// ValidateContract returns nil when c is a well-formed measurement contract.
// It fails closed on unsupported API versions, duplicate construct IDs,
// dangling label/aggregation/exclusion references, and invalid aggregations.
func ValidateContract(c *MeasurementContract) error {
	if c.APIVersion != ContractAPIVersion {
		return fmt.Errorf("contract %q: api_version %q is not supported (want %s)", c.Name, c.APIVersion, ContractAPIVersion)
	}
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("contract: name is required")
	}
	if len(c.Constructs) == 0 {
		return fmt.Errorf("contract %q: at least one construct is required", c.Name)
	}
	ids := make(map[ConstructID]struct{}, len(c.Constructs))
	for i := range c.Constructs {
		construct := &c.Constructs[i]
		if err := ValidateConstruct(construct); err != nil {
			return err
		}
		if _, dup := ids[construct.ID]; dup {
			return fmt.Errorf("contract %q: duplicate construct id %q", c.Name, construct.ID)
		}
		ids[construct.ID] = struct{}{}
	}
	if err := validateEvidencePolicy(&c.EvidencePolicy); err != nil {
		return fmt.Errorf("contract %q: %w", c.Name, err)
	}
	if err := validateLabels(c, ids); err != nil {
		return err
	}
	if err := validateAggregations(c, ids); err != nil {
		return err
	}
	if err := validateExclusions(c, ids); err != nil {
		return err
	}
	return nil
}

func validateEvidencePolicy(p *EvidencePolicy) error {
	allowed := checkKindSet(p.AllowedKinds, "allowed_kinds")
	if allowed != nil {
		return allowed
	}
	forbidden := checkKindSet(p.ForbiddenKinds, "forbidden_kinds")
	if forbidden != nil {
		return forbidden
	}
	overlap := make(map[string]struct{}, len(p.AllowedKinds))
	for _, k := range p.AllowedKinds {
		overlap[k] = struct{}{}
	}
	for _, k := range p.ForbiddenKinds {
		if _, hit := overlap[k]; hit {
			return fmt.Errorf("evidence kind %q is both allowed and forbidden", k)
		}
	}
	return nil
}

// checkKindSet rejects empty, duplicate, or blank entries in a kind list.
func checkKindSet(kinds []string, field string) error {
	seen := make(map[string]struct{}, len(kinds))
	for _, k := range kinds {
		if strings.TrimSpace(k) == "" {
			return fmt.Errorf("evidence_policy.%s contains a blank kind", field)
		}
		if _, dup := seen[k]; dup {
			return fmt.Errorf("evidence_policy.%s contains duplicate kind %q", field, k)
		}
		seen[k] = struct{}{}
	}
	return nil
}

func validateLabels(c *MeasurementContract, ids map[ConstructID]struct{}) error {
	for cid, labels := range c.Labels {
		if _, ok := ids[cid]; !ok {
			return fmt.Errorf("contract %q: labels reference unknown construct %q", c.Name, cid)
		}
		seen := make(map[string]struct{}, len(labels))
		for _, l := range labels {
			if strings.TrimSpace(l) == "" {
				return fmt.Errorf("contract %q: construct %q has a blank label", c.Name, cid)
			}
			if _, dup := seen[l]; dup {
				return fmt.Errorf("contract %q: construct %q has duplicate label %q", c.Name, cid, l)
			}
			seen[l] = struct{}{}
		}
	}
	return nil
}

func validateAggregations(c *MeasurementContract, ids map[ConstructID]struct{}) error {
	for cid := range ids {
		agg, ok := c.Aggregations[cid]
		if !ok {
			return fmt.Errorf("contract %q: construct %q has no aggregation", c.Name, cid)
		}
		if err := validateAggregation(c, cid, agg); err != nil {
			return err
		}
	}
	return nil
}

func validateAggregation(c *MeasurementContract, cid ConstructID, agg Aggregation) error {
	if !validMethods[agg.Method] {
		return fmt.Errorf("contract %q: construct %q: method %q is not recognized", c.Name, cid, agg.Method)
	}
	if !validEmptyPolicies[agg.EmptyPolicy] {
		return fmt.Errorf("contract %q: construct %q: empty_policy %q is not recognized", c.Name, cid, agg.EmptyPolicy)
	}
	if agg.Method == MethodFraction {
		labels := c.Labels[cid]
		labelSet := make(map[string]struct{}, len(labels))
		for _, l := range labels {
			labelSet[l] = struct{}{}
		}
		if strings.TrimSpace(agg.Numerator) == "" {
			return fmt.Errorf("contract %q: construct %q: fraction aggregation requires numerator", c.Name, cid)
		}
		if strings.TrimSpace(agg.Denominator) == "" {
			return fmt.Errorf("contract %q: construct %q: fraction aggregation requires denominator", c.Name, cid)
		}
		if err := validateLabelList(c.Name, cid, "numerator", agg.Numerator, labelSet); err != nil {
			return err
		}
		if err := validateLabelList(c.Name, cid, "denominator", agg.Denominator, labelSet); err != nil {
			return err
		}
	}
	return nil
}

// validateLabelList checks that a comma-separated list of labels is non-empty,
// free of duplicates, and references only declared labels.
func validateLabelList(name string, cid ConstructID, field string, list string, labelSet map[string]struct{}) error {
	parts := strings.Split(list, ",")
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			return fmt.Errorf("contract %q: construct %q: %s has a blank label", name, cid, field)
		}
		if _, ok := labelSet[p]; !ok {
			return fmt.Errorf("contract %q: construct %q: %s references undeclared label %q", name, cid, field, p)
		}
		if _, dup := seen[p]; dup {
			return fmt.Errorf("contract %q: construct %q: %s repeats label %q", name, cid, field, p)
		}
		seen[p] = struct{}{}
	}
	return nil
}

func validateExclusions(c *MeasurementContract, ids map[ConstructID]struct{}) error {
	for cid, excls := range c.Exclusions {
		if _, ok := ids[cid]; !ok {
			return fmt.Errorf("contract %q: exclusions reference unknown construct %q", c.Name, cid)
		}
		seen := make(map[string]struct{}, len(excls))
		for _, e := range excls {
			if strings.TrimSpace(e) == "" {
				return fmt.Errorf("contract %q: construct %q has a blank exclusion", c.Name, cid)
			}
			if _, dup := seen[e]; dup {
				return fmt.Errorf("contract %q: construct %q has duplicate exclusion %q", c.Name, cid, e)
			}
			seen[e] = struct{}{}
		}
	}
	return nil
}
