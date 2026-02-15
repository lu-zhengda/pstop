package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/pstop/internal/process"
)

var findCmd = &cobra.Command{
	Use:   "find <query>",
	Short: "Find processes by name, command, or port",
	Long:  `Search for processes whose name or command contains the given query string.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		query := args[0]
		procs, err := process.Find(query)
		if err != nil {
			return fmt.Errorf("failed to find processes: %w", err)
		}
		if len(procs) == 0 {
			fmt.Printf("No processes found matching %q\n", query)
			return nil
		}
		printProcessTable(procs)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(findCmd)
}
