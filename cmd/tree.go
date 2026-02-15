package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/lu-zhengda/pstop/internal/process"
)

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Display process tree",
	Long:  `Show all processes in a tree structure based on parent-child relationships.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		roots, err := process.Tree()
		if err != nil {
			return fmt.Errorf("failed to build process tree: %w", err)
		}
		for _, root := range roots {
			printTreeNode(root, "", true)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(treeCmd)
}

func printTreeNode(node *process.TreeNode, prefix string, isLast bool) {
	connector := "├── "
	if isLast {
		connector = "└── "
	}
	if prefix == "" {
		connector = ""
	}

	fmt.Printf("%s%s%s (PID %d, CPU %.1f%%, %s)\n",
		prefix, connector,
		node.Process.Name, node.Process.PID,
		node.Process.CPU, node.Process.User)

	childPrefix := prefix
	if prefix != "" {
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
	}

	for i, child := range node.Children {
		isChildLast := i == len(node.Children)-1
		printTreeNode(child, childPrefix, isChildLast)
	}

	// Print a separator after root-level entries for readability.
	if prefix == "" && len(strings.TrimSpace(prefix)) == 0 {
		// No separator needed for root level
	}
}
