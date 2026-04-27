package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/robert2411/wiki-mcp/internal/config"
)

// IndexEntry is a single bullet in an index section.
type IndexEntry struct {
	Title   string `json:"title"`
	Path    string `json:"path"`
	Summary string `json:"summary"`
}

// IndexSection is one emoji-prefixed section in index.md.
type IndexSection struct {
	Key        string       `json:"key"`
	Title      string       `json:"title"`
	HeaderLine string       `json:"-"` // raw "### 🔬 Research" line
	Entries    []IndexEntry `json:"entries"`
	Trailing   string       `json:"-"` // blank lines after last entry
}

// IndexStats holds parsed stats from the ## Stats block.
type IndexStats struct {
	SourcesIngested string   `json:"sources_ingested,omitempty"`
	WikiPages       string   `json:"wiki_pages,omitempty"`
	LastUpdated     string   `json:"last_updated,omitempty"`
	Extra           []string `json:"-"` // any unrecognised stats lines
}

// IndexDocument is the full parsed index.md.
type IndexDocument struct {
	Preamble          string         `json:"-"` // everything up to and including "## Pages\n"
	PreSectionContent string         `json:"-"` // comment + blanks before first ###
	Sections          []IndexSection `json:"sections"`
	InterBlock        string         `json:"-"` // separator between last section and stats
	Stats             IndexStats     `json:"stats"`
}

// entryRe matches "- [Title](path) — summary"
var entryRe = regexp.MustCompile(`^- \[([^\]]+)\]\(([^)]+)\) — (.+)$`)

// sectionHeaderRe matches "### <emoji> <title>"
var sectionHeaderRe = regexp.MustCompile(`^### .+$`)

// ParseIndex parses index.md content into an IndexDocument.
func ParseIndex(data []byte, cfg *config.Config) (*IndexDocument, error) {
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	doc := &IndexDocument{}

	// States: preamble -> preSections -> sections -> interBlock -> stats
	const (
		stPreamble = iota
		stPreSections
		stSections
		stInterBlock
		stStats
	)
	state := stPreamble

	var preambleLines []string
	var preSectionLines []string
	var interBlockLines []string
	var statsLines []string
	var curSection *IndexSection

	flushSection := func() {
		if curSection != nil {
			doc.Sections = append(doc.Sections, *curSection)
			curSection = nil
		}
	}

	for i, line := range lines {
		// Last element from Split on trailing \n is empty — skip
		if i == len(lines)-1 && line == "" {
			break
		}

		switch state {
		case stPreamble:
			preambleLines = append(preambleLines, line)
			if line == "## Pages" {
				state = stPreSections
			}

		case stPreSections:
			if sectionHeaderRe.MatchString(line) {
				state = stSections
				curSection = &IndexSection{HeaderLine: line, Title: parseSectionTitle(line)}
				curSection.Key = matchSectionKey(curSection.Title, cfg)
			} else {
				preSectionLines = append(preSectionLines, line)
			}

		case stSections:
			if line == "## Stats" {
				flushSection()
				state = stStats
				statsLines = append(statsLines, line)
			} else if line == "---" && !isInSection(lines, i) {
				// Separator between sections and stats
				flushSection()
				state = stInterBlock
				interBlockLines = append(interBlockLines, line)
			} else if sectionHeaderRe.MatchString(line) {
				flushSection()
				curSection = &IndexSection{HeaderLine: line, Title: parseSectionTitle(line)}
				curSection.Key = matchSectionKey(curSection.Title, cfg)
			} else if m := entryRe.FindStringSubmatch(line); m != nil {
				if curSection != nil {
					curSection.Entries = append(curSection.Entries, IndexEntry{
						Title:   m[1],
						Path:    m[2],
						Summary: m[3],
					})
				}
			} else {
				// Blank or other line — trailing content of current section
				if curSection != nil {
					curSection.Trailing += line + "\n"
				}
			}

		case stInterBlock:
			if line == "## Stats" {
				state = stStats
				statsLines = append(statsLines, line)
			} else {
				interBlockLines = append(interBlockLines, line)
			}

		case stStats:
			statsLines = append(statsLines, line)
		}
	}

	flushSection()

	doc.Preamble = strings.Join(preambleLines, "\n") + "\n"
	if len(preSectionLines) > 0 {
		doc.PreSectionContent = strings.Join(preSectionLines, "\n") + "\n"
	}
	if len(interBlockLines) > 0 {
		doc.InterBlock = strings.Join(interBlockLines, "\n") + "\n"
	}

	// Parse stats — skip the leading blank line (between "## Stats" and first bullet)
	// because RenderIndex hardcodes it; store trailing/middle blanks in Extra.
	statsHasContent := false
	for _, sl := range statsLines {
		if sl == "## Stats" {
			continue
		}
		trimmed := strings.TrimPrefix(sl, "- ")
		switch {
		case strings.HasPrefix(trimmed, "Sources ingested:"):
			doc.Stats.SourcesIngested = strings.TrimSpace(strings.TrimPrefix(trimmed, "Sources ingested:"))
			statsHasContent = true
		case strings.HasPrefix(trimmed, "Wiki pages:"):
			doc.Stats.WikiPages = strings.TrimSpace(strings.TrimPrefix(trimmed, "Wiki pages:"))
			statsHasContent = true
		case strings.HasPrefix(trimmed, "Last updated:"):
			doc.Stats.LastUpdated = strings.TrimSpace(strings.TrimPrefix(trimmed, "Last updated:"))
			statsHasContent = true
		case strings.TrimSpace(sl) == "":
			if statsHasContent {
				doc.Stats.Extra = append(doc.Stats.Extra, sl)
			}
			// else: leading blank — already rendered via hardcoded "\n" in RenderIndex
		default:
			doc.Stats.Extra = append(doc.Stats.Extra, sl)
			statsHasContent = true
		}
	}

	return doc, nil
}

