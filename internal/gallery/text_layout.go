package gallery

import (
	"fmt"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

const previewCode = "package main\n\nfunc main() {\n    println(\"Hello, native Go\")\n}\n"

func (s *State) textLayout(c *ui.Context) {
	s.heading(c, "Words, code and space.", "Readable documents, native code editing and considered loading states.")
	columns(c, func() { s.typographyPreview(c) }, func() { s.codePreview(c) })
	s.loadingPreview(c)
	s.scrollPreview(c)
	s.componentPreview(c, "kbd", func() { s.kbdCompositions.view(c, s.Popovers, s.Config.NativeMenus) })
}

func (s *State) typographyPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "A readable article", "Semantic text with inline links, code and lists.")
		display.TypographyProse(c, func(child *ui.Context) {
			display.Typography(child, tr(child, "Made for your desktop"), display.H2)
			display.TypographyRich(child, display.Paragraph, func(inline *ui.Context) {
				ui.Text(inline, tr(inline, "Compose with "))
				ui.RichText(inline, display.TypographyCodeSpan(inline, "ui.Context"))
				ui.Text(inline, tr(inline, " and explore the "))
				display.TypographyLink(inline, tr(inline, "Typography guide"), "https://ui.shadcn.com/docs/components/typography")
				ui.Text(inline, ".")
			})
			display.Typography(child, tr(child, "Clear content leaves room for your next idea."), display.Blockquote)
			display.TypographyList(child, []display.TypographyListItem{
				{Key: "state", Text: tr(child, "Keep state in your application.")},
				{Key: "theme", Text: tr(child, "Use semantic colors across your interface.")},
			}, display.TypographyListProps{Label: tr(child, "Writing principles")})
			if s.articleDetails {
				display.Typography(child, tr(child, "A simple structure"), display.H3)
				display.TypographyTable(child, display.TypographyTableProps{Label: tr(child, "Document structure"), Columns: []display.TypographyTableColumn{{Key: "block", Title: tr(child, "Block")}, {Key: "purpose", Title: tr(child, "Purpose")}}, Rows: []display.TypographyTableRow{{Key: "heading", Cells: []string{"Heading", "Introduce a section"}}, {Key: "body", Cells: []string{"Paragraph", "Explain an idea"}}}})
			}
		}).Label(tr(c, "Preview article"))
		label := "Show article details"
		if s.articleDetails {
			label = "Hide article details"
		}
		if control.Button(c, tr(c, label), control.ButtonProps{Variant: control.Outline, Size: control.Small}).Clicked() {
			s.articleDetails = !s.articleDetails
		}
		button := control.StyleButton(c, ui.ButtonBase(c), "", control.ButtonProps{Variant: control.Ghost, Size: control.Small, Label: tr(c, "Search with shortcut")})
		button.Children(func() {
			ui.Text(c, tr(c, "Search components"))
			display.KbdShortcut(c, ui.Cmd, "K").AlignSelf(ui.Center)
		})
		if button.Clicked() {
			s.Command = true
		}
	})
}

func (s *State) codePreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "A native code surface", "Select, copy and wrap the same read-only buffer.")
		if s.codeLanguage == "" {
			s.codeLanguage = "go"
		}
		control.RadioGroup(c, &s.codeLanguage, control.RadioGroupProps{Label: tr(c, "Code highlighting"), Orientation: ui.Horizontal, Horizontal: true, Options: localizedOptions(c, []control.Option{{Value: "go", Label: "Go syntax"}, {Value: "text", Label: "Plain text"}})})
		ui.Row(c).Wrap().Gap(12).Children(func() {
			control.Switch(c, &s.codeWrap, control.SwitchProps{Label: tr(c, "Wrap code"), Size: control.Small})
			control.Switch(c, &s.codeCompact, control.SwitchProps{Label: tr(c, "Compact code"), Size: control.Small})
			control.Checkbox(c, &s.codeDiff, tr(c, "Highlight changes"), control.CheckboxProps{})
		})
		size := display.CodeBlockDefaultSize
		if s.codeCompact {
			size = display.CodeBlockSmall
		}
		var lines []display.CodeBlockLine
		if s.codeDiff {
			lines = []display.CodeBlockLine{{Index: 2, Variant: display.CodeLineHighlighted}, {Index: 3, Variant: display.CodeLineAdded}}
		}
		display.CodeBlockWith(c, display.CodeBlockProps{Title: "hello.go", Label: tr(c, "Source preview"), Language: s.codeLanguage, Code: previewCode, Wrap: s.codeWrap, LineNumbers: true, StartLine: 10, Height: 168, Size: size, Lines: lines, State: &s.codeExample}, display.CodeBlockSlots{
			Actions: func(child *ui.Context) {
				display.CodeBlockCopyButton(child, previewCode, display.CodeBlockCopyProps{Label: tr(child, "Copy example"), OnCopy: func(string) { s.codeCopied = "Example copied" }, OnError: func(error) { s.codeCopied = "Could not copy example" }})
			},
			Footer: func(child *ui.Context) {
				ui.Text(child, tr(child, "Read-only · numbers and change markers stay out of the clipboard.")).FontSize(12)
			},
		})
		if s.codeCopied != "" {
			ui.Text(c, s.codeCopied).Role(ui.RoleStatus).FontSize(14)
		}
	})
}

