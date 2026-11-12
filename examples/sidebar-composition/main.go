// Sidebar content and dimensions belong to the application; the library owns
// native actions, readable colors, state feedback and accessibility.
package main

import (
	"flag"
	"image/png"
	"log"
	"os"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/navigation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

var icon = ui.MustParseSVG([]byte(`<svg viewBox="0 0 16 16" fill="currentColor"><path d="M2 2h12v12H2z"/></svg>`))

type state struct {
	tokens   *theme.Tokens
	selected string
}

func (s *state) view(c *ui.Context) {
	t := s.tokens
	c.SetTheme(t.Theme())
	ui.Column(c).Fill().Padding(24).Gap(20).Background(t.Sidebar).TextColor(t.SidebarForeground).Role(ui.RoleNavigation).Label("MangGo settings").Children(func() {
		theme.WithSurface(c, t.Sidebar, func() {
			ui.Text(c, "MangGo · Amber").FontSize(20).FontWeight(600)
			ui.Text(c, "Navigation / 14 DIP").TextColor(t.MutedForeground)
			for _, label := range []string{"General", "Translation", "Advanced"} {
				p := navigation.SidebarMenuButtonProps{Label: label, Active: s.selected == label, FontSize: 14}
				if p.Active {
					p.Color = &t.Primary
				}
				navigation.SidebarMenuButton(c.Key(label), p, func(c *ui.Context) {
					// MangGo-specific fixed icon column and centered remaining label region.
					ui.Box(c).Width(28).Shrink(0).Center().Children(func() { ui.Icon(c, icon).Size(16, 16).AccessHidden(true) })
					ui.Text(c, label).Grow(1).Basis(0).MinWidth(0).TextAlign(ui.Center).SingleLine()
				}).Role(ui.RoleLink).FillWidth().Height(36).Padding(0, 10).Border(1, ui.Transparent).Gap(0).FontWeight(400).OnClick(func() { s.selected = label })
			}
			display.Separator(c, false)
			// Pro remains colored at rest and in interactive states. All visual
			// children inherit one state foreground, including the inert badge.
			navigation.SidebarMenuButton(c.Key("pro"), navigation.SidebarMenuButtonProps{Label: "Pro footer", Color: &t.Primary}, func(c *ui.Context) {
				display.BadgeWith(c, "Pro", display.BadgeProps{Variant: display.BadgeOutline, InheritForeground: true})
				ui.Text(c, "Professional Version").SingleLine()
			}).FillWidth().Height(36).Padding(0, 10).Border(1, ui.Transparent).Gap(6).FontWeight(500).Justify(ui.Center)
			// Free follows neutral status semantics: muted at rest, Foreground when
			// hovered, pressed or focused. The badge shares exactly the same color.
			navigation.SidebarMenuButton(c.Key("free"), navigation.SidebarMenuButtonProps{Label: "Free footer", Muted: true}, func(c *ui.Context) {
				display.BadgeWith(c, "Free", display.BadgeProps{Variant: display.BadgeOutline, InheritForeground: true})
				ui.Text(c, "Free Version").SingleLine()
			}).FillWidth().Height(36).Padding(0, 10).Border(1, ui.Transparent).Gap(6).FontWeight(500).Justify(ui.Center)
			ui.Text(c, "Soft / Ghost / Outline / Link").FontSize(12).TextColor(t.MutedForeground)
			ui.Row(c).Gap(8).Children(func() {
				for _, variant := range []control.Variant{control.Soft, control.Ghost, control.Outline, control.Link} {
					control.Button(c, "Aa", control.ButtonProps{Variant: variant, Color: &t.Primary, Size: control.ExtraSmall})
				}
			})
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "write a PNG instead of opening a window")
	dark := flag.Bool("dark", false, "use the dark preset")
	interaction := flag.String("state", "idle", "idle, hover, pressed, focus, pro-hover, free-hover or free-focus")
	flag.Parse()
	s := &state{tokens: theme.Preset("amber", *dark), selected: "Translation"}
	if *snapshot != "" {
		r := ui.NewTesterWith(s.view, 400, 480, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
		name := "Translation"
		if *interaction == "pro-hover" {
			name = "Pro footer"
		}
		if *interaction == "free-hover" || *interaction == "free-focus" {
			name = "Free footer"
		}
		b, ok := r.Find(name)
		if !ok {
			log.Fatal("missing snapshot target")
		}
		x, y := b.X+b.W/2, b.Y+b.H/2
		switch *interaction {
		case "idle":
		case "hover", "pro-hover", "free-hover":
			r.Move(x, y)
		case "pressed":
			r.Press(x, y)
		case "focus", "free-focus":
			for i := 0; i < 6 && !r.Focused(name); i++ {
				r.Key(0, ui.KeyTab)
			}
		default:
			log.Fatalf("unknown state %q", *interaction)
		}
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		if err = png.Encode(f, r.Image()); err != nil {
			log.Fatal(err)
		}
		if err = f.Close(); err != nil {
			log.Fatal(err)
		}
		return
	}
	mygo.App.WhenReady(func() {
		window := mygo.NewWindow(mygo.WindowOptions{Title: "Readable sidebar composition", Width: 400, Height: 480, Hidden: true, Content: ui.View(s.view)})
		window.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
