package display

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestAlertUsesVegaInsetsAndDescriptionSpacing(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		Alert(c, AlertProps{Title: "Changes saved", Description: "Your workspace is up to date."}).Label("Alert frame")
	})
	frame, _ := r.Find("Alert frame")
	title, _ := r.Find("Changes saved")
	description, _ := r.Find("Your workspace is up to date.")
	if title.X-frame.X != 17 || title.Y-frame.Y != 13 {
		t.Fatalf("alert content inset does not match px4/py3: frame=%+v title=%+v", frame, title)
	}
	if description.Y-(title.Y+title.H) != 2 {
		t.Fatalf("title/description should be separated by 2 DIPs: title=%+v description=%+v", title, description)
	}
}

func TestEmptyUsesVegaPaddingAndOmitsEmptyText(t *testing.T) {
	r := displayTester(func(c *ui.Context) {
		Empty(c, EmptyProps{Title: "No projects"}).Label("Empty frame")
	})
	frame, _ := r.Find("Empty frame")
	title, _ := r.Find("No projects")
	if title.Y-frame.Y != 48 {
		t.Fatalf("empty header did not use p12 spacing: frame=%+v title=%+v", frame, title)
	}
	if frame.H != title.H+96 {
		t.Fatalf("empty description created an empty layout gap: frame=%+v title=%+v", frame, title)
	}
}

var surfaceGlyph = ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16"><path fill="currentColor" d="M0 0h16v16H0z"/></svg>`))

func surfaceAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	n, ok := r.AccessByRole(role, label)
	if !ok {
		t.Fatalf("missing role=%v label=%q: %+v", role, label, r.AccessAll())
	}
	return n
}

func TestAlertIconActionWrapsAndUsesLogicalEdges(t *testing.T) {
	clicks := 0
	label := "A saved update with a long title that must wrap beside the action"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Padding(20).Children(func() {
			AlertWith(c, AlertProps{Label: "Saved update", Title: label, Description: "You can undo this change."}, AlertSlots{
				Icon: func(child *ui.Context) { ui.Icon(child, surfaceGlyph).Size(16, 16).Label("Alert glyph") },
				Action: func(child *ui.Context) {
					if control.Button(child, "Undo", control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
						clicks++
					}
				},
			})
		})
	}, 300, 600)
	frame := surfaceAccess(t, r, ui.RoleAlert, "Saved update").Bounds
	undo := surfaceAccess(t, r, ui.RoleButton, "Undo").Bounds
	icon, _ := r.Find("Alert glyph")
	title, _ := r.Find(label)
	if icon.W != 16 || icon.H != 16 || icon.X-frame.X != 17 || icon.Y-frame.Y != 15 || title.X-icon.X != 26 {
		t.Fatalf("icon/title insets frame=%+v icon=%+v title=%+v", frame, icon, title)
	}
	if title.H <= 21 || title.X+title.W > undo.X-15 || undo.Y-frame.Y != 11 || frame.X+frame.W-(undo.X+undo.W) != 13 {
		t.Fatalf("action overlapped wrapped text: frame=%+v title=%+v action=%+v", frame, title, undo)
	}
	if r.Focused("Undo") {
		t.Fatal("mounting an alert stole keyboard focus")
	}
	displayClick(t, r, "Undo")
	r.Key(0, ui.KeySpace)
	if clicks != 2 {
		t.Fatalf("native action pointer+keyboard activations=%d", clicks)
	}
	r.SetDirection(ui.RightToLeft)
	frame = surfaceAccess(t, r, ui.RoleAlert, "Saved update").Bounds
	undo = surfaceAccess(t, r, ui.RoleButton, "Undo").Bounds
	icon, _ = r.Find("Alert glyph")
	title, _ = r.Find(label)
	if math.Abs(float64(undo.X-frame.X-13)) > .001 || math.Abs(float64(frame.X+frame.W-(icon.X+icon.W)-17)) > .001 || title.X < undo.X+undo.W+15 || title.X+title.W > icon.X-9 {
		t.Fatalf("RTL action inset=%g icon inset=%g text/action gap=%g icon/text gap=%g", undo.X-frame.X, frame.X+frame.W-(icon.X+icon.W), title.X-(undo.X+undo.W), icon.X-(title.X+title.W))
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("static alert requested idle animation frames")
	}
}

