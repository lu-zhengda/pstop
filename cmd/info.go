package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var infoCmd = &cobra.Command{
	Use:   "info <pid>",
	Short: "Show detailed process information",
	Long:  `Display detailed information about a process including open files, ports, and children.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid PID: %w", err)
		}

		info, err := process.GetInfo(pid)
		if err != nil {
			return fmt.Errorf("failed to get process info: %w", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintf(w, "PID:\t%d\n", info.PID)
		fmt.Fprintf(w, "Name:\t%s\n", info.Name)
		fmt.Fprintf(w, "User:\t%s\n", info.User)
		fmt.Fprintf(w, "CPU:\t%.1f%%\n", info.CPU)
		fmt.Fprintf(w, "Memory:\t%.1f%%\n", info.Mem)
		fmt.Fprintf(w, "Open Files:\t%d\n", info.OpenFiles)

		if len(info.Ports) > 0 {
			ports := make([]string, len(info.Ports))
			for i, p := range info.Ports {
				ports[i] = strconv.Itoa(p)
			}
			fmt.Fprintf(w, "Ports:\t%s\n", strings.Join(ports, ", "))
		}

		if len(info.Children) > 0 {
			children := make([]string, len(info.Children))
			for i, c := range info.Children {
				children[i] = strconv.Itoa(c)
			}
			fmt.Fprintf(w, "Children:\t%s\n", strings.Join(children, ", "))
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
