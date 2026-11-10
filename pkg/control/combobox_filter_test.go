package control

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestComboboxFilterMatchesFixedOfficialIntlOracle(t *testing.T) {
	raw, err := os.ReadFile("testdata/shadcn-combobox/filter-native/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var oracle struct {
		Source, SourceSHA256, AriaSource, AriaSHA256 string
		Node, ICU, Unicode, CLDR                     string
		Cases                                        []struct {
			Locale, Label, Query string
			Contains             bool
		}
	}
	if err := json.Unmarshal(raw, &oracle); err != nil {
		t.Fatal(err)
	}
	if oracle.Node != "v26.12.0" || oracle.ICU != "78.3" || oracle.CLDR != "48.0" || len(oracle.Cases) != 64 {
		t.Fatal("changed recorded official Intl reference")
	}
	for _, original := range []struct{ path, hash string }{{"testdata/shadcn-combobox/npm/base-ui-react/internals/filter.js", oracle.SourceSHA256}, {"testdata/shadcn-combobox/npm/base-ui-react/combobox/root/AriaCombobox.js", oracle.AriaSHA256}} {
		data, err := os.ReadFile(original.path)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(data)
		if hex.EncodeToString(hash[:]) != original.hash {
			t.Fatal("official oracle source changed: " + original.path)
		}
	}
	for _, c := range oracle.Cases {
		items := comboEntries(ComboboxWithProps{Locale: c.Locale, Options: []Option{{Value: "row", Label: c.Label}}}, c.Query, true, nil)
		// A deliberately empty accessible label uses Option.Value in the kit.
		// Give that source case an empty value too so its label is still empty.
		if c.Label == "" {
			items = comboEntries(ComboboxWithProps{Locale: c.Locale, Options: []Option{{Value: "", Label: ""}}}, c.Query, true, nil)
		}
		if got := len(items) == 1; got != c.Contains {
			t.Errorf("locale %s label %q query %q: native %v, official Intl %v", c.Locale, c.Label, c.Query, got, c.Contains)
		}
	}
}

// One actual input/selection flow exercises fixed UTF-16 windows, punctuation
// and a trimmed blank multi-query. Its oracle vectors are not a platform matrix.
func TestComboboxUnicodeQueryKeepsNativeSelectionAndBlankMultiBehavior(t *testing.T) {
	selected, query := ComboboxValue{}, ""
	multiple, multiQuery := []string{}, ""
	locale := "de"
	r := tester(func(c *ui.Context) {
		ComboboxWith(c, &selected, &query, ComboboxWithProps{Label: "Localized search", Locale: locale, Options: []Option{{Value: "sharp", Label: "Straße"}, {Value: "ascii", Label: "Strassek"}, {Value: "punct", Label: "co-op"}, {Value: "zero", Label: "..."}}})
		MultiComboboxWith(c, &multiple, &multiQuery, ComboboxWithProps{Label: "Languages", AutoHighlight: true, Options: []Option{{Value: "go", Label: "Go"}, {Value: "rust", Label: "Rust"}}})
	})
	openSelectionCombo(t, r, "Localized search")
	r.Type("strasse")
	r.Frame()
	if query != "strasse" {
		t.Fatal("filtering changed the actual editor's raw query")
	}
	if _, ok := r.AccessByRole(ui.RoleOption, "Straße"); ok {
		t.Error("the source's fixed UTF-16 query window was widened to match Straße")
	}
	if _, ok := r.AccessByRole(ui.RoleOption, "Strassek"); !ok {
		t.Fatal("missing same-length native query result")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	if selected != (ComboboxValue{Value: "ascii", Valid: true}) || query != "Strassek" {
		t.Errorf("native keyboard chose a different serialized value: %+v / %q", selected, query)
	}
	locale = "en"
	openSelectionCombo(t, r, "Localized search")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("coop!")
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleOption, "co-op"); !ok {
		t.Error("ignorePunctuation did not affect the actual native result collection")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	if selected != (ComboboxValue{Value: "punct", Valid: true}) || query != "co-op" {
		t.Errorf("punctuation match did not commit its native option: %+v / %q", selected, query)
	}
	openSelectionCombo(t, r, "Localized search")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("---")
	r.Frame()
	if _, ok := r.AccessByRole(ui.RoleOption, "..."); !ok {
		t.Error("same-length collation-empty query did not retain its native result")
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Advance(120 * time.Millisecond)
	if selected != (ComboboxValue{Value: "zero", Valid: true}) || query != "..." {
		t.Errorf("collation-empty query did not select its native option: %+v / %q", selected, query)
	}
	openSelectionCombo(t, r, "Languages")
	r.Type("\uFEFF")
	r.Frame()
	if multiQuery != "\uFEFF" {
		t.Fatal("the native editor did not retain the raw BOM query")
	}
	if active := comboBehaviorActive(r); active != "" {
		t.Errorf("ECMAScript-trimmed empty query auto-highlighted %q", active)
	}
	r.Key(0, ui.KeyDown)
	r.Key(0, ui.KeyEnter)
	r.Frame()
	if !slices.Equal(multiple, []string{"go"}) || !selectionComboInput(t, r, "Languages").Expanded {
		t.Errorf("trimmed blank multi-query lost source selection/open policy: %v expanded=%v", multiple, selectionComboInput(t, r, "Languages").Expanded)
	}
}
