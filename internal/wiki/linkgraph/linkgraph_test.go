package linkgraph

import (
	"reflect"
	"testing"
)

func TestParseOutgoing_WikiLinks(t *testing.T) {
	content := "See [[Qwen2.5 Coder]] and [[LLM Quantization|quant]]."
	links := ParseOutgoing("entities/page.md", content)

	want := []string{"Qwen2.5 Coder", "LLM Quantization"}
	if !reflect.DeepEqual(links.Internal, want) {
		t.Errorf("Internal = %v, want %v", links.Internal, want)
	}
	if len(links.External) != 0 {
		t.Errorf("External should be empty, got %v", links.External)
	}
}

func TestParseOutgoing_MDLinks_Relative(t *testing.T) {
	content := "See [Coder](../entities/qwen2.5-coder.md) and [local](./other.md)."
	links := ParseOutgoing("concepts/page.md", content)

	want := []string{"entities/qwen2.5-coder.md", "concepts/other.md"}
	if !reflect.DeepEqual(links.Internal, want) {
		t.Errorf("Internal = %v, want %v", links.Internal, want)
	}
}

func TestParseOutgoing_ExternalLinks(t *testing.T) {
	content := "See [Go](https://golang.org) and [GitHub](https://github.com/foo)."
	links := ParseOutgoing("page.md", content)

	if len(links.Internal) != 0 {
		t.Errorf("Internal should be empty, got %v", links.Internal)
	}
	if len(links.External) != 2 {
		t.Errorf("External len = %d, want 2", len(links.External))
	}
}

func TestParseOutgoing_MixedLinks(t *testing.T) {
	content := "[[Title]] and [text](./local.md) and [ext](https://example.com)."
	links := ParseOutgoing("dir/page.md", content)

	if len(links.Internal) != 2 {
		t.Errorf("Internal len = %d, want 2: %v", len(links.Internal), links.Internal)
	}
	if links.Internal[0] != "Title" {
		t.Errorf("Internal[0] = %q, want %q", links.Internal[0], "Title")
	}
	if links.Internal[1] != "dir/local.md" {
		t.Errorf("Internal[1] = %q, want %q", links.Internal[1], "dir/local.md")
	}
	if len(links.External) != 1 {
		t.Errorf("External len = %d, want 1", len(links.External))
	}
}

func TestParseOutgoing_WikiLinkAlias(t *testing.T) {
	content := "[[Target Title|display alias]] here."
	links := ParseOutgoing("page.md", content)

	if len(links.Internal) != 1 || links.Internal[0] != "Target Title" {
		t.Errorf("Internal = %v, want [Target Title]", links.Internal)
	}
}

func TestParseOutgoing_Dedup(t *testing.T) {
	content := "[A](./a.md) and [B](./a.md) and [[Title]] and [[Title]]."
	links := ParseOutgoing("page.md", content)

	if len(links.Internal) != 2 {
		t.Errorf("Internal len = %d, want 2 (deduped): %v", len(links.Internal), links.Internal)
	}
}
