package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/zhengda-lu/pstop/internal/tui"
)

var (
	// version is set via ldflags at build time.
	version = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "pstop",
	Short: "Process explorer for macOS",
	Long: `pstop is a process explorer for macOS — browse, search, and manage
processes with a live-updating TUI or handy CLI subcommands.
Launch without subcommands for interactive TUI mode.`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := tea.NewProgram(tui.New(version), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.SetVersionTemplate(fmt.Sprintf("pstop %s\n", version))
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
