package gallery

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/fonts"
	fontcatalog "github.com/pylogmon/shadcn-ui-mygo/pkg/fonts/all"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Preset is the portable design state of the creator. Font choices are stable
// bundled catalogue IDs or native categories, resolved to registered families.
type Preset struct {
	Version      int              `json:"version"`
	Base         string           `json:"base"`
	Accent       string           `json:"accent"`
	Chart        string           `json:"chart"`
	Radius       float32          `json:"radius"`
	BodyFont     string           `json:"bodyFont"`
	HeadingFont  string           `json:"headingFont"`
	Appearance   theme.Mode       `json:"appearance"`
	ReduceMotion bool             `json:"reduceMotion"`
	NativeMenus  bool             `json:"nativeMenus"`
	MenuColor    theme.MenuColor  `json:"menuColor"`
	MenuAccent   theme.MenuAccent `json:"menuAccent"`
}

var baseColors = []string{"neutral", "stone", "zinc", "mauve", "olive", "mist", "taupe"}
var galleryFonts = fonts.Catalog()
var bodyFonts = fontChoices(false)
var headingFonts = fontChoices(true)
var radiusValues = []float32{0, 4, 6, 8, 10, 12, 16, 24}

func DefaultPreset() Preset {
	return Preset{Version: 1, Base: "neutral", Accent: "neutral", Chart: "neutral", Radius: 10, BodyFont: "inter", HeadingFont: theme.FontInherit, Appearance: theme.System, NativeMenus: false, MenuColor: theme.MenuDefault, MenuAccent: theme.MenuSubtle}
}
func (p Preset) Validate() error {
	if err := (theme.MenuOptions{Color: p.MenuColor, Accent: p.MenuAccent}).Validate(); err != nil {
		return err
	}
	if p.Version != 1 {
		return fmt.Errorf("unsupported preset version %d", p.Version)
	}
	if !slices.Contains(baseColors, p.Base) {
		return fmt.Errorf("unknown base color %q", p.Base)
	}
	if !slices.Contains(theme.Presets(), p.Accent) {
		return fmt.Errorf("unknown accent theme %q", p.Accent)
	}
	if !slices.Contains(theme.Presets(), p.Chart) {
		return fmt.Errorf("unknown chart theme %q", p.Chart)
	}
	if !slices.Contains(bodyFonts, p.BodyFont) || p.HeadingFont != "same" && !slices.Contains(headingFonts, p.HeadingFont) {
		return fmt.Errorf("unsupported native font choice")
	}
	if p.Radius < 0 || p.Radius > 32 || math.IsNaN(float64(p.Radius)) || math.IsInf(float64(p.Radius), 0) {
		return fmt.Errorf("radius must be between 0 and 32 DIPs")
	}
	if p.Appearance != theme.System && p.Appearance != theme.LightMode && p.Appearance != theme.DarkMode {
		return fmt.Errorf("unknown appearance %q", p.Appearance)
	}
	return nil
}