func surfaceColorCount(im *image.RGBA, bounds ui.Rect, matches func(color.RGBA) bool) int {
	count := 0
	for y := max(0, int(bounds.Y)); y < min(im.Bounds().Dy(), int(bounds.Y+bounds.H)); y++ {
		for x := max(0, int(bounds.X)); x < min(im.Bounds().Dx(), int(bounds.X+bounds.W)); x++ {
			if matches(im.RGBAAt(x, y)) {
				count++
			}
		}
	}
	return count
}

func TestAlertSemanticSurfaceAndDestructiveDescriptionPixels(t *testing.T) {
	for _, destructive := range []bool{false, true} {
		t.Run(map[bool]string{false: "default", true: "destructive"}[destructive], func(t *testing.T) {
			tokens := theme.Light()
			tokens.Card, tokens.CardForeground, tokens.MutedForeground, tokens.Destructive = ui.RGB(250, 250, 250), ui.RGB(10, 150, 30), ui.RGB(30, 60, 180), ui.RGB(210, 40, 70)
			r := ui.NewTester(func(c *ui.Context) {
				theme.Scope(c, tokens).FillWidth().Children(func() {
					Alert(c, AlertProps{Label: "Status notice", Title: "MMMMM", Description: "NNNNN", Icon: surfaceGlyph, Destructive: destructive})
				})
			}, 360, 160)
			n := surfaceAccess(t, r, ui.RoleAlert, "Status notice")
			if n.Focusable {
				t.Fatal("assertive alert became an unsolicited keyboard stop")
			}
			if len(r.AccessAll()) == 0 {
				t.Fatal("native alert contents were removed from accessibility")
			}
			title, _ := r.Find("MMMMM")
			desc, _ := r.Find("NNNNN")
			isRed := func(p color.RGBA) bool { return int(p.R)-int(p.G) > 40 && int(p.R)-int(p.B) > 40 }
			isGreen := func(p color.RGBA) bool { return int(p.G)-int(p.R) > 40 && int(p.G)-int(p.B) > 40 }
			isBlue := func(p color.RGBA) bool { return int(p.B)-int(p.R) > 40 && int(p.B)-int(p.G) > 40 }
			titleMatch, descMatch := isGreen, isBlue
			if destructive {
				titleMatch, descMatch = isRed, isRed
			}
			if surfaceColorCount(r.Image(), title, titleMatch) < 10 || surfaceColorCount(r.Image(), desc, descMatch) < 10 {
				t.Fatal("title/description ignored scoped semantic foregrounds")
			}
			iconColor := r.Image().RGBAAt(int(n.Bounds.X+21), int(n.Bounds.Y+19))
			want := tokens.CardForeground
			if destructive {
				want = tokens.Destructive
			}
			if iconColor.R != want.R || iconColor.G != want.G || iconColor.B != want.B {
				t.Fatalf("icon did not inherit current foreground: %v want=%v", iconColor, want)
			}
		})
	}
}

func TestAlertCustomActionsRetainEditorAndNativeLinks(t *testing.T) {
	showIcon, inits := false, 0
	r := displayTester(func(c *ui.Context) {
		p := AlertProps{Label: "Session expired", Title: "Session expired"}
		if showIcon {
			p.Icon = surfaceGlyph
		}
		AlertWith(c, p, AlertSlots{
			Description: func(child *ui.Context) { ui.Link(child, "Get help", "https://example.com/help") },
			Action: func(child *ui.Context) {
				owner := ui.Column(child)
				draft := ui.Local(owner, "name", func() string { inits++; return "Draft" })
				owner.Children(func() { control.Input(child, draft, control.InputProps{Label: "Recovery name"}).Width(120) })
			},
		})
	})
	displayClick(t, r, "Get help")
	if urls := r.OpenedURLs(); len(urls) != 1 || urls[0] != "https://example.com/help" {
		t.Fatalf("description link lost native URL behavior: %v", urls)
	}
	displayClick(t, r, "Recovery name")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited")
	showIcon = true
	r.Frame()
	input := surfaceAccess(t, r, ui.RoleTextField, "Recovery name")
	if input.Value != "Edited" || !r.Focused("Recovery name") || inits != 1 {
		t.Fatalf("icon change recreated action editor: value=%q focus=%v initializations=%d", input.Value, input.Focused, inits)
	}
}

