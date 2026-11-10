package fonts_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/fonts"
	fontall "github.com/pylogmon/shadcn-ui-mygo/pkg/fonts/all"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestOriginalCreateFontAssets(t *testing.T) {
	raw, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	type original struct {
		Path, Asset, SHA256, GitBlob string
		Bytes                        int
	}
	var m struct {
		Commit   string
		Families []struct {
			ID                 string
			Sources, Documents []original
			Derived            []struct {
				Asset, SHA256 string
				Bytes         int
			}
		}
		ShadcnDefinitions, ShadcnPicker original
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m.Commit != "5bb28733cec313b63fbd415b574b5ae35c8aa8b8" || len(m.Families) != 26 {
		t.Fatal("incorrect fixed official font collection")
	}
	check := func(s original) {
		t.Helper()
		data, err := os.ReadFile(filepath.FromSlash(s.Asset))
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		blob := sha1.Sum(append([]byte(fmt.Sprintf("blob %d\x00", len(data))), data...))
		if len(data) != s.Bytes || hex.EncodeToString(hash[:]) != s.SHA256 || hex.EncodeToString(blob[:]) != s.GitBlob {
			t.Errorf("changed original font or source %s", s.Path)
		}
	}
	check(m.ShadcnDefinitions)
	check(m.ShadcnPicker)
	catalog := fonts.Catalog()
	for i, f := range m.Families {
		if catalog[i].ID != f.ID {
			t.Fatalf("source picker order changed at %d", i)
		}
		for _, s := range append(f.Sources, f.Documents...) {
			check(s)
		}
		for _, d := range f.Derived {
			data, err := os.ReadFile(filepath.FromSlash(d.Asset))
			if err != nil {
				t.Fatal(err)
			}
			hash := sha256.Sum256(data)
			if len(data) != d.Bytes || hex.EncodeToString(hash[:]) != d.SHA256 {
				t.Errorf("changed reproducible native instance %s", d.Asset)
			}
		}
	}
}

type renderedWeight struct {
	ID     string
	Weight int
	Ink    uint64
	Hash   string
}

func fontPixels(t *testing.T, r *ui.Tester, label string) (uint64, string, int) {
	t.Helper()
	a, ok := r.AccessByRole(ui.RoleHeading, label)
	if !ok || a.Offscreen || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("font sample is not present on the native surface: %s", label)
	}
	im := r.Image()
	var ink uint64
	var pixels []byte
	first, last := 1000000, -1
	x0, y0 := int(math.Floor(float64(a.Bounds.X))), int(math.Floor(float64(a.Bounds.Y)))
	w, h := int(math.Ceil(float64(a.Bounds.W))), int(math.Ceil(float64(a.Bounds.H)))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			rr, gg, bb, aa := im.At(x0+x, y0+y).RGBA()
			ink += uint64(65535-rr) + uint64(65535-gg) + uint64(65535-bb)
			pixels = append(pixels, byte(rr>>8), byte(gg>>8), byte(bb>>8), byte(aa>>8))
			if rr < 32000 && gg < 32000 && bb < 32000 {
				first, last = min(first, x), max(last, x)
			}
		}
	}
	hash := sha256.Sum256(pixels)
	return ink, hex.EncodeToString(hash[:]), max(0, last-first+1)
}

