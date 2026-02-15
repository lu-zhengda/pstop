package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var (
	topN       int
	topBattery bool
)

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top resource-consuming processes",
	Long:  `Show the top N processes sorted by CPU usage. Use --battery to highlight battery-draining processes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		procs, err := process.Top(topN)
		if err != nil {
			return fmt.Errorf("failed to get top processes: %w", err)
		}

		if jsonFlag {
			return printJSON(procs)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "PID\tNAME\tUSER\tCPU%\tMEM%\tSTATE\tCOMMAND")
		for _, p := range procs {
			prefix := ""
			if topBattery && p.CPU > 10.0 {
				prefix = "!! "
			}
			fmt.Fprintf(w, "%s%d\t%s\t%s\t%.1f\t%.1f\t%s\t%s\n",
				prefix, p.PID, p.Name, p.User, p.CPU, p.Mem, p.State, p.Command)
		}
		w.Flush()
		return nil
	},
}

func init() {
	topCmd.Flags().IntVarP(&topN, "n", "n", 10, "Number of processes to show")
	topCmd.Flags().BoolVar(&topBattery, "battery", false, "Highlight battery-draining processes (CPU > 10%)")
	rootCmd.AddCommand(topCmd)
}