func TestAlertWiderNativeActionsReserveRealWidth(t *testing.T) {
	review, clicks := false, 0
	titleText := "Your updates have been saved in the current workspace"
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Padding(20).Children(func() {
			AlertWith(c, AlertProps{Title: titleText, Icon: surfaceGlyph}, AlertSlots{Action: func(child *ui.Context) {
				control.Button(child, "Undo", control.ButtonProps{Size: control.ExtraSmall})
				if review && control.Button(child, "Review", control.ButtonProps{Size: control.ExtraSmall}).Clicked() {
					clicks++
				}
			}})
		})
	}, 400, 400)
	before, _ := r.Find(titleText)
	review = true
	r.Frame()
	after, _ := r.Find(titleText)
	undo := surfaceAccess(t, r, ui.RoleButton, "Undo").Bounds
	if after.W >= before.W-30 || after.X+after.W > undo.X-15 {
		t.Fatalf("a wider action did not reserve real content space: before=%+v after=%+v action=%+v", before, after, undo)
	}
	displayClick(t, r, "Review")
	if clicks != 1 {
		t.Fatal("second real action could not receive pointer input")
	}
}

func TestAlertOversizedActionStacksInsideBoundedNativeScroll(t *testing.T) {
	clicks := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(180).Padding(8).Children(func() {
			AlertWith(c, AlertProps{Title: "Saved", Label: "Notice"}, AlertSlots{Action: func(child *ui.Context) {
				if ui.Button(child, "Review all changes").Width(190).Clicked() {
					clicks++
				}
			}})
		})
	}, 240, 300)
	frame := surfaceAccess(t, r, ui.RoleAlert, "Notice").Bounds
	action := surfaceAccess(t, r, ui.RoleButton, "Review all changes").Bounds
	viewport := surfaceAccess(t, r, ui.RoleScroll, "Alert actions").Bounds
	title, _ := r.Find("Saved")
	if viewport.X < frame.X || viewport.X+viewport.W > frame.X+frame.W || action.Y < title.Y+title.H+11 || title.W < 60 {
		t.Fatalf("wide action viewportEnd=%g frameEnd=%g gap=%g titleWidth=%g", viewport.X+viewport.W, frame.X+frame.W, action.Y-title.Y-title.H, title.W)
	}
	// Native accessibility reports the entire scroll child's geometry. The
	// viewport must still clip both its physical pixels and pointer hit box.
	x, y := frame.X+frame.W+10, action.Y+action.H/2
	if got, want := r.Image().RGBAAt(int(x), int(y)), r.Image().RGBAAt(235, int(y)); got != want {
		t.Fatal("wide action painted outside the alert's bounded viewport")
	}
	r.ClickAt(x, y)
	if clicks != 0 {
		t.Fatal("clipped action accepted an out-of-alert pointer hit")
	}
	r.ClickAt(viewport.X+viewport.W/2, y)
	r.Key(0, ui.KeyEnter)
	if clicks != 2 {
		t.Fatal("bounded wide action lost native pointer/keyboard activation")
	}
	r.SetDirection(ui.RightToLeft)
	frame = surfaceAccess(t, r, ui.RoleAlert, "Notice").Bounds
	action = surfaceAccess(t, r, ui.RoleButton, "Review all changes").Bounds
	viewport = surfaceAccess(t, r, ui.RoleScroll, "Alert actions").Bounds
	if viewport.X < frame.X || viewport.X+viewport.W > frame.X+frame.W {
		t.Fatal("RTL action escaped its horizontal viewport")
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("responsive action measurement did not converge")
	}
}

func TestAlertResponsiveActionKeepsLocalEditorStateAndFocus(t *testing.T) {
	inits := 0
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Padding(8).Children(func() {
			AlertWith(c, AlertProps{Title: "Saved", Label: "Editable notice"}, AlertSlots{Action: func(child *ui.Context) {
				owner := ui.Row(child)
				draft := ui.Local(owner, "draft", func() string { inits++; return "Draft" })
				owner.Children(func() { control.Input(child, draft, control.InputProps{Label: "Alert draft"}).Width(120) })
			}})
		})
	}, 500, 300)
	displayClick(t, r, "Alert draft")
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("Edited")
	r.SetSize(180, 300)
	r.Frame()
	input := surfaceAccess(t, r, ui.RoleTextField, "Alert draft")
	title, _ := r.Find("Saved")
	if input.Value != "Edited" || !input.Focused || input.Bounds.Y < title.Y+title.H+11 || inits != 1 {
		t.Fatalf("stacking rebuilt/defocused the native action editor: value=%q focus=%v inits=%d", input.Value, input.Focused, inits)
	}
	r.SetSize(500, 300)
	r.Frame()
	input = surfaceAccess(t, r, ui.RoleTextField, "Alert draft")
	if input.Value != "Edited" || !input.Focused || inits != 1 {
		t.Fatal("returning inline rebuilt the action editor")
	}
}

