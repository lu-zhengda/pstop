package cli

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var (
	watchInterval int
	watchAlert    bool
	watchCPU      float64
	watchMem      float64
)

// Alert holds information about a threshold violation.
type Alert struct {
	Timestamp string       `json:"timestamp"`
	Threshold string       `json:"threshold"`
	Value     float64      `json:"value"`
	Limit     float64      `json:"limit"`
	Process   process.Info `json:"process"`
}

var watchCmd = &cobra.Command{
	Use:   "watch [pid]",
	Short: "Live-monitor a process or watch for threshold alerts",
	Long: `Watch a process in real-time, refreshing at the specified interval.

With --alert, monitor all processes and exit with code 1 when any process
exceeds the specified CPU or memory threshold.

Examples:
  pstop watch 1234                        # Watch a single process
  pstop watch --alert --cpu 80            # Alert when any process exceeds 80% CPU
  pstop watch --alert --cpu 80 --mem 90   # Alert on CPU > 80% or memory > 90%
  pstop watch --alert --mem 50 --json     # Output structured alert data`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if watchAlert {
			return runAlertMode()
		}

		if len(args) == 0 {
			return fmt.Errorf("PID argument required (or use --alert for threshold monitoring)")
		}

		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %w", err)
		}

		if jsonFlag {
			info, err := process.GetInfo(pid)
			if err != nil {
				return fmt.Errorf("process %d not found or inaccessible: %w", pid, err)
			}
			return printJSON(info)
		}

		return runWatchPID(pid)
	},
}

func init() {
	watchCmd.Flags().IntVar(&watchInterval, "interval", 2, "Refresh interval in seconds")
	watchCmd.Flags().BoolVar(&watchAlert, "alert", false, "Monitor all processes for threshold violations")
	watchCmd.Flags().Float64Var(&watchCPU, "cpu", 0, "CPU threshold percentage (used with --alert)")
	watchCmd.Flags().Float64Var(&watchMem, "mem", 0, "Memory threshold percentage (used with --alert)")
	rootCmd.AddCommand(watchCmd)
}

// runWatchPID is the original single-PID watch mode.
func runWatchPID(pid int) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
	defer ticker.Stop()

	fmt.Printf("Watching PID %d (interval: %ds). Press Ctrl+C to stop.\n\n", pid, watchInterval)

	// Print immediately, then on each tick.
	if err := printWatchInfo(pid); err != nil {
		return err
	}

	for {
		select {
		case <-sigCh:
			fmt.Println("\nStopped watching.")
			return nil
		case <-ticker.C:
			// Clear screen with ANSI escape.
			fmt.Print("\033[H\033[2J")
			fmt.Printf("Watching PID %d (interval: %ds). Press Ctrl+C to stop.\n\n", pid, watchInterval)
			if err := printWatchInfo(pid); err != nil {
				return err
			}
		}
	}
}

// runAlertMode monitors all processes and exits when a threshold is exceeded.
func runAlertMode() error {
	if watchCPU <= 0 && watchMem <= 0 {
		return fmt.Errorf("--alert requires at least one of --cpu or --mem to be set")
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(watchInterval) * time.Second)
	defer ticker.Stop()

	if !jsonFlag {
		thresholds := ""
		if watchCPU > 0 {
			thresholds += fmt.Sprintf("CPU > %.1f%%", watchCPU)
		}
		if watchMem > 0 {
			if thresholds != "" {
				thresholds += ", "
			}
			thresholds += fmt.Sprintf("MEM > %.1f%%", watchMem)
		}
		fmt.Printf("Watching for alerts (%s, interval: %ds). Press Ctrl+C to stop.\n", thresholds, watchInterval)
	}

	// Check immediately, then on each tick.
	if alert := checkThresholds(); alert != nil {
		return reportAlert(alert)
	}

	for {
		select {
		case <-sigCh:
			if !jsonFlag {
				fmt.Println("\nStopped watching.")
			}
			return nil
		case <-ticker.C:
			if alert := checkThresholds(); alert != nil {
				return reportAlert(alert)
			}
		}
	}
}

// checkThresholds scans all processes and returns an Alert if any exceed thresholds.
func checkThresholds() *Alert {
	procs, err := process.List()
	if err != nil {
		return nil
	}

	for _, p := range procs {
		if watchCPU > 0 && p.CPU > watchCPU {
			return &Alert{
				Timestamp: time.Now().Format(time.RFC3339),
				Threshold: "cpu",
				Value:     p.CPU,
				Limit:     watchCPU,
				Process:   p,
			}
		}
		if watchMem > 0 && p.Mem > watchMem {
			return &Alert{
				Timestamp: time.Now().Format(time.RFC3339),
				Threshold: "mem",
				Value:     p.Mem,
				Limit:     watchMem,
				Process:   p,
			}
		}
	}
	return nil
}

// reportAlert outputs the alert and returns an error to trigger exit code 1.
func reportAlert(alert *Alert) error {
	if jsonFlag {
		if err := printJSON(alert); err != nil {
			return fmt.Errorf("failed to output alert: %w", err)
		}
		return fmt.Errorf("threshold exceeded")
	}

	fmt.Printf("\nALERT: %s threshold exceeded!\n", alert.Threshold)
	fmt.Printf("  Process: %s (PID %d)\n", alert.Process.Name, alert.Process.PID)
	fmt.Printf("  %s: %.1f%% (limit: %.1f%%)\n", alert.Threshold, alert.Value, alert.Limit)
	fmt.Printf("  Time: %s\n", alert.Timestamp)
	return fmt.Errorf("threshold exceeded")
}

func printWatchInfo(pid int) error {
	info, err := process.GetInfo(pid)
	if err != nil {
		return fmt.Errorf("process %d not found or inaccessible: %w", pid, err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintf(w, "PID:\t%d\n", info.PID)
	fmt.Fprintf(w, "Name:\t%s\n", info.Name)
	fmt.Fprintf(w, "User:\t%s\n", info.User)
	fmt.Fprintf(w, "CPU:\t%.1f%%\n", info.CPU)
	fmt.Fprintf(w, "Memory:\t%.1f%%\n", info.Mem)
	fmt.Fprintf(w, "Open Files:\t%d\n", info.OpenFiles)

	if len(info.Ports) > 0 {
		fmt.Fprintf(w, "Ports:\t")
		for i, p := range info.Ports {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%d", p)
		}
		fmt.Fprintln(w)
	}

	if len(info.Children) > 0 {
		fmt.Fprintf(w, "Children:\t")
		for i, c := range info.Children {
			if i > 0 {
				fmt.Fprintf(w, ", ")
			}
			fmt.Fprintf(w, "%d", c)
		}
		fmt.Fprintln(w)
	}

	w.Flush()
	fmt.Printf("\nLast updated: %s\n", time.Now().Format("15:04:05"))
	return nil
}