func ParsePreset(source string) (Preset, error) {
	if len(source) > 16*1024 {
		return Preset{}, fmt.Errorf("preset exceeds 16 KiB")
	}
	var p Preset
	d := json.NewDecoder(strings.NewReader(source))
	d.DisallowUnknownFields()
	if err := d.Decode(&p); err != nil {
		return Preset{}, err
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF {
		if err == nil {
			return Preset{}, fmt.Errorf("only one preset object is allowed")
		}
		return Preset{}, err
	}
	if err := p.Validate(); err != nil {
		return Preset{}, err
	}
	return p.Normalized(), nil
}
func (p Preset) JSON() string {
	b, _ := json.MarshalIndent(p.Normalized(), "", "  ")
	return string(b) + "\n"
}

// Tokens composes full base surfaces with independent accent and chart colors.
// Neutral accent retains the base palette's own neutral emphasis colors.
func (p Preset) Tokens(systemDark bool) *theme.Tokens {
	dark := systemDark
	if p.Appearance == theme.LightMode {
		dark = false
	} else if p.Appearance == theme.DarkMode {
		dark = true
	}
	t := theme.Preset(p.Base, dark)
	if p.Accent != "neutral" {
		a := theme.Preset(p.Accent, dark)
		t.Primary = a.Primary
		t.PrimaryForeground = a.PrimaryForeground
		t.Ring = a.Ring
		t.SidebarPrimary = a.SidebarPrimary
		t.SidebarPrimaryForeground = a.SidebarPrimaryForeground
		t.SidebarRing = a.SidebarRing
	}
	chart := theme.Preset(p.Chart, dark)
	t.Chart1, t.Chart2, t.Chart3, t.Chart4, t.Chart5 = chart.Chart1, chart.Chart2, chart.Chart3, chart.Chart4, chart.Chart5
	t.Radius = p.Radius
	if (theme.MenuOptions{Color: p.MenuColor, Accent: p.MenuAccent}).Normalized().Accent == theme.MenuBold {
		t.Accent, t.AccentForeground = t.Primary, t.PrimaryForeground
	}
	return t
}
func nativeFont(name string) string {
	switch name {
	case "mono":
		return "monospace"
	case "serif":
		return "Georgia, Times New Roman, Liberation Serif, DejaVu Serif"
	}
	for _, font := range galleryFonts {
		if font.ID == name {
			return font.Family
		}
	}
	return ""
}
func (p Preset) BodyFamily() string { return nativeFont(p.BodyFont) }
func (p Preset) HeadingFamily() string {
	font := p.HeadingFont
	if font == "same" || font == theme.FontInherit {
		font = p.BodyFont
	}
	return nativeFont(font)
}

func (s *State) ApplyPreset(p Preset) error {
	if err := p.Validate(); err != nil {
		return err
	}
	p = p.Normalized()
	if err := p.registerFonts(); err != nil {
		return err
	}
	if p == s.Config {
		return nil
	}
	s.undo = append(s.undo, s.Config)
	if len(s.undo) > 100 {
		s.undo = s.undo[len(s.undo)-100:]
	}
	s.redo = nil
	s.setPreset(p)
	return nil
}
func (s *State) setPreset(p Preset) {
	s.designRevision++
	s.Config = p
	s.Theme.Mode = p.Appearance
	s.Theme.Preset = p.Accent
	s.Theme.Radius = p.Radius
	s.Theme.ReduceMotion = p.ReduceMotion
	s.Theme.MenuColor, s.Theme.MenuAccent = p.MenuColor, p.MenuAccent
	s.lastThemeMode = p.Appearance
}
func (s *State) ImportPreset(source string) error {
	p, err := ParsePreset(source)
	if err != nil {
		return err
	}
	return s.ApplyPreset(p)
}
func (s *State) ExportPreset() string { return s.Config.JSON() }
func (s *State) CanUndo() bool        { return len(s.undo) > 0 }
func (s *State) CanRedo() bool        { return len(s.redo) > 0 }
func (s *State) Undo() bool {
	if !s.CanUndo() {
		return false
	}
	p := s.undo[len(s.undo)-1]
	s.undo = s.undo[:len(s.undo)-1]
	s.redo = append(s.redo, s.Config)
	s.setPreset(p)
	return true
}
func (s *State) Redo() bool {
	if !s.CanRedo() {
		return false
	}
	p := s.redo[len(s.redo)-1]
	s.redo = s.redo[:len(s.redo)-1]
	s.undo = append(s.undo, s.Config)
	s.setPreset(p)
	return true
}
func (s *State) Reset() {
	if err := s.ApplyPreset(DefaultPreset()); err != nil {
		s.status = err.Error()
	}
}
func (s *State) Randomize(seed uint64) {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	next := s.Config
	presets := theme.Presets()
	next.Base = baseColors[r.IntN(len(baseColors))]
	next.Accent = presets[r.IntN(len(presets))]
	next.Chart = presets[r.IntN(len(presets))]
	next.Radius = radiusValues[r.IntN(len(radiusValues))]
	next.BodyFont = bodyFonts[r.IntN(len(bodyFonts))]
	next.HeadingFont = headingFonts[r.IntN(len(headingFonts))]
	colors := []theme.MenuColor{theme.MenuDefault, theme.MenuInverted, theme.MenuDefaultTranslucent, theme.MenuInvertedTranslucent}
	next.MenuColor = colors[r.IntN(len(colors))]
	next.MenuAccent = theme.MenuSubtle
	if r.IntN(2) == 1 {
		next.MenuAccent = theme.MenuBold
	}
	if err := s.ApplyPreset(next); err != nil {
		s.status = err.Error()
	}
}

func chromeTokens() *theme.Tokens {
	t := theme.Dark()
	t.Background = ui.Hex("#0d0d0f")
	t.Card = ui.Hex("#18181b")
	t.Muted = ui.Hex("#242427")
	t.Secondary = t.Muted
	t.Accent = t.Muted
	t.Border = ui.Hex("#303035")
	t.Input = t.Border
	t.MutedForeground = ui.Hex("#9a9aa3")
	t.Radius = 12
	return t
}

func (p Preset) Normalized() Preset {
	m := (theme.MenuOptions{Color: p.MenuColor, Accent: p.MenuAccent}).Normalized()
	p.MenuColor, p.MenuAccent = m.Color, m.Accent
	if p.HeadingFont == "same" {
		p.HeadingFont = theme.FontInherit
	} else if p.HeadingFont == p.BodyFont {
		// Only the source's named catalogue uses equal => inherit. Native
		// categories are older explicit choices: system headings must stay
		// system when the body is later changed to a bundled family.
		for _, font := range galleryFonts {
			if font.ID == p.BodyFont {
				p.HeadingFont = theme.FontInherit
				break
			}
		}
	}
	return p
}

func fontChoices(headings bool) []string {
	choices := make([]string, 0, len(galleryFonts)+4)
	if headings {
		choices = append(choices, theme.FontInherit)
	}
	for _, font := range galleryFonts {
		choices = append(choices, font.ID)
	}
	// Existing exported Version 1 presets keep their native category choices.
	return append(choices, "system", "serif", "mono")
}

func (p Preset) FontRoles() theme.FontRoles {
	heading := p.Normalized().HeadingFont
	if heading != theme.FontInherit {
		if heading == theme.FontSystem {
			heading = theme.FontSystem
		} else {
			heading = nativeFont(heading)
		}
	}
	return theme.FontRoles{Body: p.BodyFamily(), Heading: heading}.Normalized()
}

func (p Preset) registerFonts() error {
	for _, id := range []string{p.BodyFont, p.Normalized().HeadingFont} {
		switch id {
		case "system", "serif", "mono", theme.FontInherit:
			continue
		}
		font, found := fontcatalog.Lookup(id)
		if !found {
			return fmt.Errorf("unknown preview font %q", id)
		}
		if err := font.Register(); err != nil {
			return fmt.Errorf("could not load %s: %w", font.Info().Name, err)
		}
	}
	return nil
}

func (s *State) previewFonts() theme.FontRoles {
	if err := s.Config.registerFonts(); err != nil {
		s.fontError = "The selected font could not be loaded. Using system fonts."
		return theme.FontRoles{}.Normalized()
	}
	s.fontError = ""
	return s.Config.FontRoles()
}

func (s *State) setFont(heading bool, id string) {
	p := s.Config
	if heading {
		p.HeadingFont = id
	} else {
		p.BodyFont = id
	}
	if err := s.ApplyPreset(p); err != nil {
		s.status = err.Error()
	}
}
func (p Preset) MenuOptions() theme.MenuOptions {
	dark := p
	dark.Appearance = theme.DarkMode
	return theme.MenuOptions{Color: p.MenuColor, Accent: p.MenuAccent, InvertedTokens: dark.Tokens(true)}
}
func (s *State) setMenuColor(color theme.MenuColor) {
	p := s.Config
	if !(theme.MenuOptions{Color: p.MenuColor}).Normalized().ColorIsTranslucent() {
		s.lastSolidMenuAccent = p.MenuAccent
	}
	p.MenuColor = color
	if (theme.MenuOptions{Color: color}).Normalized().ColorIsTranslucent() {
		p.MenuAccent = theme.MenuSubtle
	} else {
		p.MenuAccent = s.lastSolidMenuAccent
	}
	_ = s.ApplyPreset(p)
}
