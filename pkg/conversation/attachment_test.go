package conversation

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestAttachmentDefaultUsesVegaMediaGeometry(t *testing.T) {
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Attachment(c, AttachmentProps{Name: "report.pdf", MediaType: "PDF", SizeBytes: 2048}).Width(320)
	}, 400, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	root, ok := r.AccessByRole(ui.RoleGroup, "report.pdf")
	if !ok {
		t.Fatal("attachment group missing")
	}
	if math.Abs(float64(root.Bounds.H-58)) > .1 {
		t.Fatalf("default source media40 + padding16 + border2 must be58: %v", root.Bounds)
	}
}

func TestAttachmentVegaSizeOrientationStateGeometry(t *testing.T) {
	for _, size := range []AttachmentSize{AttachmentDefault, AttachmentSmall, AttachmentExtraSmall} {
		for _, orientation := range []ui.Orientation{ui.Horizontal, ui.Vertical} {
			for _, state := range []string{AttachmentIdle, AttachmentUploading, AttachmentProcessing, AttachmentError, AttachmentDone} {
				t.Run(fmtAttachmentCase(size, orientation, state), func(t *testing.T) {
					r := ui.NewTesterWith(func(c *ui.Context) {
						theme.Apply(c)
						parts := AttachmentWith(c, AttachmentProps{Name: "document.pdf", Size: size, Orientation: orientation, State: state}, func(a *AttachmentParts) {
							a.Media(AttachmentIconMedia, nil).Label("Measured media")
							a.Content(func(*ui.Context) { a.Title("document.pdf"); a.Description("PDF · 2 KB") })
							a.Actions(func(*ui.Context) { a.Action("Remove document", attachmentDefaultIcon("x")) }).Label("Measured actions")
						})
						if orientation == ui.Horizontal {
							parts.Root.Width(320)
						}
					}, 400, 260, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
					root, ok := r.Find("document.pdf")
					if !ok {
						t.Fatal("attachment missing")
					}
					media, ok := r.Find("Measured media")
					if !ok {
						t.Fatal("media slot missing")
					}
					if orientation == ui.Horizontal {
						wantMedia := []float32{40, 32, 28}[size]
						wantHeight := []float32{58, 46, 42}[size]
						if math.Abs(float64(media.W-wantMedia)) > .1 || math.Abs(float64(media.H-wantMedia)) > .1 || math.Abs(float64(root.H-wantHeight)) > .1 {
							t.Fatalf("source horizontal metrics: root=%v media=%v wantMedia=%v wantHeight=%v", root, media, wantMedia, wantHeight)
						}
					} else {
						wantMedia := []float32{102, 106, 110}[size]
						if root.W != 120 || math.Abs(float64(media.W-wantMedia)) > .1 || math.Abs(float64(media.H-wantMedia)) > .1 {
							t.Fatalf("source vertical square or width: root=%v media=%v wantMedia=%v", root, media, wantMedia)
						}
						actions, _ := r.Find("Measured actions")
						if math.Abs(float64(actions.Y-root.Y-13)) > .1 || math.Abs(float64(root.X+root.W-actions.X-actions.W-13)) > .1 {
							t.Fatalf("vertical actions must float12 inside the border at physical top/right: root=%v actions=%v", root, actions)
						}
					}
					if r.FramePending() {
						t.Fatal("reduced-motion attachment did not settle")
					}
				})
			}
		}
	}
}

func fmtAttachmentCase(size AttachmentSize, orientation ui.Orientation, state string) string {
	return []string{"default", "small", "extra small"}[size] + "/" + map[ui.Orientation]string{ui.Horizontal: "horizontal", ui.Vertical: "vertical"}[orientation] + "/" + state
}

func attachmentTestImage(t *testing.T) ui.ImageSource {
	t.Helper()
	picture := image.NewRGBA(image.Rect(0, 0, 64, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 64; x++ {
			pixel := color.RGBA{R: 220, A: 255}
			if x >= 16 && x < 48 {
				pixel = color.RGBA{G: 200, B: 40, A: 255}
			}
			picture.SetRGBA(x, y, pixel)
		}
	}
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, picture); err != nil {
		t.Fatal(err)
	}
	source, err := ui.DecodeImage(encoded.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return source
}

func TestAttachmentImageUsesCoverRoundedClipAndStateOpacity(t *testing.T) {
	source := attachmentTestImage(t)
	for _, state := range []string{AttachmentIdle, AttachmentUploading, AttachmentProcessing, AttachmentError, AttachmentDone} {
		t.Run(state, func(t *testing.T) {
			tokens := theme.Light()
			tokens.Card, tokens.Muted = ui.RGB(255, 255, 255), ui.RGB(180, 180, 180)
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Background(ui.RGB(255, 255, 255)).Children(func() {
					Attachment(c, AttachmentProps{Name: "workspace.png", Image: source, ImageAlt: "Workspace", State: state}).Width(320)
				})
			}, 400, 150, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			picture, ok := r.AccessByRole(ui.RoleImage, "Workspace")
			if !ok || picture.Bounds.W != 40 || picture.Bounds.H != 40 {
				t.Fatalf("native square image/alt missing: ok=%v picture=%v", ok, picture)
			}
			pixel := r.Image().RGBAAt(int(picture.Bounds.X+4), int(picture.Bounds.Y+20))
			want := color.RGBA{G: 200, B: 40, A: 255}
			if state != AttachmentIdle && state != AttachmentDone {
				want = color.RGBA{R: 102, G: 222, B: 126, A: 255}
			}
			if colorDistance(pixel, want) > 3 {
				t.Fatalf("image did not cover crop or apply source opacity: pixel=%v want=%v", pixel, want)
			}
			corner := r.Image().RGBAAt(int(picture.Bounds.X), int(picture.Bounds.Y))
			if colorDistance(corner, color.RGBA{255, 255, 255, 255}) > 3 {
				t.Fatalf("media rounded clip failed: %v", corner)
			}
		})
	}
}

func colorDistance(a, b color.RGBA) int {
	return max(absAttachment(int(a.R)-int(b.R)), absAttachment(int(a.G)-int(b.G)), absAttachment(int(a.B)-int(b.B)), absAttachment(int(a.A)-int(b.A)))
}

func absAttachment(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func TestAttachmentTriggerKeepsChildActionsAndIMEIndependent(t *testing.T) {
	opened, removed := 0, 0
	value := ""
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		AttachmentWith(c, AttachmentProps{ID: "editable", Name: "Editable file"}, func(a *AttachmentParts) {
			a.Media(AttachmentIconMedia, nil)
			a.Content(func(c *ui.Context) { ui.TextInput(c, &value).Label("Rename file").Width(140) })
			a.Actions(func(*ui.Context) {
				if a.Action("Remove file", attachmentDefaultIcon("x")).Clicked() {
					removed++
				}
			})
			if a.Trigger("Preview file").Clicked() {
				opened++
			}
		}).Root.Width(320)
	}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if err := r.Click("Rename file"); err != nil {
		t.Fatal(err)
	}
	r.Compose("文", 1)
	r.Compose("文档", 2)
	r.Type("文档")
	if value != "文档" || opened != 0 || removed != 0 || !r.Focused("Rename file") {
		t.Fatalf("whole-card trigger stole native IME: value=%q open=%d remove=%d", value, opened, removed)
	}
	r.Key(0, ui.KeyEnter)
	if opened != 0 {
		t.Fatal("input's Enter activated ancestor trigger")
	}
	if err := r.Click("Remove file"); err != nil {
		t.Fatal(err)
	}
	if removed != 1 || opened != 0 {
		t.Fatal("child action activated whole-card trigger")
	}
	root, _ := r.Find("Preview file")
	r.ClickAt(root.X+20, root.Y+20)
	r.Key(0, ui.KeyEnter)
	r.Key(0, ui.KeySpace)
	if opened != 3 || removed != 1 {
		t.Fatalf("native trigger pointer/Enter/Space failed: open=%d remove=%d", opened, removed)
	}
	r.Press(root.X+20, root.Y+20)
	r.CancelPress()
	if opened != 3 {
		t.Fatal("canceled native trigger activated")
	}
}

func TestAttachmentProgressCancelRetryAndDisabledAreApplicationOwned(t *testing.T) {
	state, progress, disabled := AttachmentUploading, .64, false
	opened, canceled, retried, removed := 0, 0, 0, 0
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Attachment(c, AttachmentProps{Name: "upload.zip", State: state, Progress: progress, Disabled: disabled,
			OnOpen: func() { opened++ }, OnCancel: func() { canceled++ }, OnRetry: func() { retried++ }, OnRemove: func() { removed++ },
		}).Width(380)
	}, 440, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.HasText("Uploading · 64%") {
		t.Fatal("caller progress was not shown")
	}
	r.Advance(time.Hour)
	if !r.HasText("Uploading · 64%") || state != AttachmentUploading || progress != .64 {
		t.Fatal("attachment invented upload transport or advanced caller state")
	}
	if err := r.Click("Open"); err != nil {
		t.Fatal(err)
	}
	if opened != 0 {
		t.Fatal("compatible Open action became enabled while uploading")
	}
	if err := r.Click("Cancel upload.zip"); err != nil {
		t.Fatal(err)
	}
	if canceled != 1 || state != AttachmentUploading {
		t.Fatal("cancel did not remain an application callback")
	}
	state = AttachmentError
	r.Frame()
	if err := r.Click("Retry upload.zip"); err != nil {
		t.Fatal(err)
	}
	if retried != 1 || state != AttachmentError {
		t.Fatal("retry did not remain an application callback")
	}
	disabled = true
	r.Frame()
	r.Click("Remove upload.zip")
	if removed != 0 {
		t.Fatal("disabled attachment child action activated")
	}
}

