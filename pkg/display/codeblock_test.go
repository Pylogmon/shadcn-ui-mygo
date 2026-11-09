package display

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestCodeBlockCopiesCurrentSourceWithoutLineNumbers(t *testing.T) {
	source := "package main\n\nfunc main() {}\n"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		CodeBlock(c, CodeBlockProps{Title: "main.go", Code: source, Language: "go", LineNumbers: true})
	}, 500, 300)
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if r.Clipboard() != source {
		t.Fatalf("clipboard = %q", r.Clipboard())
	}
	source = "// 新的源代码\n"
	r.Frame()
	if err := r.Click("Copy code"); err != nil {
		t.Fatal(err)
	}
	if r.Clipboard() != source {
		t.Fatalf("stale clipboard = %q", r.Clipboard())
	}
}
