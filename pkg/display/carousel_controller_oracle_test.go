package display

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"math"
	"slices"
	"testing"
)

func TestCarouselControllerIndependentSourceGeometryAndCounter(t *testing.T) {
	var fixture struct {
		SourceSHA256 string
		Options      []struct {
			Name                 string
			Sizes                []float64
			Grouping             json.RawMessage
			Align                string
			ContainScroll        json.RawMessage
			Snaps, Normalized    []float64
			Registry             [][]int
			StartIndex, Selected int
			Loop                 bool
			Location, Progress   float64
		}
		Selectors []struct {
			Name              string
			Order             []int
			Snaps, Normalized []float64
			Registry          [][]int
		}
	}
	data := carouselArchiveRead(t, "controller/controller-oracle.json")
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	source := sha256.Sum256(carouselArchiveRead(t, "npm/embla-carousel/cjs/embla-carousel.cjs.js"))
	if fixture.SourceSHA256 != hex.EncodeToString(source[:]) {
		t.Fatal("controller oracle was not produced by the pinned source")
	}
	compare := func(label string, got, want []float64) {
		t.Helper()
		if len(got) != len(want) {
			t.Fatalf("%s got=%v want=%v", label, got, want)
		}
		for i := range got {
			if math.Abs(got[i]-want[i]) > .001 {
				t.Fatalf("%s got=%v want=%v", label, got, want)
			}
		}
	}
	for _, input := range fixture.Options {
		sizes := input.Sizes
		if sizes == nil {
			sizes = []float64{300, 300, 300, 300, 300}
		}
		starts := make([]float64, len(sizes))
		indices := make([]int, len(sizes))
		for i := range sizes {
			indices[i] = i
			if i > 0 {
				starts[i] = starts[i-1] + sizes[i-1]
			}
		}
		options := carouselResolveOptions(CarouselOptions{Loop: CarouselValue(input.Loop)}, CarouselEnvironment{})
		options.Align.Mode = CarouselAlignStart
		if input.Name == "" {
			options.Align.Mode = map[string]CarouselAlign{"start": CarouselAlignStart, "center": CarouselAlignCenter, "end": CarouselAlignEnd}[input.Align]
			if string(input.Grouping) == `"auto"` {
				options.SlidesToScroll = CarouselSlideGroup{Auto: true}
			} else {
				if err := json.Unmarshal(input.Grouping, &options.SlidesToScroll.Count); err != nil {
					t.Fatal(err)
				}
			}
			if string(input.ContainScroll) == "false" {
				options.Contain = CarouselUncontained
			} else {
				var contain string
				json.Unmarshal(input.ContainScroll, &contain)
				if contain == "keepSnaps" {
					options.Contain = CarouselKeepSnaps
				}
			}
		}
		g, err := carouselMeasureEngine(600, 0, starts, sizes, indices, options)
		if err != nil {
			t.Fatal(err)
		}
		if input.Name == "" {
			compare("snaps", g.snaps, input.Snaps)
			for i, group := range g.registry {
				if !slices.Equal(group, input.Registry[i]) {
					t.Fatalf("source registry got=%v want=%v", g.registry, input.Registry)
				}
			}
		}
		if input.Name == "startIndex" {
			if selected := g.counter(input.StartIndex); selected != input.Selected {
				t.Fatalf("source counter(%d,loop=%v)=%d want%d", input.StartIndex, input.Loop, selected, input.Selected)
			}
		}
		if input.Name == "overscroll-progress" {
			if got := g.progress(input.Location); math.Abs(got-input.Progress) > .0001 {
				t.Fatalf("source progress(%v)=%v want%v", input.Location, got, input.Progress)
			}
		}
		if input.Normalized != nil {
			var actual []float64
			for _, snap := range g.snaps {
				actual = append(actual, g.progress(snap))
			}
			compare("normalized", actual, input.Normalized)
		}
	}
	for _, input := range fixture.Selectors {
		if len(input.Order) == 0 {
			continue
		}
		var starts, sizes []float64
		for _, index := range input.Order {
			starts = append(starts, float64(index)*300)
			sizes = append(sizes, 300)
		}
		options := carouselResolveOptions(CarouselOptions{Align: &CarouselAlignment{Mode: CarouselAlignStart}}, CarouselEnvironment{})
		g, err := carouselMeasureEngine(600, 0, starts, sizes, input.Order, options)
		if err != nil {
			t.Fatal(err)
		}
		compare("selected node geometry", g.snaps, input.Snaps)
		for i, group := range g.registry {
			if !slices.Equal(group, input.Registry[i]) {
				t.Fatalf("selected node registry got=%v want=%v", g.registry, input.Registry)
			}
		}
	}
}
