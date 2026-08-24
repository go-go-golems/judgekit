package eval

import (
	"crypto/sha256"
	"encoding/hex"
)

// Artifact is a content-addressed piece of text or a reference to one. A text
// artifact carries its content inline; a URI artifact refers to content the
// application resolves before evaluation. Exactly one of Text or URI must be
// non-empty, and Digest must be "sha256:" plus the hex digest of the content.
type Artifact struct {
	MediaType string `json:"media_type" yaml:"media_type"`
	Text      string `json:"text,omitempty" yaml:"text,omitempty"`
	URI       string `json:"uri,omitempty" yaml:"uri,omitempty"`
	Digest    string `json:"digest" yaml:"digest"`
	SizeBytes int64  `json:"size_bytes" yaml:"size_bytes"`
}

// TextContentDigest returns "sha256:" plus the hex SHA-256 of the UTF-8 bytes
// of text. It is the canonical digest for an inline text artifact.
func TextContentDigest(text string) string {
	sum := sha256.Sum256([]byte(text))
	return "sha256:" + hex.EncodeToString(sum[:])
}

// NewTextArtifact builds a valid text artifact, computing its digest and size
// from the content so callers cannot forget to content-address it.
func NewTextArtifact(mediaType, text string) Artifact {
	return Artifact{
		MediaType: mediaType,
		Text:      text,
		Digest:    TextContentDigest(text),
		SizeBytes: int64(len(text)),
	}
}
