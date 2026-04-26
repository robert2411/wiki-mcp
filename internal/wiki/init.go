package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robertstevens/wiki-mcp/internal/config"
)

// InitResult reports what wiki_init created or skipped.
type InitResult struct {
	WikiPath string   `json:"wiki_path"`
	Created  []string `json:"created"`
	Skipped  []string `json:"skipped"`
}

// WikiInit bootstraps a wiki at cfg.Root(): creates the directory, section
// subdirectories, index.md, and log.md. Existing files are never overwritten.
func WikiInit(cfg *config.Config) (*InitResult, *ToolError) {
	if err := cfg.MustMutate(); err != nil {
		return nil, NewToolError(ErrCodeReadOnly, err.Error())
	}

	root := cfg.Root()
	result := &InitResult{WikiPath: root}

	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, NewToolError(ErrCodeInternal, fmt.Sprintf("cannot create wiki directory: %v", err))
	}

	for _, sec := range cfg.Index.Sections {
		dir := filepath.Join(root, sec.Key)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, 0o755); err != nil {
				return nil, NewToolError(ErrCodeInternal, fmt.Sprintf("cannot create %s/: %v", sec.Key, err))
			}
			result.Created = append(result.Created, sec.Key+"/")
		} else {
			result.Skipped = append(result.Skipped, sec.Key+"/")
		}
	}

	indexAbs := filepath.Join(root, "index.md")
	if _, err := os.Stat(indexAbs); os.IsNotExist(err) {
		rendered := RenderIndex(newInitIndexDocument(cfg))
		if err := os.WriteFile(indexAbs, rendered, 0o644); err != nil {
			return nil, NewToolError(ErrCodeInternal, fmt.Sprintf("cannot write index.md: %v", err))
		}
		result.Created = append(result.Created, "index.md")
	} else {
		result.Skipped = append(result.Skipped, "index.md")
	}

	logAbs := filepath.Join(root, "log.md")
	if _, err := os.Stat(logAbs); os.IsNotExist(err) {
		if err := os.WriteFile(logAbs, []byte(logTemplate), 0o644); err != nil {
			return nil, NewToolError(ErrCodeInternal, fmt.Sprintf("cannot write log.md: %v", err))
		}
		result.Created = append(result.Created, "log.md")
	} else {
		result.Skipped = append(result.Skipped, "log.md")
	}

	return result, nil
}

// ProjectInfo describes a project (sub-wiki) found under the wiki root.
type ProjectInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// ProjectList scans cfg.WikiPath for subdirectories that contain an index.md
// and returns them as projects. Always scans the wiki root regardless of any
// active project scope.
func ProjectList(cfg *config.Config) ([]ProjectInfo, *ToolError) {
	entries, err := os.ReadDir(cfg.WikiPath)
	if err != nil {
		return nil, NewToolError(ErrCodeInternal, fmt.Sprintf("cannot read wiki directory: %v", err))
	}

	var projects []ProjectInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		indexPath := filepath.Join(cfg.WikiPath, e.Name(), "index.md")
		if _, err := os.Stat(indexPath); err == nil {
			projects = append(projects, ProjectInfo{
				Name: e.Name(),
				Path: filepath.Join(cfg.WikiPath, e.Name()),
			})
		}
	}
	if projects == nil {
		projects = []ProjectInfo{}
	}
	return projects, nil
}

func newInitIndexDocument(cfg *config.Config) *IndexDocument {
	doc := &IndexDocument{
		Preamble: "# Wiki Index\n\nA personal knowledge wiki.\n\n## Pages\n",
		Stats: IndexStats{
			WikiPages:   "0",
			LastUpdated: time.Now().Format("2006-01-02"),
		},
	}
	for _, sec := range cfg.Index.Sections {
		doc.Sections = append(doc.Sections, IndexSection{
			Key:        sec.Key,
			Title:      sec.Title,
			HeaderLine: "### " + sec.Title,
			Trailing:   "\n",
		})
	}
	return doc
}
