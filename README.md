# wiki-mcp

Personal wiki server with MCP (Model Context Protocol) integration.

## Install

```bash
# Homebrew (macOS/Linux)
brew install robertstevens/tap/wiki-mcp

# One-line install (macOS/Linux)
curl -sSfL https://github.com/robertstevens/wiki-mcp/releases/latest/download/install.sh | sh

# Docker
docker pull ghcr.io/robertstevens/wiki-mcp:latest

# Go (head-of-main)
go install github.com/robertstevens/wiki-mcp/cmd/wiki-mcp@latest
```

Direct binary downloads and Windows instructions: see [docs/install.md](docs/install.md).

## Usage

```bash
wiki-mcp --wiki-path /path/to/wiki
```

See [docs/install.md](docs/install.md) for full installation options, configuration reference, multi-PC setup, and troubleshooting.
