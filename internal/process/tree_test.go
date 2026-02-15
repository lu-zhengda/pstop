package process

import (
	"testing"
)

func TestBuildTree(t *testing.T) {
	procs := []Info{
		{PID: 1, PPID: 0, Name: "init"},
		{PID: 10, PPID: 1, Name: "parent"},
		{PID: 20, PPID: 10, Name: "child1"},
		{PID: 21, PPID: 10, Name: "child2"},
		{PID: 30, PPID: 20, Name: "grandchild"},
	}

	roots := BuildTree(procs)

	if len(roots) != 1 {
		t.Fatalf("BuildTree() returned %d roots, want 1", len(roots))
	}

	root := roots[0]
	if root.Process.Name != "init" {
		t.Errorf("root name = %q, want 'init'", root.Process.Name)
	}
	if len(root.Children) != 1 {
		t.Fatalf("root has %d children, want 1", len(root.Children))
	}

	parent := root.Children[0]
	if parent.Process.Name != "parent" {
		t.Errorf("parent name = %q, want 'parent'", parent.Process.Name)
	}
	if len(parent.Children) != 2 {
		t.Fatalf("parent has %d children, want 2", len(parent.Children))
	}

	child1 := parent.Children[0]
	if child1.Process.Name != "child1" {
		t.Errorf("child1 name = %q, want 'child1'", child1.Process.Name)
	}
	if len(child1.Children) != 1 {
		t.Fatalf("child1 has %d children, want 1", len(child1.Children))
	}

	grandchild := child1.Children[0]
	if grandchild.Process.Name != "grandchild" {
		t.Errorf("grandchild name = %q, want 'grandchild'", grandchild.Process.Name)
	}
}

func TestBuildTreeMultipleRoots(t *testing.T) {
	procs := []Info{
		{PID: 1, PPID: 0, Name: "init"},
		{PID: 2, PPID: 0, Name: "kthreadd"},
		{PID: 10, PPID: 1, Name: "child"},
	}

	roots := BuildTree(procs)
	if len(roots) != 2 {
		t.Fatalf("BuildTree() returned %d roots, want 2", len(roots))
	}
}

func TestBuildTreeEmpty(t *testing.T) {
	roots := BuildTree(nil)
	if len(roots) != 0 {
		t.Errorf("BuildTree(nil) returned %d roots, want 0", len(roots))
	}
}

func TestFlattenTree(t *testing.T) {
	procs := []Info{
		{PID: 1, PPID: 0, Name: "init"},
		{PID: 10, PPID: 1, Name: "parent"},
		{PID: 20, PPID: 10, Name: "child"},
	}

	roots := BuildTree(procs)
	flat := FlattenTree(roots)

	if len(flat) != 3 {
		t.Fatalf("FlattenTree() returned %d entries, want 3", len(flat))
	}

	expected := []struct {
		name  string
		depth int
	}{
		{"init", 0},
		{"parent", 1},
		{"child", 2},
	}

	for i, e := range expected {
		if flat[i].Process.Name != e.name {
			t.Errorf("flat[%d].Name = %q, want %q", i, flat[i].Process.Name, e.name)
		}
		if flat[i].Depth != e.depth {
			t.Errorf("flat[%d].Depth = %d, want %d", i, flat[i].Depth, e.depth)
		}
	}
}

func TestTreeIntegration(t *testing.T) {
	roots, err := Tree()
	if err != nil {
		t.Fatalf("Tree() error: %v", err)
	}
	if len(roots) == 0 {
		t.Error("Tree() returned no roots")
	}
}
