// Package doc embeds judgekit's Glazed help entries and registers them with a
// Glazed help system. It is the bridge between the markdown help files under
// pkg/doc/ and the cmd/judgekit CLI.
package doc

import (
	"embed"

	"github.com/go-go-golems/glazed/pkg/help"
)

// docFS embeds the topics, tutorials, and reference help entries. Glazed
// loads sections recursively from this FS.
//
//go:embed topics tutorials reference
var docFS embed.FS

// AddDocToHelpSystem loads all embedded help sections into hs so they become
// queryable via `judgekit help <slug>`.
func AddDocToHelpSystem(hs *help.HelpSystem) error {
	return hs.LoadSectionsFromFS(docFS, ".")
}