// isInSection peeks ahead to see if a "---" is followed by a "## Stats" block.
// If so, it's a separator, not section content.
func isInSection(lines []string, idx int) bool {
	// Look ahead for ## Stats after possible blank lines
	for j := idx + 1; j < len(lines); j++ {
		l := strings.TrimSpace(lines[j])
		if l == "" {
			continue
		}
		if l == "## Stats" {
			return false
		}
		return true
	}
	return true
}

// parseSectionTitle extracts the title from "### 🔬 Research" -> "🔬 Research"
func parseSectionTitle(headerLine string) string {
	return strings.TrimPrefix(headerLine, "### ")
}

// matchSectionKey matches a section title against config sections to get the key.
func matchSectionKey(title string, cfg *config.Config) string {
	for _, s := range cfg.Index.Sections {
		if s.Title == title {
			return s.Key
		}
	}
	// Fallback: strip emoji prefix, lowercase
	parts := strings.SplitN(title, " ", 2)
	if len(parts) == 2 {
		return strings.ToLower(parts[1])
	}
	return strings.ToLower(title)
}

// RenderIndex renders an IndexDocument back to markdown.
func RenderIndex(doc *IndexDocument) []byte {
	var b strings.Builder

	b.WriteString(doc.Preamble)
	if doc.PreSectionContent != "" {
		b.WriteString(doc.PreSectionContent)
	}

	for _, sec := range doc.Sections {
		b.WriteString(sec.HeaderLine + "\n")
		for _, e := range sec.Entries {
			_, _ = fmt.Fprintf(&b, "- [%s](%s) — %s\n", e.Title, e.Path, e.Summary)
		}
		if sec.Trailing != "" {
			b.WriteString(sec.Trailing)
		}
	}

	if doc.InterBlock != "" {
		b.WriteString(doc.InterBlock)
	}

	// Stats
	b.WriteString("## Stats\n")
	b.WriteString("\n")
	if doc.Stats.SourcesIngested != "" {
		b.WriteString("- Sources ingested: " + doc.Stats.SourcesIngested + "\n")
	}
	if doc.Stats.WikiPages != "" {
		b.WriteString("- Wiki pages: " + doc.Stats.WikiPages + "\n")
	}
	if doc.Stats.LastUpdated != "" {
		b.WriteString("- Last updated: " + doc.Stats.LastUpdated + "\n")
	}
	for _, extra := range doc.Stats.Extra {
		b.WriteString(extra + "\n")
	}

	return []byte(b.String())
}