func TestAttachmentNativeFileDropIsOptInAndCopiesPaths(t *testing.T) {
	disabled := false
	var received [][]string
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		Attachment(c, AttachmentProps{Name: "Drop target", State: AttachmentIdle, Disabled: disabled, OnDropFiles: func(paths []string) { received = append(received, paths) }}).Width(320)
	}, 400, 140, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if !r.DropFiles(40, 30, "/tmp/a.pdf", "/tmp/b.png") || len(received) != 1 || !slices.Equal(received[0], []string{"/tmp/a.pdf", "/tmp/b.png"}) {
		t.Fatalf("native file drop did not reach callback once: %v", received)
	}
	r.Frame()
	if len(received) != 1 {
		t.Fatal("native drop replayed on later frame")
	}
	r.DropFiles(40, 30, "/tmp/c.pdf")
	if len(received) != 2 || received[0][0] != "/tmp/a.pdf" {
		t.Fatal("native frame storage invalidated retained application paths")
	}
	disabled = true
	r.Frame()
	if r.DropFiles(40, 30, "/tmp/blocked.pdf") || len(received) != 2 {
		t.Fatal("disabled attachment accepted native file drop")
	}
}

func TestAttachmentSourceFitWidthAndLongTitleKeepActionsReachable(t *testing.T) {
	removed := 0
	long := "a-very-long-report-title-that-must-truncate-before-the-actions.pdf"
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		ui.Column(c).Width(400).Gap(8).Children(func() {
			AttachmentWith(c, AttachmentProps{Name: "Fit card"}, func(a *AttachmentParts) {
				a.Content(func(*ui.Context) { a.Title("Tiny") })
			})
			Attachment(c, AttachmentProps{Name: long, OnRemove: func() { removed++ }}).Width(200)
		})
	}, 440, 180, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	fit, _ := r.Find("Fit card")
	if fit.W != 160 {
		t.Fatalf("source w-fit/min-w40 card stretched: %v", fit)
	}
	root, _ := r.Find(long)
	if root.W != 200 || root.H != 58 {
		t.Fatalf("long title wrapped or enlarged source card: %v", root)
	}
	button, ok := r.AccessByRole(ui.RoleButton, "Remove "+long)
	if !ok || button.Bounds.X+button.Bounds.W > root.X+root.W {
		t.Fatalf("truncation displaced native action: %v", button)
	}
	if err := r.Click("Remove " + long); err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatal("long text obscured native action")
	}
}