func (s *State) loadingPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "Loading, with context", "A shared indicator follows the color of its surrounding content.")
		control.Switch(c, &s.loadingStatic, control.SwitchProps{Label: tr(c, "Pause loading animations"), Size: control.Small})
		ui.Row(c).Wrap().Gap(16).AlignItems(ui.Center).Children(func() {
			display.Spinner(c, display.SpinnerProps{Static: s.loadingStatic, Label: tr(c, "Preview loading indicator"), Description: tr(c, "Refreshing your workspace")})
			display.Spinner(c, display.SpinnerProps{Static: s.loadingStatic, Size: 24, Decorative: true})
			display.BadgeWith(c, tr(c, "Syncing"), display.BadgeProps{Variant: display.BadgeSecondary, Leading: func(child *ui.Context) {
				display.Spinner(child, display.SpinnerProps{Static: s.loadingStatic, Size: 12, Decorative: true})
			}})
			control.Button(c, tr(c, "Sync workspace"), control.ButtonProps{Loading: !s.loadingStatic, Variant: control.Outline})
		})
		control.Field(c, control.FieldProps{Label: tr(c, "Workspace lookup")}, func() ui.Element {
			return control.InputGroup(c, control.InputGroupProps{InlineStart: func() { display.Spinner(c, display.SpinnerProps{Static: s.loadingStatic, Decorative: true}) }}, func() ui.Element {
				return control.Input(c, &s.layoutDraft, control.InputProps{Placeholder: tr(c, "Search your workspace…")})
			})
		})
		ui.Row(c).FillWidth().Gap(16).AlignItems(ui.Center).Children(func() {
			display.Skeleton(c, display.SkeletonProps{Circle: true, Height: 40, Static: s.loadingStatic})
			ui.Column(c).Grow(1).MinWidth(0).Gap(8).Children(func() {
				display.Skeleton(c, display.SkeletonProps{Static: s.loadingStatic})
				display.Skeleton(c, display.SkeletonProps{Static: s.loadingStatic}).WidthPercent(75)
			})
		})
	})
}

func (s *State) scrollPreview(c *ui.Context) {
	display.Card(c, func() {
		s.cardHeader(c, "A window onto your content", "Native wheel and keyboard scrolling with your choice of scrollbar style.")
		control.Switch(c, &s.nativeScrollbars, control.SwitchProps{Label: tr(c, "Use MyGo scrollbars"), Size: control.Small})
		ui.Row(c).Gap(8).Children(func() {
			start := control.Button(c, tr(c, "Start"), control.ButtonProps{Variant: control.Outline, Size: control.Small, Label: tr(c, "Scroll to start")}).Clicked()
			end := control.Button(c, tr(c, "End"), control.ButtonProps{Variant: control.Outline, Size: control.Small, Label: tr(c, "Scroll to end")}).Clicked()
			if start {
				s.articleScroll.X, s.articleScroll.Y = 0, 0
			}
			if end {
				s.articleScroll.X, s.articleScroll.Y = s.articleScroll.MaxX, s.articleScroll.MaxY
			}
		})
		columns(c, func() {
			t := theme.TokensOf(c)
			display.ScrollAreaWith(c, display.ScrollAreaProps{State: &s.articleScroll, Label: tr(c, "Preview notes"), NativeScrollbars: s.nativeScrollbars}, func(child *ui.Context) {
				ui.Column(child).FillWidth().Padding(16).Gap(12).Children(func() {
					for i := 1; i <= 20; i++ {
						ui.Text(child, fmt.Sprintf(tr(child, "Note %02d · Make room for your next idea."), i)).FontSize(14)
						display.Separator(child, false)
					}
				})
			}).Root.Height(180).Border(1, t.Border).Radius(t.RadiusMD())
		}, func() {
			display.AspectRatioWith(c, display.AspectRatioProps{Ratio: 16.0 / 9.0, Clip: true, Label: tr(c, "Artwork preview")}, func(child *ui.Context) {
				ui.Image(child, avatarPortrait).Fill().Fit(ui.Cover).Label(tr(child, "Portrait illustration"))
			}).Radius(theme.TokensOf(c).RadiusMD())
		})
	})
}
