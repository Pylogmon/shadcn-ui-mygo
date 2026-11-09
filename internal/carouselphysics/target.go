package carouselphysics

import "math"

type Limit struct{ Min, Max float64 }

func (l Limit) Reached(n float64) bool      { return n < l.Min || n > l.Max }
func (l Limit) Constrain(n float64) float64 { return clamp(n, l.Min, l.Max) }
func (l Limit) RemoveOffset(n float64) float64 {
	length := math.Abs(l.Min - l.Max)
	if length == 0 {
		return n
	}
	return n - length*math.Ceil((n-l.Max)/length)
}

type Target struct {
	Index    int
	Distance float64
}
type Targeter struct {
	Snaps   []float64
	Content float64
	Limit   Limit
	Loop    bool
}

func minDistance(values []float64) float64 {
	best := values[0]
	for _, v := range values[1:] {
		if math.Abs(v) < math.Abs(best) {
			best = v
		}
	}
	return best
}
func (t Targeter) Shortcut(distance, direction float64) float64 {
	if !t.Loop {
		return distance
	}
	targets := []float64{distance, distance + t.Content, distance - t.Content}
	if direction == 0 {
		return minDistance(targets)
	}
	matching := []float64{}
	for _, v := range targets {
		if sign(v) == direction {
			matching = append(matching, v)
		}
	}
	if len(matching) > 0 {
		return minDistance(matching)
	}
	return targets[2] - t.Content
}
func (t Targeter) ByIndex(current float64, index int, direction float64) Target {
	return Target{index, t.Shortcut(t.Snaps[index]-current, direction)}
}
func (t Targeter) ByDistance(current, distance float64, snap bool) Target {
	goal := current + distance
	normalized := t.Limit.Constrain(goal)
	if t.Loop {
		normalized = t.Limit.RemoveOffset(goal)
	}
	index := 0
	best := math.Inf(1)
	for i, offset := range t.Snaps {
		diff := math.Abs(t.Shortcut(offset-normalized, 0))
		if diff < best {
			index, best = i, diff
		}
	}
	if !snap || !t.Loop && t.Limit.Reached(goal) {
		return Target{index, distance}
	}
	return Target{index, distance + t.Shortcut(t.Snaps[index]-normalized, 0)}
}
func (t Targeter) counter(n int) int {
	if !t.Loop {
		return max(0, min(n, len(t.Snaps)-1))
	}
	return int(math.Abs(math.Mod(float64(len(t.Snaps)+n), float64(len(t.Snaps)))))
}

type Release struct {
	Target                                     Target
	RawForce, AllowedForce, Duration, Friction float64
}

// Release applies source DragHandler's snap/free force boosts and thresholds.
// DragThreshold is deliberately absent: it suppresses clicks, not navigation.
func (t Targeter) Release(current float64, selected int, velocity, view float64, mouse, dragFree, skipSnaps bool, axisSign float64) Release {
	boost := float64(400)
	speed := float64(25)
	if mouse {
		boost = 300
	}
	if dragFree {
		boost = 600
		speed = 43
		if mouse {
			boost = 500
		}
	}
	raw := velocity * boost
	force := raw * axisSign
	targetChanged := t.ByDistance(current, 0, false).Index != selected
	allowed := t.ByDistance(current, force, !dragFree).Distance
	threshold := clamp(view*.2, 50, 225)
	if !dragFree && math.Abs(force) >= threshold {
		if skipSnaps && targetChanged {
			allowed *= .5
		} else {
			next := t.counter(selected - int(sign(force)))
			allowed = t.ByIndex(current, next, 0).Distance
		}
	}
	factor := float64(0)
	if raw != 0 && allowed != 0 && math.Abs(raw) > math.Abs(allowed) {
		factor = math.Abs((math.Abs(raw) - math.Abs(allowed)) / raw)
	}
	return Release{Target: t.ByDistance(current, allowed, !dragFree), RawForce: raw, AllowedForce: allowed, Duration: speed - 10*factor, Friction: .68 + factor/50}
}
