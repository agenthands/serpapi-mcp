package main

import (
	"fmt"
	"os"

	_ "github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	fmt.Printf("serpapi-mcp %s (commit: %s, built: %s)\n", version, commit, date)
	// MCP server initialization coming in Phase 2
	os.Exit(0)
}
