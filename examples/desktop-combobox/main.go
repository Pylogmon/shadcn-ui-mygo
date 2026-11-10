// Run with go run ./examples/desktop-combobox.
package main

import (
	"flag"
	"fmt"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

type project struct{ ID, Name, Stack string }
type appState struct {
	host                *desktop.Host
	appearance          theme.Controller
	single, multiple    control.ComboboxState[project]
	projects            []project
	native, cancelClose bool
	message             string
}

// Providers capture the application model. Every handler belongs to this build's
// current window; no parent Context, Element or ComboboxScope crosses windows.
func (a *appState) props(window *ui.Context, multiple bool) control.ComboboxRootProps[project] {
	return control.ComboboxRootProps[project]{
		Items: a.projects, Multiple: multiple, AutoHighlight: control.ComboboxHighlightOnInput,
		Adapter: control.ComboboxAdapter[project]{
			Key: func(p project) string { return p.ID }, Label: func(p project) string { return p.Name },
			FormValue: func(p project) string { return p.ID },
		},
		Filter: func(p project, query string) bool {
			return strings.Contains(strings.ToLower(p.Name+" "+p.Stack), strings.ToLower(strings.TrimSpace(query)))
		},
		OnValueChange: func(value control.ComboboxSelection[project], _ *control.ComboboxChange) {
			a.message = fmt.Sprintf("Single callback: %s (valid: %t)", value.Value.ID, value.Valid)
			window.Invalidate()
		},
		OnValuesChange: func(values []project, _ *control.ComboboxChange) {
			a.message = fmt.Sprintf("Multiple callback: %d projects", len(values))
			window.Invalidate()
		},
		OnOpenChange: func(open bool, event *control.ComboboxChange) {
			if !open && a.cancelClose {
				event.Cancel()
				a.message = "Close canceled. Turn off Cancel popup close to dismiss."
				window.Invalidate()
			}
		},
	}
}
func (a *appState) singleView(window *ui.Context) control.ComboboxRootProps[project] {
	return a.props(window, false)
}
func (a *appState) multipleView(window *ui.Context) control.ComboboxRootProps[project] {
	return a.props(window, true)
}
func (a *appState) combo(c *ui.Context, key string, state *control.ComboboxState[project], view desktop.ComboboxView[project]) {
	draw := desktop.Combobox[project]
	if a.native {
		draw = desktop.NativeCombobox[project]
	}
	_, err := draw(c, a.host, key, state, desktop.ComboboxProps{
		Label: key, Placeholder: "Choose projects", SearchPlaceholder: "Search name or stack…", ShowClear: true,
		Popup: desktop.PopoverProps{Width: 360, Height: 260, MaxHeight: 360},
	}, view)
	if err != nil {
		a.message = err.Error()
	}
}
func (a *appState) view(c *ui.Context) {
	a.appearance.Apply(c)
	a.host.BeginFrame()
	defer a.host.EndFrame()
	t := theme.TokensOf(c)
	ui.Scroll(c).Fill().Background(t.Background).TextColor(t.Foreground).Children(func() {
		ui.Column(c).FillWidth().Padding(24).Gap(12).Children(func() {
			ui.Text(c, "Projects in a separate window").FontSize(20).FontWeight(600)
			ui.Text(c, "Search by name or stack. The popup can extend beyond this window.").TextColor(t.MutedForeground)
			control.Switch(c, &a.native, control.SwitchProps{Label: "Platform theme (NativeCombobox)"})
			control.Switch(c, &a.cancelClose, control.SwitchProps{Label: "Cancel popup close"})
			control.Label(c, "Single project")
			a.combo(c, "Single project", &a.single, a.singleView)
			single := "none"
			if a.single.Value.Valid {
				single = a.single.Value.Value.ID
			}
			ui.Text(c, "Selected ID: "+single).FontSize(13)
			control.Label(c, "Multiple projects")
			a.combo(c, "Multiple projects", &a.multiple, a.multipleView)
			ids := make([]string, 0, len(a.multiple.Values))
			for _, p := range a.multiple.Values {
				ids = append(ids, p.ID)
			}
			ui.Text(c, "Selected IDs: "+strings.Join(ids, ", ")).FontSize(13)
			ui.Text(c, a.message).FontSize(13).TextColor(t.MutedForeground).Role(ui.RoleStatus)
		})
	})
}

func main() {
	snapshot := flag.String("snapshot", "", "write the closed parent window as a PNG")
	native := flag.Bool("native", false, "start with platform-themed controls")
	flag.Parse()
	a := &appState{native: *native, message: "Values stay local to this example.", projects: []project{
		{"atlas", "Atlas", "Go"}, {"beacon", "Beacon", "Rust"}, {"cedar", "Cedar", "TypeScript"},
		{"delta", "Delta", "Python"}, {"ember", "Ember", "Go"}, {"fjord", "Fjord", "Swift"},
		{"grove", "Grove", "Kotlin"}, {"harbor", "Harbor", "Rust"}, {"iris", "Iris", "TypeScript"},
		{"juniper", "Juniper", "Go"}, {"kite", "Kite", "Python"}, {"lumen", "Lumen", "Swift"},
	}}
	a.single.Value = control.ComboboxSelection[project]{Value: a.projects[0], Valid: true}
	a.multiple.Values = []project{a.projects[1], a.projects[4]}
	if *snapshot != "" {
		f, err := os.Create(*snapshot)
		if err != nil {
			log.Fatal(err)
		}
		err = png.Encode(f, ui.Render(func(c *ui.Context) { c.DisableAnimations(); a.view(c) }, 520, 440, 1))
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			log.Fatal(err, closeErr)
		}
		return
	}
	mygo.App.WhenReady(func() {
		parentWindow := mygo.NewWindow(mygo.WindowOptions{Title: "Desktop Combobox", Width: 520, Height: 440, Hidden: true, Content: ui.View(a.view)})
		a.host = desktop.NewHost(parentWindow)
		parentWindow.Show()
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
