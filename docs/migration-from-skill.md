# Migration: Andy skill-based wiki → wiki-mcp

This guide replaces the MkDocs/systemd stack and Andy's inline wiki skill with the
`wiki-mcp` MCP server. No wiki data is moved — wiki-mcp reads the same directory the
old skill wrote to.

**Prerequisites:** wiki-mcp installed on the host (see [install.md](install.md)).

---

## Step 1 — Point wiki-mcp at the existing wiki directory

No data migration is needed. wiki-mcp reads the wiki directory in place.

Create `~/.config/wiki-mcp/config.toml`:

```bash
mkdir -p ~/.config/wiki-mcp
cat > ~/.config/wiki-mcp/config.toml <<'EOF'
wiki_path = "/home/robert/NanoClaw/groups/telegram_main/wiki"
EOF
```

Verify the server starts and can read the existing pages:

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1"}}}' \
  | wiki-mcp 2>/dev/null
```

You should see a JSON response with `"name":"wiki-mcp"`. If wiki-mcp exits with
`wiki_path is required`, the config was not picked up — check `WIKI_MCP_WIKI_PATH` or
the `--config` flag.

---

## Step 2 — Disable the old MkDocs/systemd stack

Stop and disable the timer and web server. The rebuild service is driven by the timer, so
disabling the timer is sufficient to prevent future rebuilds; stop the service only if a
rebuild is currently in progress.

```bash
systemctl --user disable --now \
  nanoclaw-wiki \
  nanoclaw-wiki-rebuild.timer
systemctl --user stop nanoclaw-wiki-rebuild.service
```

Confirm all three are no longer running:

```bash
systemctl --user status nanoclaw-wiki nanoclaw-wiki-rebuild.timer nanoclaw-wiki-rebuild.service
```

All should show `inactive (dead)`.

Port 9000 is now free. The MkDocs static site at `/tmp/nanoclaw-wiki-site` can be
removed:

```bash
rm -rf /tmp/nanoclaw-wiki-site
```

wiki-mcp's built-in web UI can replace the old read-only browser. Add the `[web]` block
to your config (see [install.md § Full config reference](install.md#full-config-reference)
for all options) and run wiki-mcp as a background service — see
[install.md § Run as a background service](install.md#run-as-a-background-service) for
the systemd user unit.

---

## Step 3 — Remove the wiki skill sections from Andy's CLAUDE.md

The MCP server's `ingest`, `query`, and `lint` prompts now carry the wiki workflow
semantics. The inline skill text in `CLAUDE.md` is redundant and should be removed to
avoid conflicts.

Open Andy's `CLAUDE.md` (host path: `/home/robert/NanoClaw/groups/main/CLAUDE.md` or
wherever `groups/main/` is kept) and remove the `## Wiki` section and its subsections
(`### Key Files`, `### Three Operations`, `### Ingest Discipline`, `### Source Handling`).

Replace the removed block with a short pointer so the history is clear:

```markdown
## Wiki

Wiki access is provided by the `wiki-mcp` MCP server — use the `ingest`, `query`, and
`lint` prompts. See the project README or `docs/migration-from-skill.md` for setup
details. The inline skill was removed when the MCP server was registered.
```

Do not remove other sections (`## What You Can Do`, scheduling, container mounts, etc.).

---

## Step 4 — Register wiki-mcp in the nanoclaw container

Andy runs as a Claude Code agent inside a container. To give Andy wiki access through
the MCP surface, add wiki-mcp to Claude Code's MCP config inside the container.

### 4a — Make the binary available inside the container

The simplest approach is to bind-mount the host binary. Find the host binary path:

```bash
which wiki-mcp   # e.g. /usr/local/bin/wiki-mcp or /home/robert/.local/bin/wiki-mcp
```

Add an additional mount to the container that runs Andy's group. In
`~/.config/nanoclaw/registered_groups.json` (or wherever nanoclaw stores container
config), add under the group's entry:

