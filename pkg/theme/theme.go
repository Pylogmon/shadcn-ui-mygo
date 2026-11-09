package theme

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"weak"

	"github.com/egoist/mygo/ui"
)

// Metadata associates an immutable token snapshot with its native theme.
// There is no global current theme: every Context and Themed subtree keeps its
// own identity. Weak keys and cleanup avoid retaining finished windows/themes.
var metadata sync.Map // map[weak.Pointer[ui.Theme]]Tokens

func (t *Tokens) Theme() *ui.Theme {
	n := &ui.Theme{
		Dark: t.Dark, Background: t.Background, Surface: t.Secondary, SurfaceHover: t.Accent,
		SurfacePressed: Mix(t.Accent, t.Foreground, .08), Border: t.Border, Text: t.Foreground, TextMuted: t.MutedForeground,
		Accent: t.Primary, AccentHover: Mix(t.Primary, t.Background, .1), AccentPressed: Mix(t.Primary, t.Background, .2), AccentText: t.PrimaryForeground,
		InactiveAccent: t.Primary, Danger: t.Destructive, Warning: OKLCH(.769, .188, 70.08), Success: OKLCH(.627, .194, 149.214),
		Selection: t.Primary.Alpha(.25), Focus: t.Ring, FocusRingWidth: 2, FocusRingOffset: 2,
		Inverse: t.Foreground, InverseText: t.Background, Scrollbar: t.MutedForeground.Alpha(.5), ScrollbarWidth: 6,
		Radius: t.Radius, Spacing: t.Spacing, FontSize: t.FontSize,
	}
	key := weak.Make(n)
	metadata.Store(key, *t)
	runtime.AddCleanup(n, func(k weak.Pointer[ui.Theme]) { metadata.Delete(k) }, key)
	runtime.KeepAlive(n)
	return n
}

// TokensOf returns a copy of the tokens of the innermost theme. Changing the
// result never changes another widget. Use Scope or Controller to apply edits.
func TokensOf(c *ui.Context) *Tokens {
	if native := c.Theme(); native != nil {
		if value, ok := metadata.Load(weak.Make(native)); ok {
			t := value.(Tokens)
			return &t
		}
		if native.Dark {
			return Dark()
		}
	}
	return Light()
}

// Scope builds a container with its own complete semantic tokens. Overlays
// opened within its children inherit the same theme through MyGo.
func Scope(c *ui.Context, t *Tokens) ui.Element { return ui.Themed(c, t.Theme()) }

// Apply installs Neutral, following the appearance supplied by MyGo for this
// frame. Call at the start of the view. Existing complete scoped themes stay.
func Apply(c *ui.Context) {
	if native := c.Theme(); native != nil {
		if _, ok := metadata.Load(weak.Make(native)); ok {
			return
		}
	}
	c.SetTheme(TokensOf(c).Theme())
}

type Mode string

const (
	System    Mode = "system"
	LightMode Mode = "light"
	DarkMode  Mode = "dark"
)

// Controller is application-owned state. Its zero value follows the desktop.
// Apply belongs at the beginning of each frame, before any theme override.
// Public fields may be changed by controls on the UI thread.
type Controller struct {
	Mode   Mode   `json:"mode"`
	Preset string `json:"preset"`
	// Radius is in DIPs. Zero keeps the theme default; RadiusOverride can
	// explicitly select square corners.
	Radius         float32  `json:"radius,omitempty"`
	RadiusOverride *float32 `json:"radiusOverride,omitempty"`
	ReduceMotion   bool     `json:"reduceMotion,omitempty"`
	last           Tokens
	native         *ui.Theme
}

func (s *Controller) Resolve(systemDark bool) *Tokens {
	dark := systemDark
	switch s.Mode {
	case LightMode:
		dark = false
	case DarkMode:
		dark = true
	}
	t := Preset(s.Preset, dark)
	if s.Radius > 0 {
		t.Radius = s.Radius
	}
	if s.RadiusOverride != nil {
		t.Radius = max(0, *s.RadiusOverride)
	}
	return t
}

func (s *Controller) Apply(c *ui.Context) {
	dark := false
	if c.Theme() != nil {
		dark = c.Theme().Dark
	}
	t := s.Resolve(dark)
	p := c.Preferences()
	if p.TextScale > 0 {
		t.FontSize *= p.TextScale
	}
	if p.HighContrast {
		t.Border = Mix(t.Border, t.Foreground, .45)
		t.Input = t.Border
		t.MutedForeground = Mix(t.MutedForeground, t.Foreground, .4)
		t.Ring = t.Foreground
		t.SidebarBorder = t.Border
		t.SidebarRing = t.Ring
	}
	if s.native == nil || s.last != *t {
		s.last = *t
		s.native = t.Theme()
	}
	c.SetTheme(s.native)
	if s.ReduceMotion {
		c.DisableAnimations()
	}
}

// Toggle chooses the opposite of the currently resolved appearance.
func (s *Controller) Toggle(systemDark bool) {
	if s.Resolve(systemDark).Dark {
		s.Mode = LightMode
	} else {
		s.Mode = DarkMode
	}
}

func (s *Controller) validate() error {
	if s.Mode != "" && s.Mode != System && s.Mode != LightMode && s.Mode != DarkMode {
		return fmt.Errorf("unknown appearance %q", s.Mode)
	}
	if s.Preset != "" {
		known := false
		for _, name := range Presets() {
			if name == s.Preset {
				known = true
			}
		}
		if !known {
			return fmt.Errorf("unknown preset %q", s.Preset)
		}
	}
	validRadius := func(v float32) bool {
		return v >= 0 && v <= 128 && !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0)
	}
	if !validRadius(s.Radius) || s.RadiusOverride != nil && !validRadius(*s.RadiusOverride) {
		return fmt.Errorf("radius must be between 0 and 128 DIPs")
	}
	return nil
}

// Save atomically replaces a preferences file so interruption cannot leave
// partial JSON. The caller chooses when and where preferences are persisted.
func (s *Controller) Save(path string) error {
	if err := s.validate(); err != nil {
		return err
	}
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".theme-*.json")
	if err != nil {
		return err
	}
	name := f.Name()
	defer os.Remove(name)
	if _, err = f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}

// Load rejects unknown fields and malformed preferences before changing state.
func (s *Controller) Load(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	var next Controller
	d := json.NewDecoder(io.LimitReader(f, 64*1024))
	d.DisallowUnknownFields()
	if err = d.Decode(&next); err != nil {
		return err
	}
	var extra any
	if err = d.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("multiple preferences objects")
		}
		return err
	}
	if err = next.validate(); err != nil {
		return err
	}
	*s = next
	return nil
}
