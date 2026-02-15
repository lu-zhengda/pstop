package process

import (
	"fmt"
	"strings"
)

// DevGroup represents a group of processes belonging to a development stack.
type DevGroup struct {
	Stack     string
	Processes []Info
	TotalCPU  float64
	TotalMem  float64
}

// ClassifyStack classifies a process name into a development stack.
// Returns an empty string if the process does not match any known stack.
func ClassifyStack(name string) string {
	lower := strings.ToLower(name)

	switch {
	case lower == "node" || lower == "npm" || lower == "npx" ||
		lower == "yarn" || lower == "pnpm" || lower == "tsx" ||
		lower == "ts-node" || lower == "next" || lower == "vite" ||
		lower == "webpack" || lower == "esbuild" || lower == "bun" ||
		lower == "deno":
		return "Node.js"

	case lower == "python" || lower == "python3" || lower == "pip" ||
		lower == "pip3" || lower == "uvicorn" || lower == "gunicorn" ||
		lower == "flask" || lower == "django" || lower == "celery" ||
		lower == "jupyter" || lower == "ipython" || lower == "conda" ||
		lower == "poetry" || lower == "uv":
		return "Python"

	case lower == "docker" || lower == "dockerd" || lower == "containerd" ||
		lower == "docker-compose" || lower == "com.docker.vmnetd" ||
		lower == "com.docker.backend" || lower == "com.docker.hyperkit" ||
		strings.HasPrefix(lower, "docker"):
		return "Docker"

	case lower == "postgres" || lower == "postgresql" || lower == "psql" ||
		lower == "mysql" || lower == "mysqld" || lower == "mongod" ||
		lower == "mongos" || lower == "redis-server" || lower == "redis-cli" ||
		lower == "sqlite3":
		return "Database"

	case lower == "java" || lower == "javac" || lower == "gradle" ||
		lower == "gradlew" || lower == "mvn" || lower == "maven" ||
		lower == "kotlin" || lower == "kotlinc" || lower == "sbt":
		return "Java"

	case lower == "go" || lower == "gopls" || lower == "dlv" ||
		lower == "go-build" || lower == "gofmt":
		return "Go"

	case lower == "ruby" || lower == "irb" || lower == "rails" ||
		lower == "rake" || lower == "bundler" || lower == "gem" ||
		lower == "puma" || lower == "sidekiq":
		return "Ruby"

	case lower == "rustc" || lower == "cargo" || lower == "rustup" ||
		lower == "rust-analyzer":
		return "Rust"

	case lower == "nginx" || lower == "apache" || lower == "httpd" ||
		lower == "caddy" || lower == "traefik":
		return "Web Server"
	}

	return ""
}

// GroupByStack groups running processes by their development stack.
func GroupByStack() ([]DevGroup, error) {
	procs, err := List()
	if err != nil {
		return nil, fmt.Errorf("failed to group processes by stack: %w", err)
	}
	return GroupProcessesByStack(procs), nil
}

// GroupProcessesByStack groups a list of processes by their development stack.
func GroupProcessesByStack(procs []Info) []DevGroup {
	groups := make(map[string]*DevGroup)
	var order []string

	for _, p := range procs {
		stack := ClassifyStack(p.Name)
		if stack == "" {
			continue
		}

		g, exists := groups[stack]
		if !exists {
			g = &DevGroup{Stack: stack}
			groups[stack] = g
			order = append(order, stack)
		}
		g.Processes = append(g.Processes, p)
		g.TotalCPU += p.CPU
		g.TotalMem += p.Mem
	}

	result := make([]DevGroup, 0, len(order))
	for _, stack := range order {
		result = append(result, *groups[stack])
	}
	return result
}
