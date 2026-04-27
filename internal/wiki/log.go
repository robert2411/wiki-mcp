package wiki

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/robert2411/wiki-mcp/internal/config"
)

var allowedLogOps = map[string]bool{
	"ingest": true,
	"query":  true,
	"lint":   true,
}

// logEntryHeaderRe matches "## [YYYY-MM-DD] <op> | <title>"
var logEntryHeaderRe = regexp.MustCompile(`^## \[(\d{4}-\d{2}-\d{2})\] (\w+) \| (.+)$`)

// logTemplate is written when log.md is missing.
const logTemplate = "# Wiki Log\n\nAppend-only chronological record of all wiki activity.\n" +
	"Format: `## [YYYY-MM-DD] <operation> | <title>`\n\n" +
	"Search recent entries: `grep \"^## \\[\" wiki/log.md | tail -10`\n\n---\n\n" +
	"<!-- entries below. Never delete entries. -->\n"

// LogEntry is a single parsed log entry.
type LogEntry struct {
	Date      string `json:"date"`
	Operation string `json:"operation"`
	Title     string `json:"title"`
	Body      string `json:"body"`
}

// strftimeToGo converts a strftime-style date format to Go time layout.
func strftimeToGo(f string) string {
	r := strings.NewReplacer("%Y", "2006", "%m", "01", "%d", "02")
	return r.Replace(f)
}

// LogAppend stamps the current date and appends a new entry to log.md.
func LogAppend(cfg *config.Config, operation, title, body string) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	if !allowedLogOps[operation] {
		return NewToolError(ErrCodeBadRequest,
			fmt.Sprintf("operation %q not allowed; must be one of: ingest, query, lint", operation))
	}

	abs, err := cfg.ResolveWikiPath("log.md")
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}

	existing, err := os.ReadFile(abs)
	if err != nil {
		if !os.IsNotExist(err) {
			return NewToolError(ErrCodeInternal, err.Error())
		}
		existing = []byte(logTemplate)
	}

	goFmt := strftimeToGo(cfg.Log.DateFormat)
	date := time.Now().Format(goFmt)

	header := fmt.Sprintf("## [%s] %s | %s", date, operation, title)

	base := strings.TrimRight(string(existing), "\n")
	var sb strings.Builder
	sb.WriteString(base)
	sb.WriteString("\n\n")
	sb.WriteString(header)
	sb.WriteString("\n")
	if body != "" {
		sb.WriteString("\n")
		sb.WriteString(strings.TrimRight(body, "\n"))
		sb.WriteString("\n")
	}

	if err := os.WriteFile(abs, []byte(sb.String()), 0o644); err != nil {
		return NewToolError(ErrCodeInternal, err.Error())
	}

	return nil
}

// LogTail returns the last n entries from log.md, newest last (chronological order).
func LogTail(cfg *config.Config, n int) ([]LogEntry, *ToolError) {
	if n <= 0 {
		n = 10
	}

	abs, err := cfg.ResolveWikiPath("log.md")
	if err != nil {
		return nil, NewToolError(ErrCodePathEscape, err.Error())
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return []LogEntry{}, nil
		}
		return nil, NewToolError(ErrCodeInternal, err.Error())
	}

	entries := parseLogEntries(string(data))
	if len(entries) > n {
		entries = entries[len(entries)-n:]
	}

	return entries, nil
}

// parseLogEntries splits log.md content into LogEntry values.
func parseLogEntries(content string) []LogEntry {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	var entries []LogEntry
	var cur *LogEntry
	var bodyLines []string

	flush := func() {
		if cur == nil {
			return
		}
		cur.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
		entries = append(entries, *cur)
		cur = nil
		bodyLines = nil
	}

	for _, line := range lines {
		if m := logEntryHeaderRe.FindStringSubmatch(line); m != nil {
			flush()
			cur = &LogEntry{
				Date:      m[1],
				Operation: m[2],
				Title:     m[3],
			}
			bodyLines = nil
		} else if cur != nil {
			bodyLines = append(bodyLines, line)
		}
	}
	flush()

	return entries
}
