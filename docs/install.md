# Installation Guide

- [Install](#install)
  - [Homebrew (macOS / Linux)](#homebrew-macos--linux)
  - [One-line installer (macOS / Linux)](#one-line-installer-macos--linux)
  - [Direct binary download (any OS)](#direct-binary-download-any-os)
  - [Docker](#docker)
  - [go install (Go developers / head-of-main)](#go-install-go-developers--head-of-main)
- [Configuration](#configuration)
  - [Discovery order](#discovery-order)
  - [Minimal config](#minimal-config)
  - [Full config reference](#full-config-reference)
  - [MCP client setup](#mcp-client-setup)
- [Multi-PC setup](#multi-pc-setup)
- [Run as a background service](#run-as-a-background-service)
  - [macOS — launchd](#macos--launchd)
  - [Linux — systemd user unit](#linux--systemd-user-unit)
  - [Windows — NSSM or Task Scheduler](#windows--nssm-or-task-scheduler)
- [Troubleshooting](#troubleshooting)

---

## Install

### Homebrew (macOS / Linux)

```bash
brew install robertstevens/tap/wiki-mcp
```

Installs the latest release binary. Upgrades via `brew upgrade wiki-mcp`.

---

### One-line installer (macOS / Linux)

```bash
curl -sSfL https://github.com/robertstevens/wiki-mcp/releases/latest/download/install.sh | sh
```

The script:

1. Detects OS and architecture.
2. Downloads the matching tarball from the latest GitHub release.
3. Verifies the SHA-256 checksum.
4. Installs to `/usr/local/bin/wiki-mcp` (falls back to `~/.local/bin/wiki-mcp` if `/usr/local/bin` is not writable).

Pin a specific version:

```bash
export WIKI_MCP_VERSION=v1.2.3
curl -sSfL https://github.com/robertstevens/wiki-mcp/releases/latest/download/install.sh | sh
```

---

### Direct binary download (any OS)

Download the release tarball that matches your platform from
`https://github.com/robertstevens/wiki-mcp/releases/latest`, then extract and install:

**macOS / Linux**

```bash
VERSION=v1.0.0   # replace with desired version
OS=$(uname -s)   # Darwin or Linux
ARCH=$(uname -m) # x86_64 or arm64
curl -L "https://github.com/robertstevens/wiki-mcp/releases/download/${VERSION}/wiki-mcp_${VERSION#v}_${OS}_${ARCH}.tar.gz" \
  -o wiki-mcp.tar.gz
tar -xzf wiki-mcp.tar.gz wiki-mcp
chmod +x wiki-mcp
mv wiki-mcp /usr/local/bin/wiki-mcp
```

**Windows (PowerShell)**

```powershell
$version = "1.0.0"   # without leading v
Invoke-WebRequest -Uri "https://github.com/robertstevens/wiki-mcp/releases/download/v$version/wiki-mcp_${version}_Windows_x86_64.zip" `
  -OutFile wiki-mcp.zip
Expand-Archive wiki-mcp.zip -DestinationPath "$Env:LOCALAPPDATA\wiki-mcp"
$env:PATH += ";$Env:LOCALAPPDATA\wiki-mcp"
[System.Environment]::SetEnvironmentVariable("PATH", $Env:PATH, "User")
```

---

### Docker

```bash
docker pull ghcr.io/robertstevens/wiki-mcp:latest
```

Run the web UI with your wiki directory mounted:

```bash
docker run --rm \
  -v /path/to/your/wiki:/wiki \
  -e WIKI_MCP_WIKI_PATH=/wiki \
  -p 9000:9000 \
  ghcr.io/robertstevens/wiki-mcp:latest \
  --serve-only
```

For MCP stdio transport, pipe stdin/stdout through `docker run`:

```bash
docker run -i --rm \
  -v /path/to/your/wiki:/wiki \
  -e WIKI_MCP_WIKI_PATH=/wiki \
  ghcr.io/robertstevens/wiki-mcp:latest
```

The image is multi-arch (`linux/amd64`, `linux/arm64`). Pin a release tag in production:

```bash
docker pull ghcr.io/robertstevens/wiki-mcp:v1.0.0
```

---

### go install (Go developers / head-of-main)

```bash
go install github.com/robertstevens/wiki-mcp/cmd/wiki-mcp@latest
```

Installs the latest commit on `main`. For the latest release:

```bash
go install github.com/robertstevens/wiki-mcp/cmd/wiki-mcp@v1.0.0
```

Requires Go 1.25+. The binary lands in `$(go env GOPATH)/bin/` (typically `~/go/bin/`). Ensure that directory is on your `$PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Configuration

### Discovery order

wiki-mcp looks for its config file in this order (first match wins):

1. `--config <path>` CLI flag
2. `WIKI_MCP_CONFIG` environment variable
3. `$XDG_CONFIG_HOME/wiki-mcp/config.toml` (falls back to `~/.config/wiki-mcp/config.toml`)
4. `./wiki-mcp.toml` in the current working directory
5. Built-in defaults — **`wiki_path` has no default and must be set** or the server exits with an error.

Individual fields can also be set via environment variables:

| Env var              | Config key       |
|----------------------|------------------|
| `WIKI_MCP_WIKI_PATH` | `wiki_path`      |
| `WIKI_MCP_WEB_PORT`  | `web.port`       |
| `WIKI_MCP_CONFIG`    | *(path override)*|

CLI flags and env vars take precedence over the config file.

---

### Minimal config

Create `~/.config/wiki-mcp/config.toml` (or `./wiki-mcp.toml`):

```toml
wiki_path = "/home/yourname/Documents/wiki"
```

That is the only required field.

---

### Full config reference

```toml
# Required. Absolute path to the wiki root directory.
wiki_path = "/home/yourname/Documents/wiki"

# Optional. Where source files live. Defaults to <wiki_path>/../sources.
sources_path = "/home/yourname/Documents/sources"

[web]
# Built-in read-only web server.
enabled = true            # default false
port = 9000               # default 9000
bind = "127.0.0.1"        # default localhost-only
theme = "default"         # "default" | "minimal" | path to custom template dir
auto_rebuild = true       # default true — re-renders on file change

[index]
# Section order in index.md. Custom sections appended in first-seen order.
sections = [
  { key = "research",       title = "🔬 Research" },
  { key = "entities",       title = "🏷️ Entities" },
  { key = "concepts",       title = "💡 Concepts" },
  { key = "infrastructure", title = "🏗️ Infrastructure" },
]

[log]
# Log entry date format. Default ISO 8601.
date_format = "%Y-%m-%d"

[links]
# How tools write links. Reading always accepts both styles.
# "obsidian" -> [[Title]]  |  "markdown" -> [Title](path)  |  "preserve" -> match existing page style
style = "preserve"

[safety]
read_only = false             # If true, all mutating tools refuse.
confine_to_wiki_path = true   # Reject paths outside wiki_path / sources_path.
max_page_bytes = 1048576      # 1 MiB default.
```

---

### MCP client setup

For per-client configuration examples (Claude Desktop, Claude Code, Cursor, Cline, etc.) see:

- [docs/clients/claude-desktop.md](clients/claude-desktop.md)
- [docs/clients/claude-code.md](clients/claude-code.md)

All clients use the same pattern: set `command` to the absolute path of the `wiki-mcp` binary and pass `WIKI_MCP_WIKI_PATH` as an env var. Use `which wiki-mcp` to get the correct binary path on your machine.

---

## Multi-PC setup

wiki-mcp does not sync files between machines — that is your sync tool's job. The pattern is:

1. **Sync the wiki directory** across all your machines (git, Syncthing, iCloud, Dropbox, rsync — your choice).
2. **Install the wiki-mcp binary** on each machine independently (any install method above).
3. **Point each install at the local sync directory** via config or env.

### Worked example — git-based sync

**One-time setup (machine A)**

```bash
mkdir -p ~/Documents/wiki
cd ~/Documents/wiki
git init
echo "# wiki" > index.md
echo "# log" > log.md
git add .
git commit -m "init"
git remote add origin git@github.com:yourname/private-wiki.git
git push -u origin main
```

**On each subsequent machine**

```bash
git clone git@github.com:yourname/private-wiki.git ~/Documents/wiki
brew install robertstevens/tap/wiki-mcp
mkdir -p ~/.config/wiki-mcp
cat > ~/.config/wiki-mcp/config.toml <<'EOF'
wiki_path = "/Users/yourname/Documents/wiki"
EOF
```

**Daily workflow**

```bash
# pull before a session
cd ~/Documents/wiki && git pull

# push after a session
cd ~/Documents/wiki && git add -A && git commit -m "wiki update" && git push
```

**Notes:**

- wiki-mcp does not automatically commit or push. Writes land immediately on disk; git operations are manual (or automated via cron / launch agent).
- If two machines write concurrently through different sync tools, conflicts are the sync tool's problem. For git, resolve merge conflicts the usual way.
- Syncthing and iCloud work the same way — install the binary locally, point at the synced dir.

---

## Run as a background service

Two flags control the web UI:

- `--serve` — runs both the MCP stdio transport and the web UI concurrently.
- `--serve-only` — runs only the web UI, no MCP stdio. Use this for a background service when no MCP client is attached.

The snippets below use `--serve-only`.

### macOS — launchd

First, find the binary path: `which wiki-mcp`. On Apple Silicon with Homebrew this is `/opt/homebrew/bin/wiki-mcp`; on Intel it is `/usr/local/bin/wiki-mcp`.

Create `~/Library/LaunchAgents/com.wiki-mcp.serve.plist`, replacing the binary path and wiki path as appropriate:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.wiki-mcp.serve</string>

  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/wiki-mcp</string>
    <string>--serve-only</string>
  </array>

  <key>EnvironmentVariables</key>
  <dict>
    <key>WIKI_MCP_WIKI_PATH</key>
    <string>/Users/yourname/Documents/wiki</string>
  </dict>

  <key>RunAtLoad</key>
  <true/>

  <key>KeepAlive</key>
  <true/>

  <key>StandardOutPath</key>
  <string>/tmp/wiki-mcp.log</string>
  <key>StandardErrorPath</key>
  <string>/tmp/wiki-mcp.err</string>
</dict>
</plist>
```

Load and start:

```bash
launchctl load ~/Library/LaunchAgents/com.wiki-mcp.serve.plist
launchctl start com.wiki-mcp.serve
```

Check status:

```bash
launchctl list | grep wiki-mcp
```

Stop and unload:

```bash
launchctl stop com.wiki-mcp.serve
launchctl unload ~/Library/LaunchAgents/com.wiki-mcp.serve.plist
```

---

### Linux — systemd user unit

Create `~/.config/systemd/user/wiki-mcp.service`:

```ini
[Unit]
Description=wiki-mcp web UI
After=network.target

[Service]
ExecStart=/usr/local/bin/wiki-mcp --serve-only
Environment=WIKI_MCP_WIKI_PATH=/home/yourname/Documents/wiki
Restart=on-failure
RestartSec=5

# Optional: log to journald (default) or redirect:
# StandardOutput=append:/home/yourname/.local/share/wiki-mcp/wiki-mcp.log
# StandardError=append:/home/yourname/.local/share/wiki-mcp/wiki-mcp.err

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user daemon-reload
systemctl --user enable wiki-mcp.service
systemctl --user start wiki-mcp.service
```

Check status and logs:

```bash
systemctl --user status wiki-mcp.service
journalctl --user -u wiki-mcp.service -f
```

Ensure lingering is enabled so the unit starts at boot without a login session:

```bash
loginctl enable-linger "$USER"
```

---

### Windows — NSSM or Task Scheduler

**Option A — NSSM (Non-Sucking Service Manager)**

1. Download NSSM from [nssm.cc](https://nssm.cc/download) and place `nssm.exe` somewhere on `%PATH%`.
2. Run in an elevated (Administrator) terminal:

```powershell
nssm install wiki-mcp "C:\Users\yourname\AppData\Local\wiki-mcp\wiki-mcp.exe"
nssm set wiki-mcp AppParameters "--serve-only"
nssm set wiki-mcp AppEnvironmentExtra "WIKI_MCP_WIKI_PATH=C:\Users\yourname\Documents\wiki"
nssm set wiki-mcp AppStdout "C:\Users\yourname\AppData\Local\wiki-mcp\wiki-mcp.log"
nssm set wiki-mcp AppStderr "C:\Users\yourname\AppData\Local\wiki-mcp\wiki-mcp.err"
nssm set wiki-mcp Start SERVICE_AUTO_START
nssm start wiki-mcp
```

Manage the service:

```powershell
nssm stop wiki-mcp
nssm restart wiki-mcp
nssm remove wiki-mcp confirm
```

**Option B — Task Scheduler (no extra tools)**

```powershell
$action  = New-ScheduledTaskAction `
  -Execute "C:\Users\yourname\AppData\Local\wiki-mcp\wiki-mcp.exe" `
  -Argument "--serve-only"

$trigger = New-ScheduledTaskTrigger -AtLogon

$settings = New-ScheduledTaskSettingsSet `
  -ExecutionTimeLimit ([TimeSpan]::Zero) `
  -RestartCount 3 `
  -RestartInterval (New-TimeSpan -Minutes 1)

$env = New-ScheduledTaskPrincipal -UserId "$Env:USERDOMAIN\$Env:USERNAME" -LogonType Interactive

Register-ScheduledTask `
  -TaskName "wiki-mcp" `
  -Action $action `
  -Trigger $trigger `
  -Settings $settings `
  -Principal $env `
  -Description "wiki-mcp web UI background service"
```

Run immediately without waiting for next login:

```powershell
Start-ScheduledTask -TaskName "wiki-mcp"
```

Set `WIKI_MCP_WIKI_PATH` as a persistent user environment variable so the task picks it up:

```powershell
[System.Environment]::SetEnvironmentVariable(
  "WIKI_MCP_WIKI_PATH",
  "C:\Users\yourname\Documents\wiki",
  "User"
)
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `wiki_path is required` on startup | `wiki_path` not set in config, env, or CLI flag | Set `WIKI_MCP_WIKI_PATH=/path/to/wiki` or add `wiki_path = "..."` to config file. See [Discovery order](#discovery-order). |
| `listen tcp :9000: bind: address already in use` | Another process holds port 9000 | Change the port: `wiki-mcp --port 9001` or `web.port = 9001` in config. Find the conflict: `lsof -i :9000` (macOS/Linux) or `netstat -ano | findstr :9000` (Windows). |
| MCP client shows no tools / "server not found" | Binary not on `$PATH`, wrong command in client config, or server crashed at startup | Verify: `which wiki-mcp && wiki-mcp --version`. Check client config `command` path is absolute. Capture stderr: wrap the command with `sh -c 'wiki-mcp 2>/tmp/wiki-mcp.err'` and inspect `/tmp/wiki-mcp.err`. |
| Windows Defender / antivirus blocks the binary | Unsigned binary from GitHub releases triggers heuristic scan | Add an exclusion for the binary path in Windows Security → Virus & threat protection → Exclusions. Or build from source: `go install github.com/robertstevens/wiki-mcp/cmd/wiki-mcp@latest`. |
| Docker container exits immediately | `wiki_path` not mounted, or mount path mismatch | Ensure `-v /host/wiki:/wiki` and `-e WIKI_MCP_WIKI_PATH=/wiki` are both set. The env var must match the *container-side* mount path. Run interactively to see the error: `docker run -it --rm -v /host/wiki:/wiki -e WIKI_MCP_WIKI_PATH=/wiki ghcr.io/robertstevens/wiki-mcp:latest`. |
| Docker: `permission denied` writing to wiki | Container process UID doesn't match volume owner | Run with matching UID: `docker run --user $(id -u):$(id -g) ...`. Or `chown -R $(id -u) /host/wiki` to make the host user the owner. |