func TestEmptyBoundedHeaderMediaAndNativeContent(t *testing.T) {
	clicks, query := 0, ""
	r := ui.NewTester(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).FillWidth().Padding(20).Children(func() {
			EmptyWith(c, EmptyProps{Label: "No projects", Bordered: true}, EmptySlots{
				Header: func(child *ui.Context) {
					EmptyMedia(child, EmptyMediaProps{Variant: EmptyMediaIcon, Icon: surfaceGlyph, Label: "Folder illustration"})
					EmptyTitle(child, "No projects yet")
					EmptyDescription(child, "There are no projects in this workspace. Create one or search for an existing project to continue.")
				},
				Content: func(child *ui.Context) {
					ui.Box(child).FillWidth().Height(1).Label("Content bounds")
					control.Input(child, &query, control.InputProps{Label: "Search projects"}).WidthPercent(75)
					if control.Button(child, "Create project", control.ButtonProps{}).Clicked() {
						clicks++
					}
				},
			})
		})
	}, 720, 650)
	frame := surfaceAccess(t, r, ui.RoleGroup, "No projects").Bounds
	media := surfaceAccess(t, r, ui.RoleImage, "Folder illustration").Bounds
	content, _ := r.Find("Content bounds")
	input := surfaceAccess(t, r, ui.RoleTextField, "Search projects").Bounds
	if media.W != 40 || media.H != 40 || content.W != 384 || input.W != 288 || media.X+20 != frame.X+frame.W/2 || content.X+192 != frame.X+frame.W/2 {
		t.Fatalf("bounded empty layout frame=%+v media=%+v content=%+v input=%+v", frame, media, content, input)
	}
	title, _ := r.Find("No projects yet")
	if title.Y-(media.Y+media.H) != 16 {
		t.Fatalf("media bottom margin + header gap lost: media=%+v title=%+v", media, title)
	}
	displayClick(t, r, "Search projects")
	r.Type("My project")
	displayClick(t, r, "Create project")
	if query != "My project" || clicks != 1 {
		t.Fatal("empty content lost native editor/action input")
	}
	r.SetSize(240, 800)
	r.Frame()
	frame = surfaceAccess(t, r, ui.RoleGroup, "No projects").Bounds
	content, _ = r.Find("Content bounds")
	input = surfaceAccess(t, r, ui.RoleTextField, "Search projects").Bounds
	if content.W > frame.W-98 || math.Abs(float64(input.W-content.W*.75)) > 1 || content.X < frame.X+49 {
		t.Fatalf("empty header/content overflow constrained width: frame=%+v content=%+v input=%+v", frame, content, input)
	}
	r.Advance(time.Second)
	if r.FramePending() {
		t.Fatal("static empty state requested idle frames")
	}
}

func TestEmptyMediaKeepsMeaningfulChildrenUnlessDecorative(t *testing.T) {
	decorative := false
	r := displayTester(func(c *ui.Context) {
		EmptyMedia(c, EmptyMediaProps{Decorative: decorative, Content: func(child *ui.Context) {
			AvatarGroup(child, AvatarGroupProps{Items: []AvatarProps{{Key: "alice", Label: "Alice", Fallback: "A"}, {Key: "bob", Label: "Bob", Fallback: "B"}}})
		}})
	})
	if _, ok := r.AccessByRole(ui.RoleImage, "Alice"); !ok {
		t.Fatal("default custom media hid meaningful avatar semantics")
	}
	decorative = true
	r.Frame()
	if r.Focused("Invite teammate") {
		t.Fatal("media turned decorative but retained hidden control focus")
	}
	if _, ok := r.AccessByRole(ui.RoleImage, "Alice"); ok {
		t.Fatal("decorative custom media remained in accessibility tree")
	}
}

