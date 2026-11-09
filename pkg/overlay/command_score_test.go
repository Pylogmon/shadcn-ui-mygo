package overlay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestCommandScoreMatchesPinnedJavaScriptOracle(t *testing.T) {
	data, err := os.ReadFile("testdata/cmdk/scores.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		SourceSHA256 string
		Cases        []struct {
			Value, Query string
			Keywords     []string
			Score        float64
		}
	}
	if err := json.Unmarshal(data, &fixtures); err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile("testdata/cmdk/command-score.ts")
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(source)
	if hex.EncodeToString(hash[:]) != fixtures.SourceSHA256 || len(fixtures.Cases) < 2200 {
		t.Fatal("oracle source or corpus changed; regenerate against the pinned reference")
	}
	for i, example := range fixtures.Cases {
		got := CommandScore(example.Value, example.Query, example.Keywords...)
		if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-example.Score) > 1e-14 || got < 0 || got > 1 {
			t.Fatalf("case %d %q / %q aliases=%q: Go=%0.17g JavaScript=%0.17g", i, example.Value, example.Query, example.Keywords, got, example.Score)
		}
	}
}

func TestCommandDefaultFuzzyRankingAndNativeSelection(t *testing.T) {
	query, picked := "OP", ""
	items := []CommandItem{{Key: "deep", Value: "open-deep", Label: "Browse other projects"}, {Key: "open", Value: "route-open", Label: "Open Project"}, {Key: "exact", Value: "route-exact", Label: "OP"}, {Key: "missing", Label: "Archives"}}
	r := ui.NewTester(view(func(c *ui.Context) {
		if result := Command(c, &query, CommandProps{Items: items, AutoFocus: true}); result != "" {
			picked = result
		}
	}), 480, 400)
	requireCommandHighlight(t, r, "OP")
	r.Key(0, ui.KeyDown)
	requireCommandHighlight(t, r, "Open Project")
	r.Key(0, ui.KeyEnter)
	if picked != "route-open" {
		t.Fatalf("fuzzy result lost its caller-owned value: %q", picked)
	}
	query = "OpnPrjct"
	r.Frame()
	requireCommandHighlight(t, r, "Open Project")
	query = "route-open"
	r.Frame()
	requireCommandHighlight(t, r, "Open Project")
	if r.HasText("Archives") {
		t.Fatal("a missing subsequence remained in the result set")
	}
}

func TestCommandEmptySearchKeepsRegistrationOrderAndBypassesCustomFilter(t *testing.T) {
	called := 0
	p := CommandProps{Items: []CommandItem{{Label: "First"}, {Label: "Second"}, {Label: "Third"}}, Filter: func(item CommandItem, query string) float64 {
		called++
		return 0
	}}
	entries := commandEntries(p, "")
	if called != 0 || len(entries) != 3 || entries[0].item.Label != "First" {
		t.Fatalf("empty search filtered or reordered registration order: called=%d entries=%v", called, entries)
	}
	p.NoFilter = true
	entries = commandEntries(p, "unrelated")
	if called != 0 || len(entries) != 3 {
		t.Fatal("NoFilter did not preserve application-owned results")
	}
}

func TestCommandDefaultFuzzyGroupsStayContinuousAndStable(t *testing.T) {
	p := CommandProps{Items: []CommandItem{{Label: "Other project", Group: "Projects"}, {Label: "Open Project", Group: "Projects"}, {Label: "OP", Group: "Short"}, {Label: "Open Profile", Group: "Projects"}, {Label: "OP", Group: "Short"}}}
	entries := commandEntries(p, "OP")
	var names []string
	for _, entry := range entries {
		names = append(names, entry.item.Group+":"+entry.item.Label)
	}
	want := []string{"Short:OP", "Short:OP", "Projects:Open Project", "Projects:Open Profile", "Projects:Other project"}
	if !slices.Equal(names, want) {
		t.Fatalf("default scorer fragmented or destabilized groups: %v", names)
	}
}

func TestCommandFuzzyUngroupedItemsStayBeforeRankedSections(t *testing.T) {
	entries := commandEntries(CommandProps{Items: []CommandItem{
		{Label: "OP", Group: "Best"}, {Label: "Other project"}, {Label: "Open Project", Group: "Projects"},
	}}, "OP")
	if len(entries) != 3 || entries[0].item.Label != "Other project" || entries[1].item.Group != "Best" || entries[2].item.Group != "Projects" {
		t.Fatalf("ungrouped results no longer precede group sections: %v", entries)
	}
}

func TestCommandUnicodeInputsThatOverflowUpstreamRemainFinite(t *testing.T) {
	data, err := os.ReadFile("testdata/cmdk/scores.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Failures []struct {
			Value, Query, Error string
			Keywords            []string
		}
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Failures) == 0 {
		t.Fatal("missing independently reproduced upstream Unicode failure corpus")
	}
	for _, input := range fixture.Failures {
		if input.Error != "RangeError" {
			t.Fatalf("unexpected upstream failure: %+v", input)
		}
		score := CommandScore(input.Value, input.Query, input.Keywords...)
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1 {
			t.Fatalf("Unicode case expansion did not terminate safely: %v", input)
		}
	}
}

func BenchmarkCommandFuzzyTenThousand(b *testing.B) {
	p := CommandProps{Items: make([]CommandItem, 10000)}
	for i := range p.Items {
		p.Items[i] = CommandItem{Value: fmt.Sprint(i), Label: fmt.Sprintf("Open Project %05d", i)}
	}
	b.ReportAllocs()
	for b.Loop() {
		if len(commandEntries(p, "OP99")) == 0 {
			b.Fatal("fuzzy search produced no matches")
		}
	}
}

func TestCommandScoreLongAndImpossibleQueriesRemainFinite(t *testing.T) {
	for _, query := range []string{strings.Repeat("a", 50), strings.Repeat("z", 1000), "🚀你世界"} {
		score := CommandScore(strings.Repeat("a ", 150)+"🚀你好世界", query)
		if math.IsNaN(score) || math.IsInf(score, 0) || score < 0 || score > 1 {
			t.Fatalf("invalid score for long query: %v", score)
		}
	}
}