```json
"containerConfig": {
  "additionalMounts": [
    {
      "hostPath": "/usr/local/bin/wiki-mcp",
      "containerPath": "bin/wiki-mcp",
      "readonly": true
    }
  ]
}
```

nanoclaw mounts additional mounts under `/workspace/extra/` in the container, so
`"containerPath": "bin/wiki-mcp"` resolves to `/workspace/extra/bin/wiki-mcp`.

> **Alternative:** If the container image can be rebuilt, install wiki-mcp in the image
> with the one-line installer (see [install.md § One-line installer](install.md#one-line-installer-macos--linux)).
> This avoids the mount entirely — skip directly to Step 4b, using the image-internal
> binary path as `command`.

### 4b — Add the MCP server to Claude Code's config inside the container

Claude Code reads MCP server config from `~/.claude.json` inside the container. The
container home is typically `/root` or `/home/claude` depending on the image.

Edit `~/.claude.json` inside the container to add the `wiki` server. If other MCP
servers are already configured, merge this entry into the existing `mcpServers` object
rather than overwriting the file. Use `jq` to merge safely:

```bash
# inside the container — creates or merges ~/.claude.json
jq '.mcpServers.wiki = {
  "type": "stdio",
  "command": "/workspace/extra/bin/wiki-mcp",
  "args": [],
  "env": { "WIKI_MCP_WIKI_PATH": "/workspace/group/wiki" }
}' ~/.claude.json > /tmp/claude.json.tmp && mv /tmp/claude.json.tmp ~/.claude.json
```

If `~/.claude.json` does not exist yet, create it directly:

```json
{
  "mcpServers": {
    "wiki": {
      "type": "stdio",
      "command": "/workspace/extra/bin/wiki-mcp",
      "args": [],
      "env": {
        "WIKI_MCP_WIKI_PATH": "/workspace/group/wiki"
      }
    }
  }
}
```

`/workspace/group` maps to the host `groups/telegram_main/` directory, so
`/workspace/group/wiki` is the same directory the old skill wrote to.

### 4c — Verify inside the container

After restarting the container session, ask Andy:

```
@Andy list the wiki tools you have access to
```

Andy should enumerate the wiki-mcp tools (`page_read`, `page_write`, `ingest`, etc.).
Run a quick query to confirm end-to-end access:

```
@Andy query: what do we know about ollama?
```

---

## Step 5 — Rollback

If anything breaks, re-enable the old stack and restore CLAUDE.md.

### Re-enable the systemd units

```bash
systemctl --user enable --now \
  nanoclaw-wiki-rebuild.timer \
  nanoclaw-wiki
```

Force an immediate rebuild so the site is available right away:

```bash
systemctl --user start nanoclaw-wiki-rebuild.service
```

Confirm the site is back up:

```bash
curl -s http://localhost:9000 | head -5
```

### Restore Andy's CLAUDE.md wiki section

The original wiki skill text lives in `existing/skill.md` in this repository. Copy the
relevant sections back into Andy's `CLAUDE.md`:

```bash
# On the host, from the wiki-mcp project root:
cat existing/skill.md
```

Paste the content under a `## Wiki` heading in Andy's `CLAUDE.md`, replacing the
pointer note added in Step 3.

### Remove the MCP server registration

Undo Step 4: remove the `wiki` entry from `~/.claude.json` inside the container and
remove the `additionalMounts` entry from the container config.

---

## Verification checklist

After completing the migration, confirm:

- [ ] `wiki-mcp` starts without errors against the existing wiki directory (Step 1)
- [ ] All three systemd units are disabled and port 9000 is free (Step 2)
- [ ] Andy's `CLAUDE.md` no longer contains the old inline skill text (Step 3)
- [ ] Andy can invoke wiki tools via MCP inside the container (Step 4)
- [ ] A live ingest or query workflow succeeds end-to-end (Step 4c)
