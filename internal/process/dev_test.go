package process

import (
	"testing"
)

func TestClassifyStack(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		// Node.js
		{"node", "Node.js"},
		{"npm", "Node.js"},
		{"npx", "Node.js"},
		{"yarn", "Node.js"},
		{"pnpm", "Node.js"},
		{"tsx", "Node.js"},
		{"ts-node", "Node.js"},
		{"next", "Node.js"},
		{"vite", "Node.js"},
		{"webpack", "Node.js"},
		{"esbuild", "Node.js"},
		{"bun", "Node.js"},
		{"deno", "Node.js"},

		// Python
		{"python", "Python"},
		{"python3", "Python"},
		{"pip", "Python"},
		{"pip3", "Python"},
		{"uvicorn", "Python"},
		{"gunicorn", "Python"},
		{"flask", "Python"},
		{"django", "Python"},
		{"celery", "Python"},
		{"jupyter", "Python"},
		{"ipython", "Python"},
		{"conda", "Python"},
		{"poetry", "Python"},
		{"uv", "Python"},

		// Docker
		{"docker", "Docker"},
		{"dockerd", "Docker"},
		{"containerd", "Docker"},
		{"docker-compose", "Docker"},
		{"com.docker.vmnetd", "Docker"},
		{"com.docker.backend", "Docker"},

		// Database
		{"postgres", "Database"},
		{"postgresql", "Database"},
		{"psql", "Database"},
		{"mysql", "Database"},
		{"mysqld", "Database"},
		{"mongod", "Database"},
		{"mongos", "Database"},
		{"redis-server", "Database"},
		{"redis-cli", "Database"},
		{"sqlite3", "Database"},

		// Java
		{"java", "Java"},
		{"javac", "Java"},
		{"gradle", "Java"},
		{"gradlew", "Java"},
		{"mvn", "Java"},
		{"maven", "Java"},
		{"kotlin", "Java"},
		{"sbt", "Java"},

		// Go
		{"go", "Go"},
		{"gopls", "Go"},
		{"dlv", "Go"},

		// Ruby
		{"ruby", "Ruby"},
		{"irb", "Ruby"},
		{"rails", "Ruby"},
		{"rake", "Ruby"},
		{"bundler", "Ruby"},
		{"gem", "Ruby"},
		{"puma", "Ruby"},
		{"sidekiq", "Ruby"},

		// Rust
		{"rustc", "Rust"},
		{"cargo", "Rust"},
		{"rustup", "Rust"},
		{"rust-analyzer", "Rust"},

		// Web Server
		{"nginx", "Web Server"},
		{"apache", "Web Server"},
		{"httpd", "Web Server"},
		{"caddy", "Web Server"},
		{"traefik", "Web Server"},

		// Unknown
		{"Finder", ""},
		{"Safari", ""},
		{"unknown_process", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyStack(tt.name)
			if got != tt.want {
				t.Errorf("ClassifyStack(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestClassifyStackCaseInsensitive(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Node", "Node.js"},
		{"PYTHON", "Python"},
		{"Docker", "Docker"},
		{"JAVA", "Java"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyStack(tt.name)
			if got != tt.want {
				t.Errorf("ClassifyStack(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestGroupProcessesByStack(t *testing.T) {
	procs := []Info{
		{PID: 1, Name: "node", CPU: 5.0, Mem: 2.0},
		{PID: 2, Name: "npm", CPU: 3.0, Mem: 1.0},
		{PID: 3, Name: "python3", CPU: 10.0, Mem: 5.0},
		{PID: 4, Name: "Finder", CPU: 1.0, Mem: 0.5},
		{PID: 5, Name: "redis-server", CPU: 2.0, Mem: 3.0},
	}

	groups := GroupProcessesByStack(procs)

	if len(groups) != 3 {
		t.Fatalf("GroupProcessesByStack() returned %d groups, want 3", len(groups))
	}

	// Verify Node.js group (should be first due to insertion order).
	nodeGroup := groups[0]
	if nodeGroup.Stack != "Node.js" {
		t.Errorf("first group stack = %q, want 'Node.js'", nodeGroup.Stack)
	}
	if len(nodeGroup.Processes) != 2 {
		t.Errorf("Node.js group has %d processes, want 2", len(nodeGroup.Processes))
	}
	if nodeGroup.TotalCPU != 8.0 {
		t.Errorf("Node.js TotalCPU = %f, want 8.0", nodeGroup.TotalCPU)
	}
	if nodeGroup.TotalMem != 3.0 {
		t.Errorf("Node.js TotalMem = %f, want 3.0", nodeGroup.TotalMem)
	}

	// Verify Python group.
	pythonGroup := groups[1]
	if pythonGroup.Stack != "Python" {
		t.Errorf("second group stack = %q, want 'Python'", pythonGroup.Stack)
	}
	if len(pythonGroup.Processes) != 1 {
		t.Errorf("Python group has %d processes, want 1", len(pythonGroup.Processes))
	}

	// Verify Database group.
	dbGroup := groups[2]
	if dbGroup.Stack != "Database" {
		t.Errorf("third group stack = %q, want 'Database'", dbGroup.Stack)
	}
}

func TestGroupProcessesByStackEmpty(t *testing.T) {
	groups := GroupProcessesByStack(nil)
	if len(groups) != 0 {
		t.Errorf("GroupProcessesByStack(nil) returned %d groups, want 0", len(groups))
	}
}

func TestGroupProcessesByStackNoDevProcesses(t *testing.T) {
	procs := []Info{
		{PID: 1, Name: "Finder", CPU: 1.0},
		{PID: 2, Name: "Safari", CPU: 2.0},
	}

	groups := GroupProcessesByStack(procs)
	if len(groups) != 0 {
		t.Errorf("GroupProcessesByStack(non-dev) returned %d groups, want 0", len(groups))
	}
}
