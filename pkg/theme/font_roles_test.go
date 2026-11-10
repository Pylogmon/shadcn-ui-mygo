package theme_test

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestFontRolesFixedSourceArchive(t *testing.T) {
	root := "testdata/shadcn-font-roles"
	raw, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		Commit string
		Files  []struct{ Path, SHA256 string }
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "8ab27715521cbd4f87ae6886cd989cbb678f6de7" || len(manifest.Files) != 12 {
		t.Fatal("incorrect fixed font role references")
	}
	for _, file := range manifest.Files {
		raw, err := os.ReadFile(filepath.Join(root, file.Path))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(raw)
		if hex.EncodeToString(hash[:]) != file.SHA256 {
			t.Errorf("changed font role source %s", file.Path)
		}
	}
}

func fontAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, name)
	if !ok || n.Offscreen {
		t.Fatalf("missing font fixture %v %q", role, name)
	}
	return n
}
func fontInkWidth(im image.Image, rect ui.Rect) int {
	x0, x1 := int(rect.X), int(rect.X+rect.W+1)
	y0, y1 := int(rect.Y), int(rect.Y+rect.H+1)
	first, last := x1, x0-1
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			r, g, b, a := im.At(x, y).RGBA()
			if a > 0 && r < 32000 && g < 32000 && b < 32000 {
				first, last = min(first, x), max(last, x)
			}
		}
	}
	return max(0, last-first+1)
}

