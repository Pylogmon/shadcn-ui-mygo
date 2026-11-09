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
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// Preset is the portable design state of the creator. Fonts name actual native
// font families/categories; it deliberately has no unimplemented style/icon API.
type Preset struct {
	Version      int        `json:"version"`
	Base         string     `json:"base"`
	Accent       string     `json:"accent"`
	Chart        string     `json:"chart"`
	Radius       float32    `json:"radius"`
	BodyFont     string     `json:"bodyFont"`
	HeadingFont  string     `json:"headingFont"`
	Appearance   theme.Mode `json:"appearance"`
	ReduceMotion bool       `json:"reduceMotion"`
	NativeMenus  bool       `json:"nativeMenus"`
}

var baseColors = []string{"neutral", "stone", "zinc", "mauve", "olive", "mist", "taupe"}
var bodyFonts = []string{"system", "serif", "mono"}
var headingFonts = []string{"same", "system", "serif", "mono"}
var radiusValues = []float32{0, 4, 6, 8, 10, 12, 16, 24}

func DefaultPreset() Preset {
	return Preset{Version: 1, Base: "neutral", Accent: "neutral", Chart: "neutral", Radius: 10, BodyFont: "system", HeadingFont: "same", Appearance: theme.System, NativeMenus: false}
}
func (p Preset) Validate() error {
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
	if !slices.Contains(bodyFonts, p.BodyFont) || !slices.Contains(headingFonts, p.HeadingFont) {
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
	return p, nil
}
func (p Preset) JSON() string { b, _ := json.MarshalIndent(p, "", "  "); return string(b) + "\n" }

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
	return t
}
func nativeFont(name string) string {
	switch name {
	case "mono":
		return "monospace"
	case "serif":
		return "Georgia, Times New Roman, Liberation Serif, DejaVu Serif"
	}
	return ""
}
func (p Preset) BodyFamily() string { return nativeFont(p.BodyFont) }
func (p Preset) HeadingFamily() string {
	font := p.HeadingFont
	if font == "same" {
		font = p.BodyFont
	}
	return nativeFont(font)
}

func (s *State) ApplyPreset(p Preset) error {
	if err := p.Validate(); err != nil {
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
func (s *State) Reset() { _ = s.ApplyPreset(DefaultPreset()) }
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
	_ = s.ApplyPreset(next)
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
