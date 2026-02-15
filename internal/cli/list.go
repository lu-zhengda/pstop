package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var (
	listSort string
	listUser string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all running processes",
	Long:  `List all running processes with optional sorting and user filtering.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		procs, err := process.List()
		if err != nil {
			return fmt.Errorf("failed to list processes: %w", err)
		}

		if listUser != "" {
			var filtered []process.Info
			for _, p := range procs {
				if p.User == listUser {
					filtered = append(filtered, p)
				}
			}
			procs = filtered
		}

		process.Sort(procs, listSort)

		if jsonFlag {
			return printJSON(procs)
		}

		printProcessTable(procs)
		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listSort, "sort", "cpu", "Sort by: cpu, mem, pid, name")
	listCmd.Flags().StringVar(&listUser, "user", "", "Filter by user")
	rootCmd.AddCommand(listCmd)
}

func printProcessTable(procs []process.Info) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tNAME\tUSER\tCPU%\tMEM%\tSTATE\tCOMMAND")
	for _, p := range procs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%.1f\t%.1f\t%s\t%s\n",
			p.PID, p.Name, p.User, p.CPU, p.Mem, p.State, p.Command)
	}
	w.Flush()
}
