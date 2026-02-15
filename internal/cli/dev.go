package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Show developer processes grouped by stack",
	Long:  `Group running processes by development stack (Node.js, Python, Docker, etc.) and show resource usage.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		groups, err := process.GroupByStack()
		if err != nil {
			return fmt.Errorf("failed to group processes: %w", err)
		}

		if jsonFlag {
			return printJSON(groups)
		}

		if len(groups) == 0 {
			fmt.Println("No developer processes found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		for _, g := range groups {
			fmt.Fprintf(w, "\n=== %s (%d processes) ===\tCPU: %.1f%%\tMEM: %.1f%%\n",
				g.Stack, len(g.Processes), g.TotalCPU, g.TotalMem)
			fmt.Fprintln(w, "PID\tNAME\tCPU%\tMEM%")
			for _, p := range g.Processes {
				fmt.Fprintf(w, "%d\t%s\t%.1f\t%.1f\n",
					p.PID, p.Name, p.CPU, p.Mem)
			}
		}
		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(devCmd)
}
