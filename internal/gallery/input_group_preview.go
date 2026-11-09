package gallery

import (
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/icons/material/outlined"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

const inputGroupInitialDraft = "console.log('Hello, world!');"

// App-owned values survive Gallery page changes; only Handle/MenuState values
// cross frames. Popup builders use their own live child Context, and clipboard
// completion captures only this model plus the requested values.
type inputGroupPreviewState struct {
	ready, native                                            bool
	link, site, fileName, query, target, draft               string
	favorite, infoOpen, settingsOpen, detailsOpen, monospace bool
	targetOpen, fileOpen                                     bool
	targetMenu, fileMenu                                     desktop.MenuState
	draftHandle                                              ui.Handle
	preview                                                  string
	previewed                                                bool
	results                                                  []string
	searched                                                 bool
	feedback, popupError                                     string
	copySerial                                               uint64
	copyPending, copyKind, copyText                          string
}

func (p *inputGroupPreviewState) initialize(native bool) {
	if p.ready {
		return
	}
	p.ready = true
	p.native = native
	p.link = "https://x.com/shadcn"
	p.site = "example.com"
	p.fileName = "script.js"
	p.target = "Documentation"
	p.draft = inputGroupInitialDraft
	p.monospace = true
}
func (p *inputGroupPreviewState) copy(c *ui.Context, text, kind string) {
	p.copySerial++
	serial := p.copySerial
	p.copyPending = kind
	p.copyKind = ""
	p.feedback = "Copying " + kind + "…"
	c.WriteClipboardThen(text, func(err error) {
		if p.copySerial != serial {
			return
		}
		p.copyPending = ""
		if err != nil {
			p.copyKind = ""
			p.feedback = "Could not copy " + kind + ". Try again."
			return
		}
		p.copyKind, p.copyText = kind, text
		p.feedback = "Copied " + kind + "."
	})
}
func (p *inputGroupPreviewState) copyIcon(kind, text string) *ui.SVG {
	if p.copyKind == kind && p.copyText == text {
		return outlined.Check()
	}
	return outlined.ContentCopy()
}
func (p *inputGroupPreviewState) search() {
	catalog := map[string][]string{
		"Documentation": {"Input groups and addons", "Native text selection", "Field labels and errors"},
		"Blog Posts":    {"Building native forms", "Designing useful previews"},
		"Changelog":     {"Native input improvements", "Menu appearance updates"},
	}
	query := strings.ToLower(strings.TrimSpace(p.query))
	p.results = nil
	p.searched = true
	for _, text := range catalog[p.target] {
		if query == "" || strings.Contains(strings.ToLower(text), query) {
			p.results = append(p.results, text)
		}
	}
	p.feedback = fmt.Sprintf("%d result(s) in %s.", len(p.results), p.target)
}
func (p *inputGroupPreviewState) menu(c *ui.Context, h *desktop.Host, native bool, key, label, caption string, icon *ui.SVG, items []overlay.MenuItem, open *bool, state *desktop.MenuState) string {
	props := control.ButtonProps{Label: label, Variant: control.Ghost, Size: control.ExtraSmall, Icon: icon}
	if caption == "" {
		props.Size = control.IconExtraSmall
	} else {
		props.Icon = nil
	}
	trailingIcon := func(e ui.Element) {
		if caption != "" && icon != nil {
			e.Children(func() {
				material.Icon(c, icon, material.Props{Size: 12})
			})
		}
	}
	var chosen string
	var err error
	if native {
		chosen, err = desktop.NativeMenu(c, overlay.MenuProps{Items: items}, func(e ui.Element) {
			control.StyleButton(c, e, caption, props)
			trailingIcon(e)
			if e.Clicked() {
				p.popupError = ""
			}
		})
	} else {
		anchor := control.InputGroupButton(c, caption, props)
		trailingIcon(anchor)
		if anchor.Clicked() {
			p.popupError = ""
		}
		chosen, err = desktop.DropdownMenu(c, h, key, anchor, open, state, desktop.MenuProps{Label: label, Items: items, Placement: mygo.PlacementBottomEnd, Gap: 8, Width: 210})
	}
	if err != nil {
		p.popupError = "Could not open menu: " + err.Error()
	}
	return chosen
}
func (p *inputGroupPreviewState) view(c *ui.Context, h *desktop.Host, native bool) {
	p.initialize(native)
	if p.native != native {
		p.native = native
		p.targetOpen, p.fileOpen, p.infoOpen = false, false, false
		p.targetMenu.Reset()
		p.fileMenu.Reset()
		if h != nil {
			h.Close("input-group-target")
			h.Close("input-group-file")
			h.Close("input-group-info")
		}
	}
	t := theme.TokensOf(c)
	display.Card(c.Key("input-group-source-examples"), func() {
		ui.Column(c).FillWidth().Gap(18).Children(func() {
			display.Typography(c, "Input groups with actions", display.H3)
			ui.Text(c, "Copy a link, choose where to search, and preview an editable draft.").TextColor(t.MutedForeground)
			ui.Row(c).FillWidth().Wrap().Gap(24).AlignItems(ui.Start).Children(func() {
				ui.Column(c).Basis(0).Grow(1).MinWidth(300).Gap(18).Children(func() {
					control.Field(c, control.FieldProps{Label: "Share link"}, func() ui.Element {
						return control.InputGroup(c.Key("share-link"), control.InputGroupProps{InlineEnd: func() {
							if control.InputGroupButton(c, "", control.ButtonProps{Label: "Copy link", Icon: p.copyIcon("link", p.link), Variant: control.Ghost, Size: control.IconExtraSmall, Loading: p.copyPending == "link"}).Clicked() {
								p.copy(c, p.link, "link")
							}
						}}, func() ui.Element { return control.Input(c, &p.link, control.InputProps{ReadOnly: true}) })
					})
					control.Field(c, control.FieldProps{Label: "Favorite site"}, func() ui.Element {
						return control.InputGroup(c.Key("favorite-site"), control.InputGroupProps{InlineStart: func() {
							anchor := control.InputGroupButton(c, "", control.ButtonProps{Label: "Site information", Icon: outlined.Info(), Variant: control.Secondary, Size: control.IconExtraSmall})
							if anchor.Clicked() {
								p.popupError = ""
							}
							if h == nil {
								p.infoOpen = false
								if anchor.Clicked() {
									p.popupError = "Could not open site information. Open this preview in the Gallery window and try again."
								}
							} else {
								build := desktop.Popover
								if native {
									build = desktop.NativePopover
								}
								site := p.site
								if err := build(c, h, "input-group-info", anchor, &p.infoOpen, desktop.PopoverProps{Label: "Site information", Width: 280, FitContent: true}, func(child *ui.Context) {
									ui.Column(child).Gap(8).Children(func() {
										ui.Text(child, "Example address").FontWeight(600)
										ui.Text(child, "https://"+site)
										ui.Text(child, "Use the star to keep this address in your favorites.")
									})
								}); err != nil {
									p.popupError = "Could not open site information: " + err.Error()
								}
							}
							ui.Text(c, "https://")
						}, InlineEnd: func() {
							fill := material.Unfilled
							if p.favorite {
								fill = material.Filled
							}
							if control.InputGroupButton(c, "", control.ButtonProps{Label: "Favorite site toggle", Icon: outlined.Star(fill), Variant: control.Ghost, Size: control.IconExtraSmall}).Checked(p.favorite).Clicked() {
								p.favorite = !p.favorite
								if p.favorite {
									p.feedback = "Site added to favorites."
								} else {
									p.feedback = "Site removed from favorites."
								}
							}
						}}, func() ui.Element { return control.Input(c, &p.site, control.InputProps{Placeholder: "example.com"}) })
					})
					control.Field(c, control.FieldProps{Label: "Search query", Description: "Search the small local example collection."}, func() ui.Element {
						return control.InputGroup(c.Key("target-search"), control.InputGroupProps{InlineStart: func() { material.Icon(c, outlined.Search(), material.Props{Size: 16}) }, InlineEnd: func() {
							items := []overlay.MenuItem{{Key: "Documentation", Label: "Documentation"}, {Key: "Blog Posts", Label: "Blog Posts"}, {Key: "Changelog", Label: "Changelog"}}
							if target := p.menu(c, h, native, "input-group-target", "Search target", p.target, outlined.ExpandMore(), items, &p.targetOpen, &p.targetMenu); target != "" {
								p.target = target
								p.searched = false
								p.results = nil
								p.feedback = "Search target: " + target + "."
							}
							if control.InputGroupButton(c, "Search", control.ButtonProps{Variant: control.Secondary}).Clicked() {
								p.search()
							}
						}}, func() ui.Element {
							return control.Input(c, &p.query, control.InputProps{Placeholder: "Type to search…", OnSubmit: func(string) { p.search() }})
						})
					})
					if p.searched {
						ui.Column(c.Key("input-group-results")).Gap(6).Children(func() {
							ui.Text(c, fmt.Sprintf("%s · %d result(s)", p.target, len(p.results))).FontWeight(500)
							for _, result := range p.results {
								ui.Text(c, result)
							}
							if len(p.results) == 0 {
								ui.Text(c, "No matching examples.").TextColor(t.MutedForeground)
							}
						})
					}
				})
				ui.Column(c).Basis(0).Grow(1).MinWidth(340).Gap(16).Children(func() {
					control.Field(c, control.FieldProps{Label: "Draft file name"}, func() ui.Element {
						return control.InputGroup(c.Key("draft-file-name"), control.InputGroupProps{InlineEnd: func() {
							switch p.menu(c, h, native, "input-group-file", "Draft file actions", "", outlined.MoreHoriz(), []overlay.MenuItem{{Key: "settings", Label: "Settings"}, {Key: "copy", Label: "Copy file name"}, {Key: "details", Label: "Show file details"}}, &p.fileOpen, &p.fileMenu) {
							case "settings":
								p.settingsOpen = !p.settingsOpen
							case "copy":
								p.copy(c, p.fileName, "file name")
							case "details":
								p.detailsOpen = !p.detailsOpen
							}
						}}, func() ui.Element {
							return control.Input(c, &p.fileName, control.InputProps{Placeholder: "Enter file name"})
						})
					})
					control.Field(c, control.FieldProps{Label: "Draft text"}, func() ui.Element {
						return control.InputGroup(c.Key("draft-editor"), control.InputGroupProps{BlockStart: func() {
							material.Icon(c, outlined.Code(), material.Props{Size: 16})
							ui.Text(c, p.fileName).Font("monospace").SingleLine().Grow(1).MinWidth(0)
							if control.InputGroupButton(c, "", control.ButtonProps{Label: "Reset draft", Icon: outlined.RestartAlt(), Variant: control.Ghost, Size: control.IconExtraSmall}).Clicked() {
								p.draft = inputGroupInitialDraft
								p.previewed = false
								p.feedback = "Draft restored."
							}
							if control.InputGroupButton(c, "", control.ButtonProps{Label: "Copy draft", Icon: p.copyIcon("draft", p.draft), Variant: control.Ghost, Size: control.IconExtraSmall, Loading: p.copyPending == "draft"}).Clicked() {
								p.copy(c, p.draft, "draft")
							}
						}, BlockEnd: func() {
							_, end := c.Resolve(p.draftHandle).TextSelection()
							units := utf16.Encode([]rune(p.draft))
							prefix := string(utf16.Decode(units[:max(0, min(end, len(units)))]))
							line := strings.Count(prefix, "\n") + 1
							column := utf8.RuneCountInString(prefix[strings.LastIndex(prefix, "\n")+1:]) + 1
							ui.Text(c, fmt.Sprintf("Line %d, Column %d", line, column)).FontSize(12).TextColor(t.MutedForeground).Grow(1)
							if control.InputGroupButton(c, "Preview draft", control.ButtonProps{Variant: control.Default, Icon: outlined.Preview()}).Clicked() {
								p.preview, p.previewed = p.draft, true
								p.feedback = "Draft preview updated."
							}
						}}, func() ui.Element {
							return control.Textarea(c, &p.draft, control.TextareaProps{Placeholder: inputGroupInitialDraft, Rows: 7, Handle: &p.draftHandle}).MinHeight(200)
						})
					})
					if p.settingsOpen {
						control.Checkbox(c, &p.monospace, "Monospace preview", control.CheckboxProps{})
					}
					if p.detailsOpen {
						ui.Text(c, fmt.Sprintf("%s · %d characters · %d lines", p.fileName, utf8.RuneCountInString(p.draft), strings.Count(p.draft, "\n")+1)).TextColor(t.MutedForeground)
					}
					if p.previewed {
						ui.Column(c.Key("draft-preview")).FillWidth().Gap(8).Padding(12).Border(1, t.Border).Radius(t.RadiusMD()).Children(func() {
							ui.Text(c, "Draft preview").FontWeight(600)
							text := p.preview
							if text == "" {
								text = "(Empty draft)"
							}
							e := ui.Text(c, text).MinWidth(0)
							if p.monospace {
								e.Font("monospace")
							}
						})
					}
				})
			})
			if p.feedback != "" {
				ui.Text(c, p.feedback).Role(ui.RoleStatus)
			}
			if p.popupError != "" {
				ui.Text(c, p.popupError).Role(ui.RoleAlert)
			}
		})
	}).Role(ui.RoleGroup).Label("Input group examples")
}