// IndexRead reads and parses index.md.
func IndexRead(cfg *config.Config) (*IndexDocument, *ToolError) {
	abs, err := cfg.ResolveWikiPath("index.md")
	if err != nil {
		return nil, NewToolError(ErrCodePathEscape, err.Error())
	}

	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, NewToolError(ErrCodeNotFound, "index.md not found")
		}
		return nil, NewToolError(ErrCodeInternal, err.Error())
	}

	doc, parseErr := ParseIndex(data, cfg)
	if parseErr != nil {
		return nil, NewToolError(ErrCodeInternal, parseErr.Error())
	}

	return doc, nil
}

// IndexUpsertEntry adds or updates a single entry in index.md.
func IndexUpsertEntry(cfg *config.Config, sectionKey, title, path, summary string) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	doc, te := IndexRead(cfg)
	if te != nil {
		return te
	}

	// Find existing section
	sectionIdx := -1
	for i, sec := range doc.Sections {
		if sec.Key == sectionKey {
			sectionIdx = i
			break
		}
	}

	if sectionIdx >= 0 {
		// Section exists — find entry by title+path
		sec := &doc.Sections[sectionIdx]
		entryIdx := -1
		for j, e := range sec.Entries {
			if e.Title == title && e.Path == path {
				entryIdx = j
				break
			}
		}
		if entryIdx >= 0 {
			sec.Entries[entryIdx].Summary = summary
		} else {
			sec.Entries = append(sec.Entries, IndexEntry{Title: title, Path: path, Summary: summary})
		}
	} else {
		// Create new section — insert in config order
		newSec := IndexSection{
			Key:     sectionKey,
			Entries: []IndexEntry{{Title: title, Path: path, Summary: summary}},
		}
		// Find title from config
		for _, cs := range cfg.Index.Sections {
			if cs.Key == sectionKey {
				newSec.Title = cs.Title
				newSec.HeaderLine = "### " + cs.Title
				break
			}
		}
		if newSec.HeaderLine == "" {
			return NewToolError(ErrCodeBadRequest, fmt.Sprintf("unknown section key %q", sectionKey))
		}
		newSec.Trailing = "\n"

		// Determine insertion position based on config order
		insertAt := len(doc.Sections)
		configOrder := configSectionOrder(cfg)
		newOrder, hasNew := configOrder[sectionKey]
		if hasNew {
			for i, sec := range doc.Sections {
				secOrder, ok := configOrder[sec.Key]
				if ok && secOrder > newOrder {
					insertAt = i
					break
				}
			}
		}

		// Insert at position
		doc.Sections = append(doc.Sections, IndexSection{})
		copy(doc.Sections[insertAt+1:], doc.Sections[insertAt:])
		doc.Sections[insertAt] = newSec
	}

	return writeIndex(cfg, doc)
}

// IndexRefreshStats recomputes page count and last-updated from disk.
func IndexRefreshStats(cfg *config.Config) *ToolError {
	if err := cfg.MustMutate(); err != nil {
		return NewToolError(ErrCodeReadOnly, err.Error())
	}

	doc, te := IndexRead(cfg)
	if te != nil {
		return te
	}

	// Count .md files (excluding index.md and log.md)
	count := 0
	root := cfg.Root()
	_ = filepath.WalkDir(root, func(abs string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(abs, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(root, abs)
		if rel == "index.md" || rel == "log.md" {
			return nil
		}
		count++
		return nil
	})

	doc.Stats.WikiPages = fmt.Sprintf("%d", count)
	doc.Stats.LastUpdated = time.Now().Format("2006-01-02")

	return writeIndex(cfg, doc)
}

func writeIndex(cfg *config.Config, doc *IndexDocument) *ToolError {
	abs, err := cfg.ResolveWikiPath("index.md")
	if err != nil {
		return NewToolError(ErrCodePathEscape, err.Error())
	}

	rendered := RenderIndex(doc)
	if err := os.WriteFile(abs, rendered, 0o644); err != nil {
		return NewToolError(ErrCodeInternal, err.Error())
	}
	return nil
}

func configSectionOrder(cfg *config.Config) map[string]int {
	m := make(map[string]int, len(cfg.Index.Sections))
	for i, s := range cfg.Index.Sections {
		m[s.Key] = i
	}
	return m
}
