package carouselphysics

import "math"

// DragTracker stores scalar coordinates/timestamps, never native event objects.
// Coordinates are the physical main-axis values; apply axis direction to the
// returned velocity/deltas in the binding, as the source DragHandler does.
type DragTracker struct {
	startPoint, lastPoint, startTime, lastTime float64
	started                                    bool
}

func (d *DragTracker) Down(point, timeMS float64) float64 {
	d.startPoint, d.lastPoint, d.startTime, d.lastTime, d.started = point, point, timeMS, timeMS, true
	return point
}
func (d *DragTracker) Move(point, timeMS float64) float64 {
	diff := point - d.lastPoint
	expired := timeMS-d.startTime > 170
	d.lastPoint, d.lastTime = point, timeMS
	if expired {
		d.startPoint, d.startTime = point, timeMS
	}
	return diff
}
func (d *DragTracker) Up(timeMS float64) float64 {
	if !d.started {
		return 0
	}
	delta, elapsed := d.lastPoint-d.startPoint, timeMS-d.startTime
	expired := timeMS-d.lastTime > 170
	if elapsed == 0 || expired {
		return 0
	}
	force := delta / elapsed
	if math.Abs(force) > .1 {
		return force
	}
	return 0
}
