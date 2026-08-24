// Command judgekit is the thin CLI host for judgekit's Glazed help system.
//
// judgekit is library-first; this binary contains no domain logic. It exists so
// the help entries (getting-started, user-guide, developer-reference) are
// real, queryable documents wired through the canonical Glazed help path:
//
//	judgekit help
//	judgekit help getting-started
//	judgekit help user-guide
//	judgekit help developer-reference
package main

import (
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	help_cmd "github.com/go-go-golems/glazed/pkg/help/cmd"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/judgekit/pkg/doc"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	rootCmd := &cobra.Command{
		Use:   "judgekit",
		Short: "judgekit is a provider-neutral evaluation library (help host)",
		Long: "judgekit is a provider-neutral Go library for evaluating " +
			"language-model outputs with explicit, typed, versioned, and " +
			"auditable measurement. This CLI only hosts the help system; " +
			"the library is the product.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return logging.InitLoggerFromCobra(cmd)
		},
		SilenceUsage: true,
	}

	if err := logging.AddLoggingSectionToRootCommand(rootCmd, "judgekit"); err != nil {
		return fmt.Errorf("init logging: %w", err)
	}

	helpSystem := help.NewHelpSystem()
	if err := doc.AddDocToHelpSystem(helpSystem); err != nil {
		return fmt.Errorf("load help docs: %w", err)
	}
	help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)

	return rootCmd.Execute()
}
