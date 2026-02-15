package process

import "fmt"

// TreeNode represents a process and its children in a tree structure.
type TreeNode struct {
	Process  Info
	Children []*TreeNode
}

// Tree builds a process tree from all running processes.
func Tree() ([]*TreeNode, error) {
	procs, err := List()
	if err != nil {
		return nil, fmt.Errorf("failed to build process tree: %w", err)
	}
	return BuildTree(procs), nil
}

// BuildTree constructs a tree from a flat list of processes.
func BuildTree(procs []Info) []*TreeNode {
	nodes := make(map[int]*TreeNode, len(procs))
	for _, p := range procs {
		nodes[p.PID] = &TreeNode{Process: p}
	}

	var roots []*TreeNode
	for _, p := range procs {
		node := nodes[p.PID]
		parent, hasParent := nodes[p.PPID]
		if hasParent && p.PID != p.PPID {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	return roots
}

// Flatten returns a flat list of process info with indentation depth for rendering.
type FlatTreeEntry struct {
	Process Info
	Depth   int
}

// FlattenTree converts a tree into a flat list with depth information.
func FlattenTree(roots []*TreeNode) []FlatTreeEntry {
	var result []FlatTreeEntry
	for _, root := range roots {
		flattenNode(root, 0, &result)
	}
	return result
}

func flattenNode(node *TreeNode, depth int, result *[]FlatTreeEntry) {
	*result = append(*result, FlatTreeEntry{Process: node.Process, Depth: depth})
	for _, child := range node.Children {
		flattenNode(child, depth+1, result)
	}
}