func TestAttachmentUsesCompleteCardTokensAndSourceErrorTint(t *testing.T) {
	for _, state := range []string{AttachmentDone, AttachmentError, AttachmentIdle} {
		t.Run(state, func(t *testing.T) {
			tokens := theme.Light()
			tokens.Card, tokens.CardForeground = ui.RGB(245, 235, 225), ui.RGB(20, 90, 140)
			tokens.Foreground, tokens.Muted, tokens.MutedForeground = ui.RGB(180, 20, 40), ui.RGB(190, 180, 170), ui.RGB(80, 70, 60)
			tokens.Destructive, tokens.Border = ui.RGB(200, 0, 0), ui.RGB(10, 120, 20)
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Scope(c, tokens).Fill().Children(func() {
					AttachmentWith(c, AttachmentProps{Name: "Colored attachment", State: state}, func(a *AttachmentParts) {
						a.Media(AttachmentIconMedia, nil).Label("Colored media")
						a.Content(func(*ui.Context) { a.Title("Native title"); a.Description("Native metadata") })
					}).Root.Width(320)
				})
			}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			root, _ := r.Find("Colored attachment")
			media, _ := r.Find("Colored media")
			card := r.Image().RGBAAt(int(root.X+root.W/2), int(root.Y+root.H-4))
			if colorDistance(card, color.RGBA{245, 235, 225, 255}) > 1 {
				t.Fatalf("root lost Card token: %v", card)
			}
			pixel := r.Image().RGBAAt(int(media.X+4), int(media.Y+media.H/2))
			want := color.RGBA{190, 180, 170, 255}
			if state == AttachmentError {
				want = attachmentComposite(tokens.Destructive.Alpha(.1), tokens.Card)
			}
			if colorDistance(pixel, want) > 2 {
				t.Fatalf("media lost semantic background or source error10%%: got=%v want=%v", pixel, want)
			}
			title, _ := r.Find("Native title")
			if attachmentClosestPixel(r.Image(), title, color.RGBA{20, 90, 140, 255}) > 4 {
				t.Fatal("title did not inherit CardForeground")
			}
			if state == AttachmentError {
				border := r.Image().RGBAAt(int(root.X+root.W/2), int(root.Y))
				want := attachmentComposite(tokens.Destructive.Alpha(.3), tokens.Card)
				if colorDistance(border, want) > 2 {
					t.Fatalf("source destructive30%% border: got=%v want=%v", border, want)
				}
			}
			if state == AttachmentIdle {
				gaps, strokes := 0, 0
				for x := int(root.X + 20); x < int(root.X+root.W-20); x++ {
					pixel := r.Image().RGBAAt(x, int(root.Y))
					if colorDistance(pixel, color.RGBA{245, 235, 225, 255}) < 2 {
						gaps++
					}
					if colorDistance(pixel, color.RGBA{10, 120, 20, 255}) < 2 {
						strokes++
					}
				}
				if gaps < 5 || strokes < 5 {
					t.Fatalf("source idle border was not truly dashed: gaps=%d strokes=%d", gaps, strokes)
				}
			}
		})
	}
}

