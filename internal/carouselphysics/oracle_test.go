package carouselphysics

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

type event struct {
	Action      string
	Value       float64
	PointerDown bool
}
type bodySample struct {
	Location, Offset, Previous, Target, Duration, Velocity, Direction float64
	Settled                                                           bool
}
type bodyCase struct {
	Name              string
	Initial, Duration float64
	Events            []event
	Trace             []bodySample
	Bounds            *struct{ Min, Max, View float64 }
	Loop              *struct{ Min, Max, Content float64 }
}
type targetCase struct {
	Current, Distance, Direction, Content float64
	Index                                 *int
	Loop, Snap                            bool
	Snaps                                 []float64
	Output                                Target
}
type dragEvent struct {
	Point, Time float64
	Up          *float64
}
type releaseCase struct {
	Current, Velocity, View, Content float64
	Selected                         int
	Mouse, DragFree, SkipSnaps, Loop bool
	AxisSign                         *float64
	Snaps                            []float64
	Output                           Release
}
type schedulerSample struct {
	Updates                                   int
	Alpha, Location, Offset, Previous, Target float64
	Active                                    bool
}
type schedulerCase struct {
	Name            string
	Timestamps      []float64
	ResetAt, StopAt *int
	Trace           []schedulerSample
}
type fixture struct {
	SourceSHA256 string
	Bodies       []bodyCase
	Targets      []targetCase
	Drags        []struct {
		Events []dragEvent
		Trace  []float64
	}
	Releases   []releaseCase
	Schedulers []schedulerCase
}

func oracle(t *testing.T) fixture {
	t.Helper()
	data, err := os.ReadFile("../../pkg/display/testdata/shadcn-carousel/physics/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var f fixture
	if err = json.Unmarshal(data, &f); err != nil {
		t.Fatal(err)
	}
	if f.SourceSHA256 != "a90f9c822bd7102e4b13c985c7926f910c13fc299a368f72c9c044eb730b2507" {
		t.Fatal("unrecognized official Embla source")
	}
	return f
}
func near(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.IsNaN(got) || math.IsInf(got, 0) || math.Abs(got-want) > 1e-9 {
		t.Fatalf("%s got%.16g want%.16g", label, got, want)
	}
}
func targeter(snaps []float64, content float64, loop bool) Targeter {
	min := snaps[len(snaps)-1]
	if loop {
		min = snaps[0] - content
	}
	return Targeter{Snaps: snaps, Content: content, Loop: loop, Limit: Limit{min, snaps[0]}}
}

func TestOfficialScrollBodyBoundsAndLoop(t *testing.T) {
	for _, c := range oracle(t).Bodies {
		t.Run(c.Name, func(t *testing.T) {
			b := NewBody(c.Initial, c.Duration, .68)
			var bounds Bounds
			if c.Bounds != nil {
				bounds = Bounds{Limit: Limit{c.Bounds.Min, c.Bounds.Max}, View: c.Bounds.View}
			}
			for i, e := range c.Events {
				switch e.Action {
				case "target":
					b.State.Target = e.Value
				case "seek":
					b.Seek()
				case "interpolate":
					b.Interpolate(e.Value)
				case "duration":
					b.UseDuration(e.Value)
				case "friction":
					b.UseFriction(e.Value)
				case "baseDuration":
					b.UseBaseDuration()
				case "baseFriction":
					b.UseBaseFriction()
				case "offset":
					b.State.Offset = e.Value
				case "bounds":
					bounds.Constrain(b, e.PointerDown)
				case "activeBounds":
					bounds.Disabled = e.Value == 0
				case "wrap":
					b.Wrap(c.Loop.Content, Limit{c.Loop.Min, c.Loop.Max})
				default:
					t.Fatal("unknown operation", e.Action)
				}
				w := c.Trace[i]
				near(t, "location", b.State.Location, w.Location)
				near(t, "offset", b.State.Offset, w.Offset)
				near(t, "previous", b.State.Previous, w.Previous)
				near(t, "target", b.State.Target, w.Target)
				near(t, "duration", b.Duration(), w.Duration)
				near(t, "velocity", b.Velocity(), w.Velocity)
				near(t, "direction", b.Direction(), w.Direction)
				if b.Settled() != w.Settled {
					t.Fatalf("event%d settled=%t want%t", i, b.Settled(), w.Settled)
				}
			}
		})
	}
}
func TestOfficialScrollTargets(t *testing.T) {
	for i, c := range oracle(t).Targets {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			api := targeter(c.Snaps, c.Content, c.Loop)
			var got Target
			if c.Index == nil {
				got = api.ByDistance(c.Current, c.Distance, c.Snap)
			} else {
				got = api.ByIndex(c.Current, *c.Index, c.Direction)
			}
			if got.Index != c.Output.Index {
				t.Fatalf("index%d want%d", got.Index, c.Output.Index)
			}
			near(t, "distance", got.Distance, c.Output.Distance)
		})
	}
}
func TestOfficialDragTracker(t *testing.T) {
	for i, c := range oracle(t).Drags {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			var d DragTracker
			for j, e := range c.Events {
				var got float64
				if e.Up != nil {
					got = d.Up(*e.Up)
				} else if j == 0 {
					got = d.Down(e.Point, e.Time)
				} else {
					got = d.Move(e.Point, e.Time)
				}
				near(t, "delta/velocity", got, c.Trace[j])
			}
		})
	}
}
func TestOfficialReleaseForces(t *testing.T) {
	for i, c := range oracle(t).Releases {
		t.Run(string(rune('a'+i)), func(t *testing.T) {
			direction := float64(1)
			if c.AxisSign != nil {
				direction = *c.AxisSign
			}
			got := targeter(c.Snaps, c.Content, c.Loop).Release(c.Current, c.Selected, c.Velocity, c.View, c.Mouse, c.DragFree, c.SkipSnaps, direction)
			w := c.Output
			if got.Target.Index != w.Target.Index {
				t.Fatalf("index%d want%d", got.Target.Index, w.Target.Index)
			}
			near(t, "target distance", got.Target.Distance, w.Target.Distance)
			near(t, "raw force", got.RawForce, w.RawForce)
			near(t, "allowed force", got.AllowedForce, w.AllowedForce)
			near(t, "duration", got.Duration, w.Duration)
			near(t, "friction", got.Friction, w.Friction)
		})
	}
}
func TestOfficialFixedTimeStepScheduler(t *testing.T) {
	for _, c := range oracle(t).Schedulers {
		t.Run(c.Name, func(t *testing.T) {
			b := NewBody(0, 25, .68)
			b.State.Target = 600
			var a Scheduler
			a.Start()
			lastAlpha := float64(0)
			for i, now := range c.Timestamps {
				if c.ResetAt != nil && *c.ResetAt == i {
					a.Reset()
				}
				updates := 0
				a.Advance(now, func() { updates++; b.Seek() }, func(alpha float64) {
					lastAlpha = alpha
					b.Interpolate(alpha)
					if c.StopAt != nil && *c.StopAt == i {
						a.Stop()
					}
				})
				w := c.Trace[i]
				if updates != w.Updates || a.Active() != w.Active {
					t.Fatalf("frame%d updates%d/%d active%t/%t", i, updates, w.Updates, a.Active(), w.Active)
				}
				near(t, "alpha", lastAlpha, w.Alpha)
				near(t, "location", b.State.Location, w.Location)
				near(t, "offset", b.State.Offset, w.Offset)
				near(t, "previous", b.State.Previous, w.Previous)
				near(t, "target", b.State.Target, w.Target)
			}
		})
	}
}
