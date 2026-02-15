package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/pstop/internal/process"
)

var watchInterval int

var watchCmd = &cobra.Command{
	Use:   "watch <pid>",
	Short: "Live-monitor a process",
	Long:  `Watch a process in real-time, refreshing at the specified interval.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %w", err)
		}

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
	},
}

func init() {
	watchCmd.Flags().IntVar(&watchInterval, "interval", 2, "Refresh interval in seconds")
	rootCmd.AddCommand(watchCmd)
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
