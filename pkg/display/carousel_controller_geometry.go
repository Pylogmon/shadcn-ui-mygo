package display

import (
	"math"
	"slices"
)

// Engine geometry consumes measured native rectangles in the user's selected
// order. It is deliberately separate from the live layout: disabled watchers
// freeze this engine snapshot, not the native tree that continues to lay out.
type carouselEngineGeometry struct {
	view, viewport, total float64
	starts, sizes, gaps   []float64
	raw, aligned, snaps   []float64
	registry              [][]int
	renderIndices         []int
	loop                  bool
}

func carouselMeasureEngine(viewport, spacing float64, starts, sizes []float64, renderIndices []int, options CarouselResolvedOptions) (carouselEngineGeometry, error) {
	g := carouselEngineGeometry{view: viewport + spacing, viewport: viewport, starts: slices.Clone(starts), sizes: slices.Clone(sizes), renderIndices: slices.Clone(renderIndices)}
	if len(sizes) == 0 || g.view <= 0 {
		return g, nil
	}
	startGap := math.Abs(starts[0])
	g.gaps = make([]float64, len(sizes))
	for i, size := range sizes {
		gap := size
		if i == 0 {
			gap += startGap
		} else if i+1 < len(sizes) {
			gap = starts[i+1] - starts[i]
		}
		g.gaps[i] = math.Abs(gap)
		g.raw = append(g.raw, -math.Abs(starts[i]))
	}
	g.total = -g.raw[len(g.raw)-1] + g.gaps[len(g.gaps)-1]
	measure := func(loop bool) ([][]int, []float64, error) {
		var groups [][]int
		if !options.SlidesToScroll.Auto {
			count := max(1, options.SlidesToScroll.Count)
			for i := 0; i < len(sizes); i += count {
				var group []int
				for j := i; j < min(i+count, len(sizes)); j++ {
					group = append(group, j)
				}
				groups = append(groups, group)
			}
		} else {
			first := 0
			for i := range sizes {
				leading := float64(0)
				if !loop && first == 0 {
					leading = startGap
				}
				chunk := math.Abs(starts[i] + sizes[i] - starts[first] + leading)
				if i > 0 && chunk > g.view+2 {
					var group []int
					for j := first; j < i; j++ {
						group = append(group, j)
					}
					groups = append(groups, group)
					first = i
				}
				if i == len(sizes)-1 {
					var group []int
					for j := first; j <= i; j++ {
						group = append(group, j)
					}
					groups = append(groups, group)
				}
			}
		}
		var aligned []float64
		for index, group := range groups {
			first, last := group[0], group[len(group)-1]
			size := math.Abs(starts[last] + sizes[last] - starts[first])
			offset := (g.view - size) / 2
			if options.Align.Resolver != nil {
				offset = options.Align.Resolver.resolve(CarouselAlignInput{ViewSize: g.view, SnapSize: size, Index: index})
			} else if options.Align.Mode == CarouselAlignStart {
				offset = 0
			} else if options.Align.Mode == CarouselAlignEnd {
				offset = g.view - size
			}
			if math.IsNaN(offset) || math.IsInf(offset, 0) {
				return nil, nil, &CarouselOptionError{Field: "Align.Resolver", Code: CarouselInvalidOption}
			}
			aligned = append(aligned, g.raw[first]+offset)
		}
		return groups, aligned, nil
	}
	groups, aligned, err := measure(options.Loop)
	if err != nil {
		return g, err
	}
	g.aligned = aligned
	g.loop = options.Loop && g.canLoop()
	if options.Loop && !g.loop {
		groups, aligned, err = measure(false)
		if err != nil {
			return g, err
		}
		g.aligned = aligned
	}
	minimum, maximum := 0, len(aligned)
	if g.loop || options.Contain == CarouselUncontained {
		g.snaps = slices.Clone(aligned)
	} else if g.total <= g.view+2 {
		g.snaps = []float64{0}
	} else {
		bounded := make([]float64, len(aligned))
		lower := g.view - g.total
		for i, snap := range aligned {
			snap = max(lower, min(0, snap))
			if i == 0 {
				snap = 0
			} else if i == len(aligned)-1 {
				snap = lower
			} else if math.Abs(lower-snap) <= 1 {
				snap = lower
			} else if math.Abs(snap) <= 1 {
				snap = 0
			}
			bounded[i] = math.Round(snap*1000) / 1000
		}
		if options.Contain == CarouselKeepSnaps {
			g.snaps = bounded
		} else {
			for i, snap := range bounded {
				if snap == bounded[0] {
					minimum = i
				}
				if snap == bounded[len(bounded)-1] && maximum == len(bounded) {
					maximum = i + 1
				}
			}
			g.snaps = slices.Clone(bounded[minimum:maximum])
		}
	}
	if len(g.snaps) == 1 {
		var group []int
		for i := range sizes {
			group = append(group, i)
		}
		g.registry = [][]int{group}
	} else if g.loop || options.Contain == CarouselUncontained || options.Contain == CarouselKeepSnaps {
		g.registry = groups
	} else {
		g.registry = make([][]int, maximum-minimum)
		for i, group := range groups[minimum:maximum] {
			g.registry[i] = slices.Clone(group)
		}
		if len(g.registry) > 0 {
			last := g.registry[0][len(g.registry[0])-1]
			g.registry[0] = nil
			for i := 0; i <= last; i++ {
				g.registry[0] = append(g.registry[0], i)
			}
			lastGroup := len(g.registry) - 1
			first := g.registry[lastGroup][0]
			g.registry[lastGroup] = nil
			for i := first; i < len(sizes); i++ {
				g.registry[lastGroup] = append(g.registry[lastGroup], i)
			}
		}
	}
	return g, nil
}

