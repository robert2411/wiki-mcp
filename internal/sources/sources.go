// Package sources provides helpers for fetching and extracting external source material.
package sources

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/ledongthuc/pdf"
)

// Error codes returned in structured tool errors.
const (
	ErrCodeBadRequest = "sources.BadRequest"
	ErrCodePathEscape = "sources.PathEscape"
	ErrCodeNotFound   = "sources.NotFound"
	ErrCodeInternal   = "sources.Internal"
)

// ToolError is a structured error returned by source tools.
type ToolError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ToolError) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

func (e *ToolError) JSON() string {
	b, _ := json.Marshal(e)
	return string(b)
}

func newToolError(code, msg string) *ToolError {
	return &ToolError{Code: code, Message: msg}
}

// SourceEntry describes one file in the sources directory.
type SourceEntry struct {
	Name  string    `json:"name"`
	Size  int64     `json:"size"`
	Mtime time.Time `json:"mtime"`
}

// PDFResult is what source_pdf_text returns.
type PDFResult struct {
	Text      string `json:"text"`
	PageCount int    `json:"page_count"`
}

// resolvePath joins sourcesPath with rel and ensures the result stays within sourcesPath.
func resolvePath(sourcesPath, rel string) (string, error) {
	abs := filepath.Clean(filepath.Join(sourcesPath, rel))
	if abs != sourcesPath && !strings.HasPrefix(abs, sourcesPath+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q escapes sources root", rel)
	}
	return abs, nil
}

// slugRe collapses runs of non-alphanumeric, non-dot characters to a single dash.
var slugRe = regexp.MustCompile(`[^a-zA-Z0-9.]+`)

// slugFromURL derives a safe filename slug from a URL.
// e.g. "https://example.com/blog/my-post?q=1" → "example.com-blog-my-post"
func slugFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	combined := strings.Trim(u.Host+u.Path, "/")
	slug := strings.Trim(slugRe.ReplaceAllString(combined, "-"), "-")
	if slug == "" {
		return "", fmt.Errorf("cannot derive slug from URL %q", rawURL)
	}
	return slug, nil
}

// FetchURL fetches rawURL and saves it as a markdown file under sourcesPath.
// If slug is empty, one is derived from the URL. HTML responses are converted
// to markdown; other content types are saved as-is.
func FetchURL(sourcesPath, rawURL, slug string) (string, *ToolError) {
	if rawURL == "" {
		return "", newToolError(ErrCodeBadRequest, "url is required")
	}

	if slug == "" {
		var err error
		slug, err = slugFromURL(rawURL)
		if err != nil {
			return "", newToolError(ErrCodeBadRequest, err.Error())
		}
	}
	if !strings.HasSuffix(slug, ".md") {
		slug += ".md"
	}

	abs, err := resolvePath(sourcesPath, slug)
	if err != nil {
		return "", newToolError(ErrCodePathEscape, err.Error())
	}

	resp, err := http.Get(rawURL) //nolint:gosec // URL comes from caller; no server-side redirect concern here
	if err != nil {
		return "", newToolError(ErrCodeInternal, fmt.Sprintf("fetch %q: %v", rawURL, err))
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", newToolError(ErrCodeInternal, fmt.Sprintf("read response: %v", err))
	}

	var content string
	if strings.Contains(resp.Header.Get("Content-Type"), "text/html") {
		md, convErr := htmltomarkdown.ConvertString(string(body))
		if convErr != nil {
			content = string(body) // raw fallback on parse failure
		} else {
			content = md
		}
	} else {
		content = string(body)
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
		return "", newToolError(ErrCodeInternal, fmt.Sprintf("create dir: %v", err))
	}
	if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
		return "", newToolError(ErrCodeInternal, err.Error())
	}

	rel, _ := filepath.Rel(sourcesPath, abs)
	return rel, nil
}

// PDFText extracts text from the PDF at relPath (relative to sourcesPath).
// If env WIKI_MCP_PREFER_PDFTOTEXT=1 and pdftotext is on PATH, it shells out
// for higher-quality extraction; otherwise it uses the pure-Go pdf library.
func PDFText(sourcesPath, relPath string) (*PDFResult, *ToolError) {
	if relPath == "" {
		return nil, newToolError(ErrCodeBadRequest, "path is required")
	}
	abs, err := resolvePath(sourcesPath, relPath)
	if err != nil {
		return nil, newToolError(ErrCodePathEscape, err.Error())
	}
	if _, err := os.Stat(abs); os.IsNotExist(err) {
		return nil, newToolError(ErrCodeNotFound, fmt.Sprintf("file %q not found", relPath))
	}

	if os.Getenv("WIKI_MCP_PREFER_PDFTOTEXT") == "1" {
		if ptPath, lookErr := exec.LookPath("pdftotext"); lookErr == nil {
			return pdfViaCommand(ptPath, abs)
		}
	}

	return pdfViaLibrary(abs)
}

func pdfViaLibrary(abs string) (*PDFResult, *ToolError) {
	f, r, err := pdf.Open(abs)
	if err != nil {
		return nil, newToolError(ErrCodeInternal, fmt.Sprintf("open PDF: %v", err))
	}
	defer func() { _ = f.Close() }()

	pageCount := r.NumPage()
	plainReader, err := r.GetPlainText()
	if err != nil {
		return nil, newToolError(ErrCodeInternal, fmt.Sprintf("extract text: %v", err))
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(plainReader); err != nil {
		return nil, newToolError(ErrCodeInternal, fmt.Sprintf("read text: %v", err))
	}

	return &PDFResult{Text: buf.String(), PageCount: pageCount}, nil
}

func pdfViaCommand(pdfToText, abs string) (*PDFResult, *ToolError) {
	out, err := exec.Command(pdfToText, abs, "-").Output()
	if err != nil {
		return nil, newToolError(ErrCodeInternal, fmt.Sprintf("pdftotext: %v", err))
	}
	// page count not available via CLI without extra invocation
	return &PDFResult{Text: string(out), PageCount: 0}, nil
}

// List returns all files in sourcesPath with their size and mtime.
// Returns an empty slice (not an error) when the directory does not exist.
func List(sourcesPath string) ([]SourceEntry, *ToolError) {
	entries, err := os.ReadDir(sourcesPath)
	if os.IsNotExist(err) {
		return []SourceEntry{}, nil
	}
	if err != nil {
		return nil, newToolError(ErrCodeInternal, err.Error())
	}

	result := make([]SourceEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, infoErr := e.Info()
		if infoErr != nil {
			continue
		}
		result = append(result, SourceEntry{
			Name:  e.Name(),
			Size:  info.Size(),
			Mtime: info.ModTime(),
		})
	}
	return result, nil
}
