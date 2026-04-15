package main

import (
	"os"
	"os/exec"
	"runtime"
	"testing"
)

func TestBinaryBuilds(t *testing.T) {
	binary := "wiki-mcp"
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}

	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = "."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
	defer func() { _ = os.Remove(binary) }()

	help := exec.Command("./"+binary, "--help")
	out, err := help.CombinedOutput()
	if err != nil {
		// --help exits 0 with flag package, but check output regardless
		_ = err
	}
	if len(out) == 0 {
		t.Fatal("expected --help to produce output")
	}
	t.Logf("--help output:\n%s", out)
}