func TestEmptyNamedCustomMediaPreservesInteractiveAccessibility(t *testing.T) {
	decorative, calls := false, 0
	r := displayTester(func(c *ui.Context) {
		ui.Button(c, "Before media")
		EmptyMedia(c, EmptyMediaProps{Label: "Team artwork", Decorative: decorative, Content: func(child *ui.Context) {
			if ui.Button(child, "Invite teammate").Clicked() {
				calls++
			}
		}})
		ui.Button(c, "After media")
	})
	r.Key(0, ui.KeyTab)
	r.Key(0, ui.KeyTab)
	if !r.Focused("Invite teammate") {
		t.Fatal("named custom media removed its native button from keyboard order")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Invite teammate"); !ok {
		t.Fatal("named custom media hid a keyboard-focusable native control")
	}
	r.Key(0, ui.KeySpace)
	if calls != 1 {
		t.Fatal("custom media control did not activate")
	}
	decorative = true
	r.Frame()
	displayClick(t, r, "Before media")
	r.Key(0, ui.KeyTab)
	if !r.Focused("After media") {
		t.Fatal("decorative media left a hidden keyboard stop")
	}
	if _, ok := r.AccessByRole(ui.RoleButton, "Invite teammate"); ok {
		t.Fatal("decorative media still exposed its native control")
	}
}

func TestEmptyThemedDashedBorderAndFinitePadding(t *testing.T) {
	for _, dark := range []bool{false, true} {
		t.Run(map[bool]string{false: "light", true: "dark"}[dark], func(t *testing.T) {
			tokens := theme.Light()
			if dark {
				tokens = theme.Dark()
			}
			padding := float32(math.NaN())
			r := ui.NewTester(func(c *ui.Context) {
				theme.Scope(c, tokens).FillWidth().Background(tokens.Background).Children(func() {
					Empty(c, EmptyProps{Label: "Themed empty", Title: "No content", Bordered: true, Muted: true, Padding: &padding, ContentWidth: float32(math.Inf(1))})
				})
			}, 600, 300)
			frame := surfaceAccess(t, r, ui.RoleGroup, "Themed empty").Bounds
			title, _ := r.Find("No content")
			if title.Y-frame.Y != 49 {
				t.Fatal("nonfinite padding was not replaced with the finite default")
			}
			p := r.Image().RGBAAt(int(frame.X+20), int(frame.Y+20))
			if p.R != tokens.Muted.R || p.G != tokens.Muted.G || p.B != tokens.Muted.B {
				t.Fatalf("muted empty background ignored semantic color: got=%v want=%v", p, tokens.Muted)
			}
			transitions := 0
			last := r.Image().RGBAAt(int(frame.X+20), int(frame.Y))
			for x := int(frame.X + 21); x < int(frame.X+frame.W-20); x++ {
				current := r.Image().RGBAAt(x, int(frame.Y))
				if current != last {
					transitions++
				}
				last = current
			}
			if transitions < 10 {
				t.Fatal("border did not actually paint alternating dash segments")
			}
			padding = 0
			r.Frame()
			frame = surfaceAccess(t, r, ui.RoleGroup, "Themed empty").Bounds
			title, _ = r.Find("No content")
			if title.Y-frame.Y != 1 {
				t.Fatal("explicit zero padding was ignored")
			}
		})
	}
}

func TestStatusSurfaceEmptyCallbacksDoNotLeakSiblingIdentity(t *testing.T) {
	for _, slot := range []string{"alert-icon", "alert-title", "alert-description", "alert-content", "alert-action", "empty-header", "empty-media", "empty-title", "empty-description", "empty-content"} {
		t.Run(slot, func(t *testing.T) {
			show, text := false, "Outside draft"
			r := displayTester(func(c *ui.Context) {
				build := func(child *ui.Context) {
					if show {
						ui.Text(child, "Optional content")
					}
				}
				a, e := AlertSlots{}, EmptySlots{}
				switch slot {
				case "alert-icon":
					a.Icon = build
				case "alert-title":
					a.Title = build
				case "alert-description":
					a.Description = build
				case "alert-content":
					a.Content = build
				case "alert-action":
					a.Action = build
				case "empty-header":
					e.Header = build
				case "empty-media":
					e.Media = build
				case "empty-title":
					e.Title = build
				case "empty-description":
					e.Description = build
				case "empty-content":
					e.Content = build
				}
				AlertWith(c, AlertProps{}, a)
				zero := float32(0)
				EmptyWith(c, EmptyProps{Padding: &zero}, e)
				control.Input(c, &text, control.InputProps{Label: "Following editor"})
			})
			displayClick(t, r, "Following editor")
			show = true
			r.Frame()
			if !r.Focused("Following editor") {
				t.Fatal("empty callback leaked identity into the following editor")
			}
		})
	}
}