// One native workflow registers the source collection, checks its actual used
// weights, then changes the application's font roles while editing a real field.
// It does not assert platform/AT acceptance from these headless native images.
func TestCreateFontFamiliesNativeWeightsAndEditor(t *testing.T) {
	if err := fontall.Register(); err != nil {
		t.Fatal(err)
	}
	if err := fonts.Inter().Register(); err != nil {
		t.Fatal(err)
	}
	registered := map[string]bool{}
	for _, family := range ui.FontFamilies() {
		registered[family.Name] = family.Registered
	}
	all := fontall.Families()
	for _, f := range all {
		if !registered[f.Family()] || len(f.Licenses()) == 0 {
			t.Fatalf("font %s is not registered with a distributable notice", f.Info().ID)
		}
	}
	roles := theme.FontRoles{Body: fonts.Inter().Family(), Heading: fonts.PlayfairDisplay().Family()}
	usedWeights := []int{400, 500, 600, 700, 800}
	draft := "Native font draft"
	var handle ui.Handle
	var selection [2]int
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		theme.FontScope(c, roles).Fill().Padding(16).Gap(8).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(0, 0, 0)).Children(func() {
			theme.StyleHeading(c, ui.Text(c, "iiiiiiiiiiii").FontSize(24).LineHeight(1).FontWeight(600)).Role(ui.RoleHeading).Label("Live heading")
			ui.TextInputBase(c, &draft).Bind(&handle).Label("Font draft")
			for _, f := range all {
				info := f.Info()
				ui.Row(c.Key(info.ID)).Height(36).Gap(8).AlignItems(ui.Center).Children(func() {
					ui.Text(c, info.Name).Width(180).Shrink(0).FontSize(13)
					for _, weight := range usedWeights {
						if slices.Contains(info.Weights, weight) {
							ui.Text(c, "Hambg 013").Width(190).Shrink(0).Font(f.Family()).FontSize(20).LineHeight(1).FontWeight(weight).Role(ui.RoleHeading).Label(fmt.Sprintf("%s-%d", info.ID, weight))
						} else {
							ui.Text(c, "—").Width(190).Shrink(0).FontSize(20)
						}
					}
				})
			}
			ui.Text(c, "Hambg 013").Width(190).Shrink(0).Font(fonts.Inter().Family()).FontSize(20).LineHeight(1).FontWeight(400).Italic().Role(ui.RoleHeading).Label("Inter italic")
		})
		selection[0], selection[1] = c.Resolve(handle).TextSelection()
	}, 1380, 1320, ui.TesterOptions{})
	var observations []renderedWeight
	for _, f := range all {
		info := f.Info()
		var previous string
		for _, weight := range usedWeights {
			if !slices.Contains(info.Weights, weight) {
				continue
			}
			ink, hash, width := fontPixels(t, r, fmt.Sprintf("%s-%d", info.ID, weight))
			if ink == 0 || width == 0 || previous == hash {
				t.Errorf("%s weight %d does not produce distinct native glyphs (previous %s, current %s)", info.ID, weight, previous, hash)
			}
			previous = hash
			observations = append(observations, renderedWeight{info.ID, weight, ink, hash})
		}
	}
	_, regular, _ := fontPixels(t, r, "inter-400")
	_, italic, _ := fontPixels(t, r, "Inter italic")
	if regular == italic {
		t.Fatal("the registered original italic face was not selected")
	}
	writeImage(t, r.Image(), "create-fonts-native.png")
	if err := r.Click("Font draft"); err != nil {
		t.Fatal(err)
	}
	r.Key(0, ui.KeyEnd)
	r.Type(" kept")
	r.Key(0, ui.KeyLeft)
	beforeSelection := selection
	_, _, beforeHeading := fontPixels(t, r, "Live heading")
	roles = theme.FontRoles{Body: fonts.GeistMono().Family(), Heading: theme.FontInherit}
	r.Frame()
	_, _, afterHeading := fontPixels(t, r, "Live heading")
	if !r.Focused("Font draft") || draft != "Native font draft kept" || selection != beforeSelection || beforeHeading == afterHeading {
		t.Fatalf("font role changes lost native edit state or did not change the title: focused=%t draft=%q selection=%v/%v heading=%d/%d", r.Focused("Font draft"), draft, selection, beforeSelection, beforeHeading, afterHeading)
	}
	if dir := os.Getenv("SHADCN_FONTS_ARTIFACT_DIR"); dir != "" {
		data, err := json.MarshalIndent(observations, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "native-weights.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func writeImage(t *testing.T, im image.Image, name string) {
	t.Helper()
	dir := os.Getenv("SHADCN_FONTS_ARTIFACT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
}
