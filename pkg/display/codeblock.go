package display

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type CodeBlockProps struct {
	Title, Language, Code string
	LineNumbers           bool
	// Height defaults to 200 DIPs. Code remains selectable and horizontally scrollable.
	Height   float32
	HideCopy bool
}
type codeState struct {
	source, language string
	editor           editor.State
}

// CodeBlock uses MyGo's read-only native code editor for syntax highlighting,
// selection, keyboard copying and large files. The header copy action copies
// the original source without line numbers.
func CodeBlock(c *ui.Context, p CodeBlockProps) ui.Element {
	t := theme.TokensOf(c)
	root := ui.Column(c).FillWidth().Radius(t.Radius).Border(1, t.Border).Clip().Background(t.Card)
	s := ui.Local(root, "code-block", func() codeState { return codeState{} })
	if s.editor.Buffer == nil || s.source != p.Code || s.language != p.Language {
		s.source, s.language = p.Code, p.Language
		s.editor.Buffer = editor.NewBuffer(p.Code)
		s.editor.Buffer.SetLanguage(editor.LanguageNamed(p.Language))
	}
	s.editor.ReadOnly = true
	s.editor.NoActiveLine = true
	s.editor.NoFolding = true
	s.editor.NoIndentGuides = true
	s.editor.NoScrollPastEnd = true
	s.editor.Label = p.Title
	s.editor.LineNumbers = editor.LineNumbersOff
	if p.LineNumbers {
		s.editor.LineNumbers = editor.LineNumbersOn
	}
	height := p.Height
	if height <= 0 {
		height = 200
	}
	return root.Children(func() {
		ui.Row(c).FillWidth().Height(42).PaddingX(12).Gap(8).AlignItems(ui.Center).Background(t.Muted).Children(func() {
			title := p.Title
			if title == "" {
				title = p.Language
			}
			ui.Text(c, title).FontSize(12).TextColor(t.MutedForeground).Grow(1)
			if !p.HideCopy {
				b := ui.ButtonBase(c).Height(28).PaddingX(8).Radius(4).Label("Copy code").Children(func() { ui.Text(c, "Copy").FontSize(12) })
				if b.Clicked() {
					c.WriteClipboard(p.Code)
				}
			}
		})
		editor.View(c, &s.editor).FillWidth().Height(height)
	})
}