func attachmentComposite(foreground, background ui.Color) color.RGBA {
	alpha := float64(foreground.A) / 255
	channel := func(front, back uint8) uint8 {
		return uint8(math.Round(float64(front)*alpha + float64(back)*(1-alpha)))
	}
	return color.RGBA{channel(foreground.R, background.R), channel(foreground.G, background.G), channel(foreground.B, background.B), 255}
}

func attachmentClosestPixel(picture *image.RGBA, rect ui.Rect, want color.RGBA) int {
	distance := 255
	for y := int(rect.Y); y < int(rect.Y+rect.H); y++ {
		for x := int(rect.X); x < int(rect.X+rect.W); x++ {
			distance = min(distance, colorDistance(picture.RGBAAt(x, y), want))
		}
	}
	return distance
}

func TestAttachmentPreservesAncestorFontAndIsIdleWhenDone(t *testing.T) {
	render := func(font string) *ui.Tester {
		return ui.NewTester(func(c *ui.Context) {
			theme.Apply(c)
			ui.Column(c).Font(font).Children(func() {
				AttachmentWith(c, AttachmentProps{Name: "Font sample"}, func(a *AttachmentParts) {
					a.Content(func(*ui.Context) { a.Title("iii WWW File metadata") })
				}).Root.Width(320)
			})
		}, 400, 140)
	}
	a, b := render("monospace"), render("serif")
	if bytes.Equal(a.Image().Pix, b.Image().Pix) {
		t.Fatal("attachment reset inherited native font")
	}
	if a.FramePending() || b.FramePending() {
		t.Fatal("ordinary done attachment scheduled motion")
	}
	a.Advance(time.Hour)
	if a.FramePending() {
		t.Fatal("ordinary attachment did not remain idle")
	}
}

