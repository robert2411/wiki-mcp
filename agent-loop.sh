#!/usr/bin/env bash
# agent-loop.sh
#
# Continuously picks the next To Do task from the backlog and runs Claude on it.
# If Claude hits the session rate limit, sleeps SLEEP_MINUTES and then resumes
# the same Claude session via `claude -p --continue` so the in-progress task
# keeps its context. Gives up after MAX_RATE_LIMIT_RETRIES consecutive hits.
# Exits cleanly when no open tasks remain, when the run-file is removed, or on
# Ctrl-C.
#
# Usage:
#   ./agent-loop.sh
#   SLEEP_MINUTES=15 ./agent-loop.sh
#   MAX_RATE_LIMIT_RETRIES=3 ./agent-loop.sh
#
# Stop the loop:
#   rm <script-dir>/agent-loop-run    (or Ctrl-C)

set -euo pipefail

SLEEP_MINUTES="${SLEEP_MINUTES:-30}"
SLEEP_SECONDS=$(( SLEEP_MINUTES * 60 ))
MAX_RATE_LIMIT_RETRIES="${MAX_RATE_LIMIT_RETRIES:-11}"

# Anchor the run-file to the script's directory so the loop is controllable
# regardless of where the user invokes it from.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_FILE="${SCRIPT_DIR}/agent-loop-run"

# Rate limit patterns — including the actual Claude CLI message
RATE_LIMIT_PATTERN="You've hit your limit|You're out of extra usage|rate limit|usage limit|session limit|too many requests|overloaded"

# Prompt used for a fresh Claude session (new task).
INITIAL_PROMPT='Take the next "To Do" task from the backlog and implement it.
Use the backlog MCP to read the task, set it in progress, write an implementation plan,
implement it following the acceptance criteria, mark each AC complete as you go,
add a final summary, and set the task status to Done when finished.
Run the skill: /simplify
Finally, commit the changes.'

# Prompt used when resuming after a rate-limit sleep — keep it short; the
# resumed session already has full context.
RESUME_PROMPT='Continue where you left off on the task you were working on before the session limit was hit. If it is already finished, move on to the next "To Do" task following the original instructions.'

log() { echo "[$(date '+%H:%M:%S')] $*"; }

# Temp file for capturing Claude output; cleaned up by the EXIT trap.
tmp=""
cleanup_tmp() { rm -f "${tmp:-}"; tmp=""; }

rate_limit_hits=0

# Create run-file and register cleanup on exit. INT/TERM exit cleanly so
# Ctrl-C during a long sleep removes the run-file and stops the loop.
touch "$RUN_FILE"
trap 'rm -f "$RUN_FILE" "${tmp:-}"; log "Run-file removed. Exiting."' EXIT
trap 'log "Interrupted."; exit 130' INT TERM

# Sleep that wakes up early if the run-file is removed, so stopping the loop
# via `rm agent-loop-run` takes effect within a few seconds.
interruptible_sleep() {
    local seconds="$1" elapsed=0
    while (( elapsed < seconds )); do
        [[ -f "$RUN_FILE" ]] || return 0
        sleep 15
        elapsed=$(( elapsed + 15 ))
    done
}

check_open_tasks() {
    local output
    if ! output=$(backlog task list -s "To Do" --plain 2>&1); then
        log "backlog task list failed:"
        log "$output"
        exit 1
    fi
    grep -qi "task-" <<<"$output"
}

# ── main loop ──────────────────────────────────────────────────────────────────

log "Agent loop started (run-file: $RUN_FILE, sleep on rate limit = ${SLEEP_MINUTES}m)"

while true; do

    # 1. Check run-file — stop if it has been removed externally
    if [[ ! -f "$RUN_FILE" ]]; then
        log "Run-file not found. Stopping loop."
        exit 0
    fi

    # 2. Check for open tasks
    if ! check_open_tasks; then
        log "No open 'To Do' tasks found — all done. Exiting."
        exit 0
    fi
    log "Open tasks found. Invoking Claude..."

    # 3. Run Claude; capture output so we can detect rate limit errors.
    #    First call in a task uses the full prompt. After a rate-limit sleep
    #    we use `--continue` to pick up the same session.
    tmp=$(mktemp)
    if (( rate_limit_hits == 0 )); then
        log "Starting new Claude session..."
        claude --dangerously-skip-permissions -p "$INITIAL_PROMPT" 2>&1 | tee "$tmp" || true
    else
        log "Resuming previous Claude session with --continue..."
        claude --dangerously-skip-permissions -p --continue "$RESUME_PROMPT" 2>&1 | tee "$tmp" || true
    fi
    exit_code=${PIPESTATUS[0]}

    # 4. Check if Claude hit a session/rate limit
    if grep -qiE "$RATE_LIMIT_PATTERN" "$tmp"; then
        rate_limit_hits=$(( rate_limit_hits + 1 ))
        if (( rate_limit_hits > MAX_RATE_LIMIT_RETRIES )); then
            log "Rate limit hit ${rate_limit_hits}x in a row (max=${MAX_RATE_LIMIT_RETRIES}). Giving up."
            cleanup_tmp
            exit 1
        fi
        log "Session limit hit (${rate_limit_hits}/${MAX_RATE_LIMIT_RETRIES}). Sleeping ${SLEEP_MINUTES}m, then resuming the same session..."
        cleanup_tmp
        interruptible_sleep "$SLEEP_SECONDS"
        continue
    fi

    # 5. Run completed without hitting the limit → reset for the next task
    cleanup_tmp
    rate_limit_hits=0
    log "Claude session complete (exit code: ${exit_code}). Rechecking backlog..."

done