func TestFontRolesSourceTitlesAndLiveNativeEditor(t *testing.T) {
	const sample = "iiiiiiiiiiiiiiii"
	fonts := theme.FontRoles{Body: "monospace", Heading: "Georgia"}
	controller := theme.Controller{Fonts: &fonts}
	draft, open := "Keep this native draft", false
	var nested, after theme.FontRoles
	var selection [2]int
	r := ui.NewTesterWith(func(c *ui.Context) {
		controller.Apply(c)
		ui.Column(c).Fill().Padding(20).Gap(12).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(0, 0, 0)).Children(func() {
			ui.Row(c).Children(func() { display.CardTitle(c, sample).Role(ui.RoleHeading).Label("Source card title") })
			ui.Row(c).Children(func() { ui.Text(c, sample).FontSize(16).FontWeight(500).Role(ui.RoleHeading).Label("Body reference") })
			padding := float32(0)
			display.EmptyWith(c, display.EmptyProps{Padding: &padding}, display.EmptySlots{Title: func(child *ui.Context) {
				ui.Text(child, sample).Role(ui.RoleHeading).Label("Custom empty title")
			}})
			ui.Row(c).Children(func() {
				ui.Text(c, sample).Font("Georgia").FontSize(18).FontWeight(500).Role(ui.RoleHeading).Label("Custom title reference")
			})
			theme.Scope(c, theme.Preset("rose", false)).Children(func() {
				theme.MenuScope(c, theme.Preset("blue", false), theme.MenuOptions{Color: theme.MenuInverted}).Children(func() {
					nested = theme.FontsOf(c)
				})
			})
			after = theme.FontsOf(c)
			editor := ui.TextInputBase(c, &draft).Label("Font editor")
			a, b := editor.TextSelection()
			selection = [2]int{a, b}
			if ui.Button(c, "Open heading dialog").Clicked() {
				open = true
			}
			// The close glyph otherwise overlaps the heading's stretched AX box;
			// omit it from this native text-pixel fixture and dismiss with Escape.
			overlay.Dialog(c, &open, overlay.DialogProps{Title: sample, Description: "The description keeps the body role.", HideClose: true}, func() {
				ui.Row(c).Children(func() {
					ui.Text(c, sample).Font("monospace").FontSize(18).FontWeight(500).Role(ui.RoleHeading).Label("Dialog reference")
				})
			})
		})
	}, 700, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	ink := func(name string) int { return fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, name).Bounds) }
	if nested != fonts.Normalized() || after != fonts.Normalized() {
		t.Fatalf("color/menu scopes lost role values: nested%+v after%+v", nested, after)
	}
	if heading, body := ink("Source card title"), ink("Body reference"); heading == 0 || float64(heading) > float64(body)*.75 {
		t.Fatalf("actual title pixels did not use a distinct serif: heading%d body%d", heading, body)
	}
	if title, reference := ink("Custom empty title"), ink("Custom title reference"); title < reference-1 || title > reference+1 {
		t.Fatalf("custom Empty title did not inherit wrapper heading: title%d Georgia reference%d", title, reference)
	}
	fonts.Heading = theme.FontInherit
	r.Frame()
	if heading, body := ink("Source card title"), ink("Body reference"); heading != body {
		t.Fatalf("inherit retained a stale explicit heading font: %d/%d", heading, body)
	}
	if err := r.Click("Font editor"); err != nil {
		t.Fatal(err)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Frame()
	before := selection
	fonts.Body, fonts.Heading = "Georgia", "monospace"
	r.Frame()
	if draft != "Keep this native draft" || selection != before || !r.Focused("Font editor") || after != fonts.Normalized() {
		t.Fatalf("same-pointer role change replaced native editor/state: %q %v/%v %+v", draft, selection, before, after)
	}
	if heading, body := ink("Source card title"), ink("Body reference"); body == 0 || float64(body) > float64(heading)*.75 {
		t.Fatalf("live roles did not change actual glyphs: heading%d body%d", heading, body)
	}
	if err := r.Click("Open heading dialog"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	// The real modal correctly removes the underlying card from accessibility.
	if heading := ink(sample); heading != ink("Dialog reference") {
		t.Fatalf("dialog source title did not use heading role: %d", heading)
	}
	r.Key(0, ui.KeyEscape)
	r.Frame()
	if open || !r.Focused("Open heading dialog") {
		t.Fatal("heading role styling changed dialog dismissal/focus")
	}
}

func TestFontScopeNativePropertiesAndDefaultLocalInheritance(t *testing.T) {
	native := *ui.LightTheme()
	native.Font, native.FontSize, native.TooltipDelay = "monospace", 17, 123*time.Millisecond
	var inside, outside, nativeRoles theme.FontRoles
	var preserved ui.Theme
	r := ui.NewTester(func(c *ui.Context) {
		c.SetTheme(&native)
		ui.Column(c).Fill().Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(0, 0, 0)).Children(func() {
			theme.FontScope(c, theme.FontRoles{Body: "Georgia", Heading: "monospace"}).Children(func() {
				inside, preserved = theme.FontsOf(c), *c.Theme()
				theme.NativeScope(c).Children(func() { nativeRoles = theme.FontsOf(c) })
			})
			outside = theme.FontsOf(c)
			ui.Column(c).Font("Georgia").Children(func() {
				ui.Row(c).Children(func() { display.CardTitle(c, "iiiiiiiiiiiiiiii").Role(ui.RoleHeading).Label("Local title") })
				ui.Row(c).Children(func() {
					ui.Text(c, "iiiiiiiiiiiiiiii").FontSize(16).FontWeight(500).Role(ui.RoleHeading).Label("Local reference")
				})
			})
			theme.FontScope(c, theme.FontRoles{Body: "monospace", Heading: theme.FontSystem}).Children(func() {
				ui.Row(c).Children(func() { display.EmptyTitle(c, "iiiiiiiiiiiiiiii").Role(ui.RoleHeading).Label("System heading") })
			})
			// Independent system-font context, rather than another Font("")
			// under the mono theme: the two APIs must not share a false oracle.
			theme.FontScope(c, theme.FontRoles{Body: theme.FontSystem}).Children(func() {
				ui.Row(c).Children(func() {
					ui.Text(c, "iiiiiiiiiiiiiiii").FontSize(18).FontWeight(500).Role(ui.RoleHeading).Label("System reference")
				})
			})
			ui.Row(c).Children(func() {
				ui.Text(c, "iiiiiiiiiiiiiiii").FontSize(18).FontWeight(500).Role(ui.RoleHeading).Label("Mono reference")
			})
		})
	}, 700, 350)
	if inside != (theme.FontRoles{Body: "Georgia", Heading: "monospace"}) || outside != (theme.FontRoles{Body: "monospace", Heading: theme.FontInherit}) || nativeRoles != (theme.FontRoles{Body: "Georgia", Heading: theme.FontInherit}) {
		t.Fatalf("scope/native role boundary: %+v %+v %+v", inside, outside, nativeRoles)
	}
	preserved.Font = native.Font
	if preserved != native {
		t.Fatal("font-only scope changed non-font native properties")
	}
	for _, pair := range [][2]string{{"Local title", "Local reference"}, {"System heading", "System reference"}} {
		a := fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, pair[0]).Bounds)
		b := fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, pair[1]).Bounds)
		if a == 0 || a != b {
			t.Fatalf("default inheritance/explicit system pixels %v: %d/%d", pair, a, b)
		}
	}
	system := fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, "System heading").Bounds)
	mono := fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, "Mono reference").Bounds)
	t.Logf("native ink widths: explicit system heading=%d system-body reference=%d mono reference=%d", system, fontInkWidth(r.Image(), fontAccess(t, r, ui.RoleHeading, "System reference").Bounds), mono)
	if system == mono {
		t.Fatal("explicit system heading still inherits the mono body")
	}
}
