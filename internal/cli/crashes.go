package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var (
	crashesLast    string
	crashesProcess string
)

var crashesCmd = &cobra.Command{
	Use:   "crashes",
	Short: "Show recent crash reports and app hangs",
	Long: `List recent crash reports, app hangs, spin reports, and kernel panics
from macOS DiagnosticReports.

Scans ~/Library/Logs/DiagnosticReports/ and /Library/Logs/DiagnosticReports/
for .ips (crash), .hang (app hang), .spin (spin), and .panic (kernel panic) files.

Examples:
  pstop crashes                      # List crashes from last 7 days
  pstop crashes --last 24h           # Last 24 hours
  pstop crashes --last 30d           # Last 30 days
  pstop crashes --process Safari     # Filter by process name
  pstop crashes info <path>          # Show details of a specific report
  pstop crashes --json               # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		reports, err := process.ListCrashReports(crashesLast, crashesProcess)
		if err != nil {
			return fmt.Errorf("failed to list crash reports: %w", err)
		}

		if jsonFlag {
			if len(reports) == 0 {
				return printJSON([]process.CrashReport{})
			}
			return printJSON(reports)
		}

		if len(reports) == 0 {
			if crashesProcess != "" {
				fmt.Printf("No crash reports found for %q in the last %s\n", crashesProcess, crashesLast)
			} else {
				fmt.Printf("No crash reports found in the last %s\n", crashesLast)
			}
			return nil
		}

		printCrashTable(reports)
		return nil
	},
}

var crashesInfoCmd = &cobra.Command{
	Use:   "info <report-path>",
	Short: "Show details of a specific crash report",
	Long:  `Display detailed information about a specific crash report file.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		detail, err := process.GetCrashDetail(args[0])
		if err != nil {
			return fmt.Errorf("failed to read crash report: %w", err)
		}

		if jsonFlag {
			return printJSON(detail)
		}

		printCrashDetail(detail)
		return nil
	},
}

func init() {
	crashesCmd.Flags().StringVar(&crashesLast, "last", "7d", "Time window (e.g., 24h, 7d, 30d)")
	crashesCmd.Flags().StringVar(&crashesProcess, "process", "", "Filter by process name")
	crashesCmd.AddCommand(crashesInfoCmd)
	rootCmd.AddCommand(crashesCmd)
}

func printCrashTable(reports []process.CrashReport) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "TIMESTAMP\tPROCESS\tTYPE\tSIGNAL\tPATH")
	for _, r := range reports {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			r.Timestamp, r.Process, r.ReportType, r.Signal, r.Path)
	}
	w.Flush()
}

func printCrashDetail(d *process.CrashDetail) {
	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "Process:\t%s\n", d.Process)
	fmt.Fprintf(w, "PID:\t%d\n", d.PID)
	fmt.Fprintf(w, "Type:\t%s\n", d.ReportType)
	fmt.Fprintf(w, "Timestamp:\t%s\n", d.Timestamp)

	if d.ExceptType != "" {
		fmt.Fprintf(w, "Exception:\t%s\n", d.ExceptType)
	}
	if d.Signal != "" {
		fmt.Fprintf(w, "Signal:\t%s\n", d.Signal)
	}
	if d.OSVersion != "" {
		fmt.Fprintf(w, "OS Version:\t%s\n", d.OSVersion)
	}
	if d.Version != "" {
		fmt.Fprintf(w, "App Version:\t%s\n", d.Version)
	}
	if d.ReportType == "crash" {
		fmt.Fprintf(w, "Crash Thread:\t%d\n", d.CrashThread)
	}
	fmt.Fprintf(w, "Path:\t%s\n", d.Path)
	w.Flush()

	if len(d.Backtrace) > 0 {
		fmt.Println("\nBacktrace (faulting thread):")
		for i, frame := range d.Backtrace {
			fmt.Printf("  %2d: %s\n", i, frame)
		}
	}
}
