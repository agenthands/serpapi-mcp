# Installation

Install serpapi-mcp on Linux, macOS, or Windows.

## Prerequisites

| Method | Requirements |
|--------|-------------|
| Binary download | None — download and run |
| `go install` | Go 1.25+ |
| Build from source | Go 1.25+, Git |

All methods require a SerpApi API key from the [SerpApi dashboard](https://serpapi.com/dashboard).

## Binary download

Download a pre-built binary from [GitHub Releases](https://github.com/agenthands/serpapi-mcp/releases). This is the fastest method and requires no Go installation.

Replace `X.Y.Z` with the latest release version.

### Linux amd64

```bash
curl -L -o serpapi-mcp.tar.gz https://github.com/agenthands/serpapi-mcp/releases/download/vX.Y.Z/serpapi-mcp-vX.Y.Z-linux-amd64.tar.gz
tar xzf serpapi-mcp.tar.gz
chmod +x serpapi-mcp
```

### Linux arm64

```bash
curl -L -o serpapi-mcp.tar.gz https://github.com/agenthands/serpapi-mcp/releases/download/vX.Y.Z/serpapi-mcp-vX.Y.Z-linux-arm64.tar.gz
tar xzf serpapi-mcp.tar.gz
chmod +x serpapi-mcp
```

### macOS amd64 (Intel)

```bash
curl -L -o serpapi-mcp.tar.gz https://github.com/agenthands/serpapi-mcp/releases/download/vX.Y.Z/serpapi-mcp-vX.Y.Z-darwin-amd64.tar.gz
tar xzf serpapi-mcp.tar.gz
chmod +x serpapi-mcp
```

### macOS arm64 (Apple Silicon)

```bash
curl -L -o serpapi-mcp.tar.gz https://github.com/agenthands/serpapi-mcp/releases/download/vX.Y.Z/serpapi-mcp-vX.Y.Z-darwin-arm64.tar.gz
tar xzf serpapi-mcp.tar.gz
chmod +x serpapi-mcp
```

### Windows amd64

```powershell
Invoke-WebRequest -Uri "https://github.com/agenthands/serpapi-mcp/releases/download/vX.Y.Z/serpapi-mcp-vX.Y.Z-windows-amd64.zip" -OutFile "serpapi-mcp.zip"
Expand-Archive -Path serpapi-mcp.zip -DestinationPath .
```

The Windows executable is `serpapi-mcp.exe`.

### Verify installation

```bash
./serpapi-mcp --version
# Output: serpapi-mcp X.Y.Z (commit: abc1234, built: 2026-01-15T12:00:00Z)
```

### Upgrade

Download the new version from [GitHub Releases](https://github.com/agenthands/serpapi-mcp/releases) and replace the old binary.

## go install

Install the binary using Go's built-in tooling. Requires Go 1.25 or later.

```bash
go install github.com/agenthands/serpapi-mcp@latest
```

This downloads, compiles, and installs the binary to `$GOPATH/bin` (or `$HOME/go/bin` if `GOPATH` is unset).

### Add to PATH

If `$GOPATH/bin` is not in your `PATH`:

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Verify installation

```bash
serpapi-mcp --version
# Output: serpapi-mcp X.Y.Z (commit: abc1234, built: 2026-01-15T12:00:00Z)
```

### Upgrade

Re-run the install command with the desired version tag:

```bash
go install github.com/agenthands/serpapi-mcp@vX.Y.Z
```

## Build from source

### Clone the repository

```bash
git clone https://github.com/agenthands/serpapi-mcp.git
cd serpapi-mcp
```

### go build

Build the binary directly:

```bash
go build -o serpapi-mcp ./cmd/serpapi-mcp
```

The resulting binary reports `dev` as the version since no version information is injected via ldflags.

To embed version information:

```bash
go build -ldflags "-s -w -X main.version=X.Y.Z -X main.commit=$(git rev-parse --short HEAD) -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o serpapi-mcp ./cmd/serpapi-mcp
```

### goreleaser

Build cross-platform binaries using [goreleaser](https://goreleaser.dev/):

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser/v2@latest

# Local snapshot build (no Git tag required)
goreleaser build --snapshot --clean

# Full release build (requires a Git tag)
goreleaser release --clean
```

Snapshot builds place binaries in `dist/`. Release builds require a Git tag and produce all 5 platform archives.

### Verify build

```bash
./serpapi-mcp --version
```

### Upgrade

```bash
git pull
go build -o serpapi-mcp ./cmd/serpapi-mcp
```

Or with goreleaser:

```bash
git pull
goreleaser build --snapshot --clean
```

## Docker

Docker deployment is not supported. The server ships as static binaries with no runtime dependencies — use binary download or `go install` above.

## Next steps

- [USAGE.md](USAGE.md) — configuration, running the server, MCP client integration, and troubleshooting
- [README.md](README.md) — project overview and quickstart guide