package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/zhengda-lu/pstop/internal/process"
)

var (
	killForce  bool
	killSignal string
)

var killCmd = &cobra.Command{
	Use:   "kill <pid>",
	Short: "Kill a process by PID",
	Long:  `Send a signal to a process. Defaults to SIGTERM. Use --force for SIGKILL or --signal for a specific signal.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %w", err)
		}

		if killSignal != "" {
			sig, err := parseSignal(killSignal)
			if err != nil {
				return err
			}
			if err := process.KillWithSignal(pid, sig); err != nil {
				return fmt.Errorf("failed to kill process: %w", err)
			}
			fmt.Printf("Sent %s to PID %d\n", killSignal, pid)
			return nil
		}

		if err := process.Kill(pid, killForce); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}

		if killForce {
			fmt.Printf("Sent SIGKILL to PID %d\n", pid)
		} else {
			fmt.Printf("Sent SIGTERM to PID %d\n", pid)
		}
		return nil
	},
}

func init() {
	killCmd.Flags().BoolVarP(&killForce, "force", "f", false, "Send SIGKILL instead of SIGTERM")
	killCmd.Flags().StringVar(&killSignal, "signal", "", "Signal to send (e.g., SIGTERM, SIGKILL, SIGHUP)")
	rootCmd.AddCommand(killCmd)
}

func parseSignal(s string) (syscall.Signal, error) {
	s = strings.ToUpper(strings.TrimPrefix(strings.ToUpper(s), "SIG"))
	signals := map[string]syscall.Signal{
		"HUP":  syscall.SIGHUP,
		"INT":  syscall.SIGINT,
		"QUIT": syscall.SIGQUIT,
		"KILL": syscall.SIGKILL,
		"TERM": syscall.SIGTERM,
		"USR1": syscall.SIGUSR1,
		"USR2": syscall.SIGUSR2,
		"STOP": syscall.SIGSTOP,
		"CONT": syscall.SIGCONT,
	}
	sig, ok := signals[s]
	if !ok {
		return 0, fmt.Errorf("unknown signal: %s", s)
	}
	return sig, nil
}
