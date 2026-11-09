package display

import (
	"math"
	"slices"
)

// CarouselAlign positions each snap's item in the scrolling viewport.
type CarouselAlign uint8

const (
	CarouselAlignCenter CarouselAlign = iota
	CarouselAlignStart
	CarouselAlignEnd
)

// CarouselContain controls duplicate and out-of-bounds snaps when not looping.
type CarouselContain uint8

const (
	CarouselTrimSnaps CarouselContain = iota
	CarouselKeepSnaps
	CarouselUncontained
)

// CarouselBasis is an item's outer size, including its leading spacing. Its
// zero value fills the track's view size. Fractions use the measured native
// viewport plus the compensated leading spacing, like the source flex track.
type CarouselBasis struct{ fraction, dip float32 }

func CarouselFraction(fraction float32) CarouselBasis {
	if !carouselFinite(fraction) || fraction <= 0 {
		panic("display: CarouselFraction requires a finite positive fraction")
	}
	return CarouselBasis{fraction: fraction}
}

func CarouselFixed(dip float32) CarouselBasis {
	if !carouselFinite(dip) || dip <= 0 {
		panic("display: CarouselFixed requires a finite positive size")
	}
	return CarouselBasis{dip: dip}
}

func carouselFinite(n float32) bool { return !math.IsNaN(float64(n)) && !math.IsInf(float64(n), 0) }

type carouselGeometry struct {
	viewport, spacing, view, total float32
	starts, sizes, snaps           []float32
	loop                           bool
}

func (g carouselGeometry) equal(other carouselGeometry) bool {
	return g.viewport == other.viewport && g.spacing == other.spacing && g.loop == other.loop &&
		slices.Equal(g.sizes, other.sizes) && slices.Equal(g.snaps, other.snaps)
}

// This is a native geometry translation of the pinned Embla 8.5.2 Alignment,
// ScrollSnaps, ScrollContain and SlideLooper.canLoop behavior. The verified
// upstream functions and independently generated oracle live in testdata.
func carouselMeasure(viewport, spacing float32, items []CarouselBasis, align CarouselAlign, contain CarouselContain, loop bool) carouselGeometry {
	g := carouselGeometry{viewport: viewport, spacing: spacing, view: viewport + spacing}
	if len(items) == 0 || g.view <= 0 {
		return g
	}
	g.starts = make([]float32, len(items))
	g.sizes = make([]float32, len(items))
	raw := make([]float32, len(items))
	for i, basis := range items {
		size := g.view
		if basis.dip > 0 {
			size = basis.dip
		} else if basis.fraction > 0 {
			size *= basis.fraction
		}
		// Both CSS border-box flex items and native padded boxes have an
		// outer used size at least as large as their leading padding.
		size = max(size, spacing)
		g.starts[i], g.sizes[i] = g.total, size
		offset := (g.view - size) / 2
		if align == CarouselAlignStart {
			offset = 0
		} else if align == CarouselAlignEnd {
			offset = g.view - size
		}
		raw[i] = -g.total + offset
		g.total += size
	}
	g.loop = loop && carouselCanLoop(g, raw[0])
	if g.loop || contain == CarouselUncontained {
		g.snaps = raw
		return g
	}
	if g.total <= g.view+2 { // Embla's pixel tolerance.
		g.snaps = []float32{0}
		return g
	}
	minimum := g.view - g.total
	for i, snap := range raw {
		snap = max(minimum, min(0, snap))
		if i == 0 || carouselAbs(snap) <= 1 {
			snap = 0
		} else if i == len(raw)-1 || carouselAbs(snap-minimum) <= 1 {
			snap = minimum
		}
		raw[i] = float32(math.Round(float64(snap)*1000) / 1000)
	}
	if contain == CarouselKeepSnaps {
		g.snaps = raw
		return g
	}
	first, last := 0, len(raw)
	for first+1 < len(raw) && raw[first+1] == raw[0] {
		first++
	}
	for i, snap := range raw {
		if snap == raw[len(raw)-1] {
			last = i + 1
			break
		}
	}
	g.snaps = raw[first:last]
	return g
}

func carouselAbs(n float32) float32 { return float32(math.Abs(float64(n))) }

func carouselCanLoop(g carouselGeometry, first float32) bool {
	check := func(reverse bool, gap float32) bool {
		for j := range g.sizes {
			if gap <= 0 {
				break
			}
			i := j
			if reverse {
				i = len(g.sizes) - 1 - j
			}
			if g.view-(g.total-g.sizes[i]) > .1 {
				return false
			}
			gap -= g.sizes[i]
		}
		return true
	}
	return check(true, first) && check(false, g.view-first-1)
}

func (g carouselGeometry) position(virtual int) float32 {
	if len(g.snaps) == 0 {
		return 0
	}
	if !g.loop {
		return g.snaps[max(0, min(virtual, len(g.snaps)-1))]
	}
	index := carouselModulo(virtual, len(g.snaps))
	cycle := (virtual - index) / len(g.snaps)
	return g.snaps[index] - float32(cycle)*g.total
}

// occurrence moves a stable keyed item only after its preceding occurrence
// has completely left the viewport. It works in both directions, including a
// reversal of a transition that has not reached its selected snap yet.
func (g carouselGeometry) occurrence(index int, position float32) float32 {
	if !g.loop || g.total <= 0 {
		return 0
	}
	cycle := math.Floor(float64((-position+g.spacing-g.starts[index]-g.sizes[index])/g.total)) + 1
	return float32(cycle) * g.total
}

func (g carouselGeometry) visible(index int, position float32) bool {
	if index >= len(g.sizes) || g.viewport <= 0 {
		return false
	}
	start := g.starts[index] - g.spacing + position + g.occurrence(index, position)
	return start < g.viewport-.01 && start+g.sizes[index] > .01
}

func carouselModulo(n, count int) int { return (n%count + count) % count }
