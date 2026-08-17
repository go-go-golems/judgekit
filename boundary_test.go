package judgekit

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// forbiddenImportPrefixes are packages core must never depend on. They are
// either application frameworks (Glazed, Cobra, Bubble Tea), provider SDKs,
// or sibling products whose concerns do not belong in a provider-neutral
// evaluation library.
var forbiddenImportPrefixes = []string{
	"github.com/go-go-golems/glazed",
	"github.com/spf13/cobra",
	"github.com/spf13/pflag",
	"github.com/charmbracelet/bubbletea",
	"github.com/charmbracelet/bubbles",
	"github.com/go-go-golems/geppetto",
	"github.com/go-go-golems/pinocchio",
	"github.com/go-go-golems/coinvault",
	"github.com/go-go-golems/ragopt",
	"github.com/go-go-golems/ragkit",
	"github.com/sashabaranov/go-openai",
	"github.com/anthropics/anthropic-sdk-go",
}

type listPackage struct {
	ImportPath  string
	Imports     []string
	TestImports []string
	Deps        []string
}

// TestCorePackageBoundaries asserts that no core package (spec, eval, protocol,
// assessment, judging, or internal helpers) imports a forbidden framework,
// provider SDK, or sibling product. cmd/judgekit is intentionally excluded: it
// is the only place allowed to import Glazed/Cobra to host the help system.
func TestCorePackageBoundaries(t *testing.T) {
	cmd := exec.Command("go", "list", "-json", "./spec", "./eval", "./protocol", "./assessment", "./judging", "./audit", "./calibration", "./suite", "./internal/...")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("go list: %v: %s", err, out)
	}
	dec := json.NewDecoder(strings.NewReader(string(out)))
	checked := 0
	for {
		var pkg listPackage
		if err := dec.Decode(&pkg); err != nil {
			break
		}
		checked++
		imports := append([]string{}, pkg.Imports...)
		imports = append(imports, pkg.TestImports...)
		for _, imp := range imports {
			for _, f := range forbiddenImportPrefixes {
				if strings.HasPrefix(imp, f) {
					t.Errorf("core package %s imports forbidden %q", pkg.ImportPath, imp)
				}
			}
		}
	}
	if checked == 0 {
		t.Fatalf("go list returned no packages; boundary check did not run")
	}
}
