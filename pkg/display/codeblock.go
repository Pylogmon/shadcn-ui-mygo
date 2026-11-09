package display

import (
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/editor"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type CodeBlockVariant uint8

const (
	CodeBlockDefault CodeBlockVariant = iota
	CodeBlockDestructive
)

type CodeBlockSize uint8

const (
	CodeBlockDefaultSize CodeBlockSize = iota
	CodeBlockSmall
)

type CodeBlockLineVariant uint8

const (
	CodeLineDefault CodeBlockLineVariant = iota
	CodeLineAdded
	CodeLineRemoved
	CodeLineHighlighted
)

// CodeBlockLine describes a logical native-editor line. Its sign is decorative,
// so selecting/copying the text never adds diff signs or line numbers.
type CodeBlockLine struct {
	Index   int
	Variant CodeBlockLineVariant
}

// CodeBlockState optionally lets the app retain and inspect the actual native
// editor across owner removal. Code is controlled by CodeBlockProps, while the
// native Editor owns selections, scroll position and find state.
type CodeBlockState struct {
	Editor editor.State
	source string
	ready  bool
}

type CodeBlockProps struct {
	Title, Language, Code string
	LineNumbers           bool
	// Height is the native viewport's height. CodeBlock defaults it to 200;
	// CodeBlockWith defaults to content height. FitContent requests auto sizing
	// explicitly, with optional MinHeight/MaxHeight bounds.
	Height, MinHeight, MaxHeight float32
	FitContent, Wrap             bool
	HideCopy, HideHeader         bool
	Variant                      CodeBlockVariant
	Size                         CodeBlockSize
	Label, Description           string
	StartLine                    int
	Lines                        []CodeBlockLine
	State                        *CodeBlockState
	Handle                       *ui.Handle
	OnCopy                       func(string)
	OnCopyError                  func(error)
}

type CodeBlockSlots struct {
	// Header replaces the title area; Actions replaces the default copy action.
	// Without a header, Actions float at the logical end as in the source.
	Header, Actions, Footer func(*ui.Context)
}

type CodeBlockContentProps struct {
	Language, Label, Description string
	Wrap, LineNumbers            bool
	StartLine                    int
	Height, MinHeight, MaxHeight float32
	Lines                        []CodeBlockLine
	State                        *CodeBlockState
	Handle                       *ui.Handle
}

type CodeBlockCopyProps struct {
	Label, CopiedLabel string
	Disabled           bool
	OnCopy             func(string)
	OnError            func(error)
}

type codeBlockSlot uint8

const (
	codeHeader codeBlockSlot = iota
	codeTitle
	codeActions
	codeContent
	codeFooter
	codeNativeEditor
	codeGutter
	codeCopy
)

type codeBlockScopeKey struct{}
type codeBlockContentStateKey struct{}
type codeBlockGeometryKey struct{}
type codeBlockTargetKey struct{}
type codeCopyStateKey struct{}
type codeBlockScope struct {
	size           CodeBlockSize
	variant        CodeBlockVariant
	header, footer bool
	floating       bool
}

func codeScope(c *ui.Context) *codeBlockScope {
	return ui.Local(c.Root(), codeBlockScopeKey{}, func() codeBlockScope { return codeBlockScope{} })
}

func codeMetrics(c *ui.Context) (font, pad float32) {
	font, pad = 14, 16
	if codeScope(c).size == CodeBlockSmall {
		font, pad = 12, 12
	}
	return
}

func codeForeground(c *ui.Context) ui.Color {
	t := theme.TokensOf(c)
	if codeScope(c).variant == CodeBlockDestructive {
		return t.Destructive
	}
	return t.Foreground
}

// CodeBlock retains the original bounded native-editor API. CodeBlockWith is
// the compound/content-sized entry point.
func CodeBlock(c *ui.Context, p CodeBlockProps) ui.Element {
	if p.Height == 0 && !p.FitContent {
		p.Height = 200
	}
	return CodeBlockWith(c, p, CodeBlockSlots{})
}

func CodeBlockWith(c *ui.Context, p CodeBlockProps, slots CodeBlockSlots) ui.Element {
	t := theme.TokensOf(c)
	bg, border := ui.Transparent, t.Input
	foreground := t.Foreground
	if t.Dark {
		bg = t.Input.Alpha(.3)
	}
	if p.Variant == CodeBlockDestructive {
		foreground, border = t.Destructive, t.Destructive.Alpha(.3)
		if t.Dark {
			border = t.Destructive.Alpha(.4)
		}
		bg = ui.Transparent
	}
	root := ui.Column(c).FillWidth().MinWidth(0).Radius(t.RadiusLG()).Border(1, border).Clip().Background(bg).TextColor(foreground)
	font, leading := float32(14), float32(20.0/14.0)
	if p.Size == CodeBlockSmall {
		font, leading = 12, 16.0/12.0
	}
	root.FontSize(font).LineHeight(leading)
	if p.Label != "" {
		root.Role(ui.RoleGroup).Label(p.Label).Description(p.Description)
	}
	root.Invalid(!codeFinite(p.Height) || !codeFinite(p.MinHeight) || !codeFinite(p.MaxHeight) || p.Height < 0 || p.MinHeight < 0 || p.MaxHeight < 0)
	scope := codeScope(c)
	previous := *scope
	header := !p.HideHeader && (slots.Header != nil || p.Title != "" || p.Language != "")
	*scope = codeBlockScope{size: p.Size, variant: p.Variant, header: header, footer: slots.Footer != nil, floating: !header}
	defer func() { *scope = previous }()
	var body ui.Element
	actions := func(child *ui.Context) {
		if slots.Actions != nil {
			slots.Actions(child)
		} else if !p.HideCopy {
			CodeBlockCopyButton(child, p.Code, CodeBlockCopyProps{Label: "Copy code", OnCopy: p.OnCopy, OnError: p.OnCopyError})
		}
	}
	root.Children(func() {
		if header {
			CodeBlockHeader(c.Key(codeHeader), func(child *ui.Context) {
				ui.Row(child).Key(codeTitle).Basis(0).Grow(1).MinWidth(0).Children(func() {
					if slots.Header != nil {
						slots.Header(child)
					} else {
						title := p.Title
						if title == "" {
							title = p.Language
						}
						CodeBlockTitle(child, title)
					}
				})
				if slots.Actions != nil || !p.HideCopy {
					CodeBlockActions(child.Key(codeActions), actions)
				}
			})
		}
		height := p.Height
		if p.FitContent {
			height = 0
		}
		label := p.Label
		if label == "" {
			label = p.Title
		}
		body = CodeBlockContent(c.Key(codeContent), p.Code, CodeBlockContentProps{Language: p.Language, Label: label, Description: p.Description,
			Wrap: p.Wrap, LineNumbers: p.LineNumbers, StartLine: p.StartLine, Height: height, MinHeight: p.MinHeight, MaxHeight: p.MaxHeight,
			Lines: p.Lines, State: p.State, Handle: p.Handle})
		if slots.Footer != nil {
			CodeBlockFooter(c.Key(codeFooter), slots.Footer)
		}
		if !header && (slots.Actions != nil || !p.HideCopy) {
			_, pad := codeMetrics(c)
			a := CodeBlockActions(c.Key(codeActions), actions).Absolute().Top(pad - 4).ZIndex(1).Margin(0)
			if c.Direction() == ui.RightToLeft {
				a.Left(pad - 4)
			} else {
				a.Right(pad - 4)
			}
			if !root.Hovered() && !root.FocusWithin() {
				a.Opacity(0)
			}
			if !motion.Reduced(c) {
				a.Transition(ui.ElementTransition{Duration: motion.Normal, Ease: motion.EaseOut})
			}
		}
	})
	if geometry := ui.Local(body, codeBlockGeometryKey{}, func() codeContentGeometry { return codeContentGeometry{} }); geometry.invalid {
		root.Invalid(true)
	}
	if handle := ui.Local(body, codeBlockTargetKey{}, func() ui.Handle { return ui.Handle{} }); c.Resolve(*handle).FocusVisible() {
		root.Shadow(0, 0, 0, 3, t.Ring.Alpha(.5))
		if p.Variant != CodeBlockDestructive {
			root.BorderColor(t.Ring)
		}
	}
	return root
}

func CodeBlockHeader(c *ui.Context, content func(*ui.Context)) ui.Element {
	_, pad := codeMetrics(c)
	color := theme.TokensOf(c).MutedForeground
	if codeScope(c).variant == CodeBlockDestructive {
		color = codeForeground(c)
	}
	return ui.Row(c).FillWidth().MinWidth(0).Shrink(0).AlignItems(ui.Center).Gap(8).Padding(pad, pad, 0).TextColor(color).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

func CodeBlockTitle(c *ui.Context, title string) ui.Element {
	font, _ := codeMetrics(c)
	leading := float32(20.0 / 14.0)
	if codeScope(c).size == CodeBlockSmall {
		leading = 16.0 / 12.0
	}
	return ui.Text(c, title).MinWidth(0).Font(c.Theme().Font).FontSize(font).LineHeight(leading).FontWeight(500).TextColor(codeForeground(c)).SingleLine()
}

func CodeBlockActions(c *ui.Context, content func(*ui.Context)) ui.Element {
	return ui.Row(c).AlignItems(ui.Center).Shrink(0).Gap(4).Margin(-4, 0).MarginEnd(-6).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

// CodeBlockAction applies the source's floating-action surface to an arbitrary
// native button. Header actions retain their ordinary button appearance.
func CodeBlockAction(c *ui.Context, button ui.Element) ui.Element {
	if codeScope(c).floating {
		t := theme.TokensOf(c)
		button.Background(t.Background)
		if !button.FocusVisible() {
			button.Shadow(0, 0, 0, 1, t.Foreground.Alpha(.1))
		}
	}
	return button
}

func CodeBlockFooter(c *ui.Context, content func(*ui.Context)) ui.Element {
	_, pad := codeMetrics(c)
	color := theme.TokensOf(c).MutedForeground
	return ui.Row(c).FillWidth().MinWidth(0).Shrink(0).AlignItems(ui.Center).Gap(8).Padding(0, pad, pad).TextColor(color).Children(func() {
		if content != nil {
			content(c)
		}
	})
}

type codeCopyState struct {
	sequence      uint64
	ready, copied bool
	err           error
	value         string
	until         time.Time
}

// CodeBlockCopyButton confirms the native clipboard write before showing the
// source's two-second feedback. The deferred completion keeps only local data,
// never a Context or an Element; callbacks/announcements run in a live build.
func CodeBlockCopyButton(c *ui.Context, value string, options ...CodeBlockCopyProps) ui.Element {
	p := CodeBlockCopyProps{}
	if len(options) > 0 {
		p = options[0]
	}
	root := ui.Box(c)
	s := ui.Local(root, codeCopyStateKey{}, func() codeCopyState { return codeCopyState{} })
	if s.copied && !c.Now().Before(s.until) {
		s.copied = false
	}
	if s.ready {
		s.ready = false
		if s.err == nil {
			if !s.copied {
				s.until = c.Now().Add(2 * time.Second)
			}
			s.copied = true
			c.Announce("Copied")
			root.SetChanged()
			if p.OnCopy != nil {
				p.OnCopy(s.value)
			}
		} else if p.OnError != nil {
			p.OnError(s.err)
		}
	}
	label := p.Label
	if label == "" {
		label = "Copy"
	}
	icon := codeCopyIcon
	if s.copied {
		icon = codeCopiedIcon
		label = p.CopiedLabel
		if label == "" {
			label = "Copied"
		}
		c.After(s.until.Sub(c.Now()))
	}
	root.Children(func() {
		button := control.Button(c.Key(codeCopy), "", control.ButtonProps{Variant: control.Ghost, Size: control.IconExtraSmall, Label: label, Icon: icon, Disabled: p.Disabled})
		if !button.Hovered() && !button.Pressed() {
			button.TextColor(codeForeground(c))
		}
		CodeBlockAction(c, button)
		if button.Clicked() && !button.IsDisabled() {
			s.sequence++
			sequence := s.sequence
			c.WriteClipboardThen(value, func(err error) {
				if s.sequence == sequence {
					s.err, s.value, s.ready = err, value, true
				}
			})
		}
	})
	return root
}

type codeContentGeometry struct {
	height  float32
	invalid bool
}

func codeFinite(value float32) bool {
	return !math.IsNaN(float64(value)) && !math.IsInf(float64(value), 0)
}

func codeSync(s *CodeBlockState, code, language string) {
	normalized := strings.ReplaceAll(code, "\r\n", "\n")
	if s.Editor.Buffer == nil {
		s.Editor.Buffer = editor.NewBuffer(code)
	} else if !s.ready || s.source != code {
		old := s.Editor.Buffer.String()
		prefix := 0
		for prefix < len(old) && prefix < len(normalized) && old[prefix] == normalized[prefix] {
			prefix++
		}
		for prefix > 0 && prefix < len(normalized) && !utf8.RuneStart(normalized[prefix]) {
			prefix--
		}
		suffix := 0
		for suffix < len(old)-prefix && suffix < len(normalized)-prefix && old[len(old)-suffix-1] == normalized[len(normalized)-suffix-1] {
			suffix++
		}
		for suffix > 0 && !utf8.RuneStart(normalized[len(normalized)-suffix]) {
			suffix--
		}
		if prefix != len(old) || prefix != len(normalized) {
			s.Editor.Buffer.Edit(editor.Edit{Start: prefix, End: len(old) - suffix, Text: normalized[prefix : len(normalized)-suffix]})
		}
	}
	s.source, s.ready = code, true
	s.Editor.Buffer.SetLanguage(editor.LanguageNamed(language))
}

var codeAppearanceBindings = func() []editor.Binding {
	var bindings []editor.Binding
	for _, binding := range editor.DefaultBindings() {
		if binding.Command == editor.ToggleSoftWrap || binding.Command == editor.ToggleLineNumbers {
			binding.Command = editor.NoCommand
			bindings = append(bindings, binding)
		}
	}
	return bindings
}()

func codeViewerBindings(previous []editor.Binding) []editor.Binding {
	bindings := append(make([]editor.Binding, 0, len(previous)+len(codeAppearanceBindings)), codeAppearanceBindings...)
	for _, binding := range previous {
		reserved := false
		for _, appearance := range codeAppearanceBindings {
			if binding.Mods == appearance.Mods && binding.Key == appearance.Key && binding.ThenMods == appearance.ThenMods && binding.ThenKey == appearance.ThenKey {
				reserved = true
				break
			}
		}
		if !reserved {
			bindings = append(bindings, binding)
		}
	}
	return bindings
}

func CodeBlockContent(c *ui.Context, code string, p CodeBlockContentProps) ui.Element {
	font, pad := codeMetrics(c)
	scope := *codeScope(c)
	root := ui.Column(c).FillWidth().MinWidth(0).Direction(ui.LeftToRight)
	s := p.State
	if s == nil {
		s = ui.Local(root, codeBlockContentStateKey{}, func() CodeBlockState { return CodeBlockState{} })
	}
	geometry := ui.Local(root, codeBlockGeometryKey{}, func() codeContentGeometry { return codeContentGeometry{} })
	codeSync(s, code, p.Language)
	scale := c.Preferences().TextScale
	if !codeFinite(scale) || scale <= 0 {
		scale = 1
	}
	size := font * scale
	geometry.invalid = false
	if !codeFinite(size) || !codeFinite(size*1.625) {
		size, geometry.invalid = font, true
	}
	lineHeight := size * 1.625
	minimum, maximum := max(lineHeight, p.MinHeight), p.MaxHeight
	valid := codeFinite(p.Height) && codeFinite(p.MinHeight) && codeFinite(maximum) && p.Height >= 0 && p.MinHeight >= 0 && maximum >= 0
	if !valid {
		minimum, maximum = lineHeight, 0
	}
	if maximum > 0 && maximum < minimum {
		maximum, valid = minimum, false
	}
	height := p.Height
	if !valid {
		height = 200
	}
	auto := height == 0
	if auto {
		height = max(lineHeight, geometry.height)
	}
	height = max(minimum, height)
	if maximum > 0 {
		height = min(height, maximum)
	}
	st := &s.Editor
	st.Font = ui.Font{Family: "monospace", Size: size}
	st.LineHeight = 1.625
	st.ReadOnly, st.NoActiveLine, st.NoFolding, st.NoIndentGuides, st.NoScrollPastEnd = true, true, true, true, true
	st.Whitespace = editor.WhitespaceNone
	st.SoftWrap, st.FitContent = p.Wrap, false
	st.LineNumbers = editor.LineNumbersOff
	// The viewer owns Wrap and its offset/decoration gutter. Prevent the native
	// editor's display-only commands from installing a second gutter or reversing
	// the app's Wrap value; ordinary selection/find/copy bindings remain native.
	st.Bindings = codeViewerBindings(st.Bindings)
	st.Label = p.Label
	if st.Label == "" {
		st.Label = "Code"
	}
	t := theme.TokensOf(c)
	ct := editor.DefaultTheme(c.Theme())
	// A nonzero RGB with zero alpha survives the editor theme's zero-means-
	// default resolver and leaves the actual themed parent surface visible.
	ct.Background = ui.RGB(1, 1, 1).Alpha(0)
	ct.Text, ct.LineNumber, ct.ActiveLineNumber = codeForeground(c), t.MutedForeground, t.MutedForeground
	st.Theme = ct
	st.LineBackgrounds = st.LineBackgrounds[:0]
	start := p.StartLine
	if start <= 0 {
		start = 1
	}
	if start > math.MaxInt-st.Buffer.LineCount()+1 {
		start, valid = 1, false
	}
	geometry.invalid = geometry.invalid || !valid
	span := func(text string, color ui.Color) ui.Span {
		return ui.Span{Text: text, Font: "monospace", Size: size, Color: color}
	}
	gutterWidth := float32(0)
	if p.LineNumbers {
		gutterWidth, _ = c.MeasureText(0, span(strconv.Itoa(start+st.Buffer.LineCount()-1), t.MutedForeground))
		gutterWidth = max(24, gutterWidth) + 12
	} else if len(p.Lines) > 0 {
		gutterWidth = 16
	}
	top, bottom := pad, pad
	if scope.header {
		top /= 2
	}
	if scope.footer {
		bottom /= 2
	}
	// The native editor already reserves eight DIPs before its text.
	root.Padding(top, pad-8, bottom).Invalid(!valid || geometry.invalid)
	target := ui.Local(root, codeBlockTargetKey{}, func() ui.Handle { return ui.Handle{} })
	root.Children(func() {
		ui.Row(c).FillWidth().MinWidth(0).Direction(ui.LeftToRight).Children(func() {
			if gutterWidth > 0 {
				ui.Box(c.Key(codeGutter)).Width(gutterWidth).Height(height).Shrink(0).AccessHidden(true).IgnoreInput(true).Draw(func(painter *ui.Painter, bounds ui.Rect) {
					clip := bounds
					clip.X -= pad - 8
					clip.W += pad - 8
					painter.Clip(clip, 0, func() {
						count := st.Buffer.LineCount()
						low, high := 0, count
						for low < high {
							middle := low + (high-low)/2
							if st.LineTop(middle) < st.ScrollY() {
								low = middle + 1
							} else {
								high = middle
							}
						}
						for index := max(0, low-1); index < count; index++ {
							y := bounds.Y + float32(st.LineTop(index)-st.ScrollY())
							if y >= bounds.Y+bounds.H {
								break
							}
							if p.LineNumbers {
								text := span(strconv.Itoa(start+index), t.MutedForeground)
								w, h := painter.MeasureText(0, text)
								painter.RichText(bounds.X+bounds.W-4-w, y+(lineHeight-h)/2, 0, text)
							}
							for _, line := range p.Lines {
								if line.Index != index {
									continue
								}
								text, color := "", t.MutedForeground
								if line.Variant == CodeLineAdded {
									text = "+"
								}
								if line.Variant == CodeLineRemoved {
									text, color = "−", t.Destructive
								}
								if text != "" {
									painter.RichText(bounds.X+4-(pad-8), y, 0, span(text, color))
								}
							}
						}
					})
				})
			}
			e := editor.View(c.Key(codeNativeEditor), st).Grow(1).MinWidth(0).Height(height).Bind(target).Description(p.Description).Invalid(geometry.invalid)
			if p.Handle != nil {
				e.Bind(p.Handle)
			}
		})
	})
	root.Draw(func(painter *ui.Painter, bounds ui.Rect) {
		if auto {
			next := float32(st.LineTop(st.Buffer.LineCount()))
			if codeFinite(next) && math.Abs(float64(next-geometry.height)) > .25 {
				geometry.height = next
				c.Invalidate()
			}
		}
		painter.Clip(ui.Rect{X: bounds.X, Y: bounds.Y + top, W: bounds.W, H: height}, 0, func() {
			for _, line := range p.Lines {
				if line.Index < 0 || line.Index >= st.Buffer.LineCount() {
					continue
				}
				color := ui.Transparent
				switch line.Variant {
				case CodeLineAdded, CodeLineHighlighted:
					color = t.Muted
				case CodeLineRemoved:
					color = t.Destructive.Alpha(.1)
				}
				if color.A == 0 {
					continue
				}
				y := bounds.Y + top + float32(st.LineTop(line.Index)-st.ScrollY())
				h := float32(st.LineTop(line.Index+1) - st.LineTop(line.Index))
				painter.Fill(ui.Rect{X: bounds.X, Y: y, W: bounds.W, H: h}, color, 0)
			}
		})
	})
	return root
}