func TestAttachmentInvalidConfigurationStaysFiniteAndProgressIsClamped(t *testing.T) {
	for _, test := range []struct {
		name     string
		progress float64
		want     string
		invalid  bool
	}{
		{"below zero", -1, "Uploading · 0%", false},
		{"above one", 2, "Uploading · 100%", false},
		{"not a number", math.NaN(), "Uploading · 0%", true},
		{"infinity", math.Inf(1), "Uploading · 0%", true},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := ui.NewTesterWith(func(c *ui.Context) {
				theme.Apply(c)
				Attachment(c, AttachmentProps{Name: "Progress", State: AttachmentUploading, Progress: test.progress}).Width(320)
			}, 400, 160, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
			root, ok := r.AccessByRole(ui.RoleGroup, "Progress")
			if !ok || root.Invalid != test.invalid || math.IsNaN(float64(root.Bounds.H)) || !r.HasText(test.want) {
				t.Fatalf("nonfinite progress leaked to native state: root=%v text=%v", root, r.Texts())
			}
		})
	}
}

func TestAttachmentProcessingShimmerChangesNativePixelsAndHonorsMotion(t *testing.T) {
	for _, size := range []AttachmentSize{AttachmentDefault, AttachmentSmall, AttachmentExtraSmall} {
		for _, direction := range []ui.Direction{ui.LeftToRight, ui.RightToLeft} {
			t.Run([]string{"default", "small", "extra small"}[size]+"/"+map[ui.Direction]string{ui.LeftToRight: "LTR", ui.RightToLeft: "RTL"}[direction], func(t *testing.T) {
				state, instant := AttachmentProcessing, false
				r := ui.NewTester(func(c *ui.Context) {
					theme.Apply(c)
					if instant {
						c.DisableAnimations()
					}
					ui.Column(c).Font("monospace").Children(func() {
						AttachmentWith(c, AttachmentProps{ID: "phase", Name: "Processing card", State: state, Size: size}, func(a *AttachmentParts) {
							a.Content(func(*ui.Context) { a.Title("Processing native attachment glyphs") })
						}).Root.Width(340).FontSize(22)
					})
				}, 400, 140)
				r.SetDirection(direction)
				before := append([]byte(nil), r.Image().Pix...)
				title, _ := r.Find("Processing native attachment glyphs")
				if title.H != 27.5 {
					t.Fatalf("active shimmer reset inherited source font: %v", title)
				}
				if !r.FramePending() || !r.HasText("Processing native attachment glyphs") {
					t.Fatal("processing shimmer lost native text or animation")
				}
				r.Advance(700 * time.Millisecond)
				if bytes.Equal(before, r.Image().Pix) {
					t.Fatal("processing shimmer scheduled frames without changing real native glyph pixels")
				}
				r.SetPreferences(ui.Preferences{ReduceMotion: true})
				plain := append([]byte(nil), r.Image().Pix...)
				if r.FramePending() {
					t.Fatal("ReducedMotion left shader animation running")
				}
				state = AttachmentDone
				r.Frame()
				if !bytes.Equal(plain, r.Image().Pix) {
					t.Fatal("ReducedMotion shimmer did not use the same plain native glyphs as done")
				}
				titleDone, _ := r.Find("Processing native attachment glyphs")
				if titleDone != title {
					t.Fatalf("state change changed inherited native text geometry: processing=%v done=%v", title, titleDone)
				}
				r.Advance(time.Hour)
				if r.FramePending() {
					t.Fatal("done attachment did not remain idle")
				}
				state, instant = AttachmentProcessing, true
				r.SetPreferences(ui.Preferences{})
				if r.FramePending() || !bytes.Equal(plain, r.Image().Pix) {
					t.Fatal("Instant attachment did not remain plain and idle")
				}
			})
		}
	}
}
