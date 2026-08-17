package spec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadContract reads, strictly decodes, validates, and digests a measurement
// contract file. The format is chosen by extension: ".json" for JSON,
// ".yaml" or ".yml" for YAML. Unknown fields are rejected in both formats so a
// typo cannot silently create a partial contract.
func LoadContract(path string) (ContractDocument, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ContractDocument{}, fmt.Errorf("load contract %s: %w", path, err)
	}
	return LoadContractFromBytes(path, raw)
}

// LoadContractFromBytes is like LoadContract but accepts the raw bytes, so
// contracts can be loaded from fixtures and embedded resources without a
// file on disk. The path is used only for error messages and the ByteDigest.
func LoadContractFromBytes(path string, raw []byte) (ContractDocument, error) {
	var c MeasurementContract
	if err := decodeContract(path, raw, &c); err != nil {
		return ContractDocument{}, err
	}
	if err := ValidateContract(&c); err != nil {
		return ContractDocument{}, fmt.Errorf("load contract %s: %w", path, err)
	}
	digest, err := SemanticDigest(&c)
	if err != nil {
		return ContractDocument{}, fmt.Errorf("load contract %s: semantic digest: %w", path, err)
	}
	return ContractDocument{
		Contract:   c,
		Digest:     digest,
		ByteDigest: ByteDigest(raw),
		Path:       path,
	}, nil
}

func decodeContract(path string, raw []byte, c *MeasurementContract) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(c); err != nil {
			return fmt.Errorf("load contract %s: decode json: %w", path, err)
		}
		var skip any
		if err := dec.Decode(&skip); err != io.EOF {
			return fmt.Errorf("load contract %s: expected exactly one JSON object, found trailing data: %w", path, err)
		}
		return nil
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		dec.KnownFields(true)
		if err := dec.Decode(c); err != nil {
			return fmt.Errorf("load contract %s: decode yaml: %w", path, err)
		}
		return nil
	default:
		return fmt.Errorf("load contract %s: unsupported extension %q (use .json, .yaml, or .yml)", path, ext)
	}
}
