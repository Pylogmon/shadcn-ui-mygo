package fonts_test

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/fonts"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestPersianCalendarUsesItsOptionalSourceFont(t *testing.T) {
	var manifest struct {
		Extras []struct {
			ID                 string
			Sources, Documents []struct {
				Asset, GitBlob, SHA256 string
				Bytes                  int
			}
		}
	}
	raw, err := os.ReadFile("manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Extras) != 1 || manifest.Extras[0].ID != "vazirmatn" {
		t.Fatal("missing fixed calendar-example font")
	}
	family := manifest.Extras[0]
	for _, s := range append(family.Sources, family.Documents...) {
		data, err := os.ReadFile(s.Asset)
		if err != nil {
			t.Fatal(err)
		}
		sha := sha256.Sum256(data)
		blob := sha1.Sum(append([]byte(fmt.Sprintf("blob %d\x00", len(data))), data...))
		if len(data) != s.Bytes || hex.EncodeToString(sha[:]) != s.SHA256 || hex.EncodeToString(blob[:]) != s.GitBlob {
			t.Fatal("changed original calendar font or notice", s.Asset)
		}
	}
	font := fonts.Vazirmatn()
	if err := font.Register(); err != nil {
		t.Fatal(err)
	}
	registered := false
	for _, f := range ui.FontFamilies() {
		registered = registered || f.Name == font.Family() && f.Registered
	}
	if !registered || len(font.Licenses()) != 1 || len(fonts.Catalog()) != 26 {
		t.Fatal("optional font was not registered or changed the Create catalog")
	}
	location := time.FixedZone("Tehran", 3*3600+30*60)
	selected := time.Date(2025, 6, 12, 14, 35, 0, 0, location)
	month := selected
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		theme.FontScope(c, theme.FontRoles{Body: font.Family()}).Column().Fill().Padding(16).Gap(16).Background(ui.RGB(255, 255, 255)).TextColor(ui.RGB(0, 0, 0)).Children(func() {
			ui.Row(c).Gap(8).Children(func() {
				for _, weight := range []int{400, 500, 600, 800} {
					ui.Text(c, "فارسی ۱۴۰۴").Width(140).FontSize(20).LineHeight(1).FontWeight(weight).Role(ui.RoleHeading).Label(fmt.Sprintf("Persian %d", weight))
				}
			})
			display.CalendarWith(c, &selected, display.CalendarProps{System: display.CalendarPersian, Location: location, Month: &month, AutoFocus: true})
		})
	}, 640, 500, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	previous := ""
	for _, weight := range []int{400, 500, 600, 800} {
		ink, hash, width := fontPixels(t, r, fmt.Sprintf("Persian %d", weight))
		if ink == 0 || width == 0 || hash == previous {
			t.Fatal("source font weight did not produce actual distinct Persian glyphs", weight)
		}
		previous = hash
	}
	if !r.HasText("خرداد ۱۴۰۴") {
		t.Fatal("Persian calendar caption missing")
	}
	if err := r.Click("جمعه ۲۳-ام خرداد ۱۴۰۴"); err != nil {
		t.Fatal(err)
	}
	if selected.Day() != 13 || selected.Hour() != 14 || selected.Minute() != 35 || selected.Location() != location {
		t.Fatal("source font changed native date selection or the selected civil time")
	}
	writeImage(t, r.Image(), "persian-vazirmatn-native.png")
}
