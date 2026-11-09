package carouselautoplay

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

type record struct {
	Now         float64
	Kind, Event string
	Index       int
	Jump        bool
	Selected    int
	Playing     bool
	Remaining   *float64
}
type sourceCase struct {
	Name         string
	Options      json.RawMessage
	Delay        []float64
	Count        int
	Loop, Hidden bool
	WatchDrag    *bool
	Ops          [][]json.RawMessage
	Records      []record
}

func TestPinnedOfficialAutoplayStateMachine(t *testing.T) {
	data, err := os.ReadFile("../../pkg/display/testdata/shadcn-carousel/autoplay/oracle.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		SourceSHA256 string
		Cases        []sourceCase
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 13 {
		t.Fatal("missing source scenarios")
	}
	if fixture.SourceSHA256 != "84911d2f9b684038360e7be1019923462ffb8a34659b868f6fb046eb13edd42a" {
		t.Fatal("wrong official source")
	}
	for _, spec := range fixture.Cases {
		t.Run(spec.Name, func(t *testing.T) {
			count := spec.Count
			if count == 0 {
				count = 3
			}
			p := Defaults()
			if len(spec.Options) > 0 {
				if err := json.Unmarshal(spec.Options, &p); err != nil {
					t.Fatal(err)
				}
			}
			if spec.WatchDrag != nil {
				p.WatchDrag = *spec.WatchDrag
			}
			p.Delays = append([]float64(nil), spec.Delay...)
			if p.Delays == nil {
				for range count {
					p.Delays = append(p.Delays, 4000)
				}
			}
			m := &Core{}
			now := float64(1000)
			var got []record
			capture := func(kind, event string, index int, jump bool) {
				got = append(got, record{Now: now, Kind: kind, Event: event, Index: index, Jump: jump, Selected: m.Selected(), Playing: m.IsPlaying(), Remaining: m.TimeUntilNext(now)})
			}
			emit := func(e Effect) { capture(e.Kind, e.Event, e.Index, e.Jump) }
			m.Init(p, count, spec.Loop, spec.Hidden, now, emit)
			capture("state", "", 0, false)
			for _, step := range spec.Ops {
				var op string
				json.Unmarshal(step[0], &op)
				json.Unmarshal(step[1], &now)
				switch op {
				case "tick":
					m.Advance(now, emit)
				case "select":
					var i int
					json.Unmarshal(step[2], &i)
					m.SetSelected(i)
				case "play":
					var override *bool
					if len(step) > 2 {
						json.Unmarshal(step[2], &override)
					}
					m.Play(now, override, emit)
				case "stop":
					m.Stop(emit)
				case "reset":
					m.Reset(now, emit)
				case "destroy":
					m.Destroy(emit)
				case "visibility":
					var hidden bool
					json.Unmarshal(step[2], &hidden)
					m.Visibility(hidden, now, emit)
				case "pointerDown":
					m.PointerDown(emit)
				case "pointerUp":
					m.PointerUp(now, emit)
				case "mouseEnter":
					m.MouseEnter(emit)
				case "mouseLeave":
					m.MouseLeave(now, emit)
				case "focusStart":
					m.FocusStart(emit)
				case "focusOut":
					m.FocusOut(now, emit)
				case "observe":
				default:
					t.Fatal(op)
				}
				capture("state", "", 0, false)
			}
			if !reflect.DeepEqual(got, spec.Records) {
				for i := range min(len(got), len(spec.Records)) {
					if !reflect.DeepEqual(got[i], spec.Records[i]) {
						a, _ := json.Marshal(got[i])
						b, _ := json.Marshal(spec.Records[i])
						t.Fatalf("point%d got%s want%s (counts %d/%d)", i, a, b, len(got), len(spec.Records))
					}
				}
				t.Fatalf("trace count got%d want%d", len(got), len(spec.Records))
			}
		})
	}
}