func (g carouselEngineGeometry) loopPointIndices() (start, end []int) {
	if len(g.aligned) == 0 {
		return nil, nil
	}
	gap := g.aligned[0]
	for i := len(g.sizes) - 1; i >= 0 && gap > 0; i-- {
		start = append(start, i)
		gap -= g.gaps[i]
	}
	gap = g.view - g.aligned[0] - 1
	for i := 0; i < len(g.sizes) && gap > 0; i++ {
		end = append(end, i)
		gap -= g.gaps[i]
	}
	return
}

func (g carouselEngineGeometry) canLoop() bool {
	start, end := g.loopPointIndices()
	for _, i := range append(start, end...) {
		remaining := g.view
		for j, size := range g.gaps {
			if i != j {
				remaining -= size
			}
		}
		if remaining > .1 {
			return false
		}
	}
	return true
}

func (g carouselEngineGeometry) shift(index int, location float64) float64 {
	if !g.loop {
		return 0
	}
	start, end := g.loopPointIndices()
	if slices.Contains(start, index) {
		point := g.raw[index] - g.sizes[index] + .5 + g.total
		if location > point {
			return -g.total
		}
		return 0
	}
	if slices.Contains(end, index) {
		point := g.raw[index] + g.view - .5 - g.total
		if location > point {
			return 0
		}
		return g.total
	}
	return 0
}

func (g carouselEngineGeometry) limit() (minimum, maximum float64) {
	if len(g.snaps) == 0 {
		return 0, 0
	}
	maximum = g.snaps[0]
	minimum = g.snaps[len(g.snaps)-1]
	if g.loop {
		minimum = maximum - g.total
	}
	return
}

func (g carouselEngineGeometry) progress(location float64) float64 {
	minimum, maximum := g.limit()
	length := math.Abs(maximum - minimum)
	if length == 0 {
		return 0
	}
	return (location - maximum) / -length
}

// Counter preserves the pinned source's abs((index+count)%count) behavior even
// for large negative indices; it is not a replacement mathematical modulo.
func (g carouselEngineGeometry) counter(index int) int {
	if len(g.snaps) == 0 {
		return 0
	}
	if !g.loop {
		return max(0, min(index, len(g.snaps)-1))
	}
	result := (len(g.snaps) + index) % len(g.snaps)
	if result < 0 {
		result = -result
	}
	return result
}

func (g carouselEngineGeometry) shortcut(distance float64, direction int) float64 {
	if !g.loop {
		return distance
	}
	targets := []float64{distance, distance + g.total, distance - g.total}
	best := math.Inf(1)
	found := false
	for _, target := range targets {
		if direction != 0 && carouselSign(target) != direction {
			continue
		}
		if !found || math.Abs(target) < math.Abs(best) {
			best, found = target, true
		}
	}
	if found {
		return best
	}
	return targets[2] - g.total
}

func carouselSign(value float64) int {
	if value > 0 {
		return 1
	}
	if value < 0 {
		return -1
	}
	return 0
}
