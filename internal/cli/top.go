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
	topFormat  string
)

// sparkChars maps a 0.0-1.0 value to a sparkline character.
var sparkChars = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// spark returns a sparkline character for the given percentage (0-100).
func spark(pct float64) string {
	if pct <= 0 {
		return string(sparkChars[0])
	}
	if pct >= 100 {
		return string(sparkChars[len(sparkChars)-1])
	}
	idx := int(pct / 100.0 * float64(len(sparkChars)-1))
	return string(sparkChars[idx])
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Show top resource-consuming processes",
	Long: `Show the top N processes sorted by CPU usage.
Use --battery to highlight battery-draining processes.
Use --format spark to show inline sparkline bars for CPU and memory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		procs, err := process.Top(topN)
		if err != nil {
			return fmt.Errorf("failed to get top processes: %w", err)
		}

		if jsonFlag {
			return printJSON(procs)
		}

		if topFormat == "spark" {
			return printSparkTable(procs)
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
	topCmd.Flags().StringVar(&topFormat, "format", "", "Output format: spark (inline sparklines)")
	rootCmd.AddCommand(topCmd)
}

func printSparkTable(procs []process.Info) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tNAME\tCPU\tMEM\tCOMMAND")
	for _, p := range procs {
		fmt.Fprintf(w, "%d\t%s\t%s %.1f%%\t%s %.1f%%\t%s\n",
			p.PID, p.Name,
			spark(p.CPU), p.CPU,
			spark(p.Mem), p.Mem,
			p.Command)
	}
	w.Flush()
	return nil
}
