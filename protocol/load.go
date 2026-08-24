package protocol

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/judgekit/internal/canonicaljson"
)

// SemanticDigest returns the canonical identity of a validated protocol.
func SemanticDigest(p *Protocol) (string, error) {
	return canonicaljson.Sum(p)
}

// ByteDigest returns "sha256:" plus the hex digest of the raw source bytes.
func ByteDigest(raw []byte) string {
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// LoadProtocol reads, strictly decodes, validates, and digests a protocol file.
func LoadProtocol(path string) (Document, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Document{}, fmt.Errorf("load protocol %s: %w", path, err)
	}
	return LoadProtocolFromBytes(path, raw)
}

// LoadProtocolFromBytes is like LoadProtocol but accepts the raw bytes.
func LoadProtocolFromBytes(path string, raw []byte) (Document, error) {
	var p Protocol
	if err := decodeProtocol(path, raw, &p); err != nil {
		return Document{}, err
	}
	if err := Validate(&p); err != nil {
		return Document{}, fmt.Errorf("load protocol %s: %w", path, err)
	}
	digest, err := SemanticDigest(&p)
	if err != nil {
		return Document{}, fmt.Errorf("load protocol %s: semantic digest: %w", path, err)
	}
	return Document{
		Protocol:   p,
		Digest:     digest,
		ByteDigest: ByteDigest(raw),
		Path:       path,
	}, nil
}

func decodeProtocol(path string, raw []byte, p *Protocol) error {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(p); err != nil {
			return fmt.Errorf("load protocol %s: decode json: %w", path, err)
		}
		var skip any
		if err := dec.Decode(&skip); err != io.EOF {
			return fmt.Errorf("load protocol %s: expected exactly one JSON object, found trailing data: %w", path, err)
		}
		return nil
	case ".yaml", ".yml":
		dec := yaml.NewDecoder(bytes.NewReader(raw))
		dec.KnownFields(true)
		if err := dec.Decode(p); err != nil {
			return fmt.Errorf("load protocol %s: decode yaml: %w", path, err)
		}
		return nil
	default:
		return fmt.Errorf("load protocol %s: unsupported extension %q (use .json, .yaml, or .yml)", path, ext)
	}
}
