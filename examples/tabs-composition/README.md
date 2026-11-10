# Tabs composition

Run the real native example:

```sh
go run ./examples/tabs-composition
```

Choose any of the 13 fixed examples. The input example has a working search field
and action; the dropdown example has a real menu. Panel notes stay in application
state across tab changes. Manual activation, no-loop and RTL controls exercise
the native keyboard behavior.

Render this same view without opening a window:

```sh
go run ./examples/tabs-composition -recipe input -snapshot /tmp/tabs-input.png
go run ./examples/tabs-composition -recipe vertical -dark -snapshot /tmp/tabs-vertical.png
```

TabsRoot owns the scope, TabsList only owns its trigger group, and TabsContent can
sit in a separate header/body layout. The example uses the public in-window menu
adapter; a desktop Flyout or NativeMenu can occupy the same header slot.
Icon-only triggers use HideLabel with a useful native Label. TabsMultiple remains
single-selection. The original int-based Tabs convenience API remains available.


## Gallery source catalogue

The Gallery Navigation page exposes **Tabs configuration** with all six docs
previews, the guide's distinct inline Usage example and all thirteen Create
sections: **20 independent recipes**. The fixed revision is
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`.
[`source-map.json`](source-map.json) records eight original example/guide files
and one style reference with byte lengths, SHA-256, Git blobs and pinned URLs.

| Source family | Material retained |
| --- | --- |
| Docs dashboard | Four native triggers and the complete Overview/Analytics/Reports/Settings Card titles, descriptions and bodies, default Overview and400DIP maximum |
| Guide Usage | Account/Password triggers with the original plain panel text and400DIP maximum |
| Docs line/vertical/disabled/icons | Exact labels/defaults and no invented panel content |
| Translated dashboard | Verbatim English/Arabic/Hebrew strings, Arabic default and384DIP maximum |
| Create basic/line/disabled/icons/icon-only/multiple | Exact source values/defaults and layouts; Multiple still means multiple triggers in a single-selection Tabs |
| Variants Alignment | Two independent Roots, default/line variants, original16DIP gap, nonwrapping native horizontal viewport |
| With Content/Line With Content | Account/Password/Notifications wording,24DIP Vega padding and rounded border |
| Line Disabled | Reports is actually disabled and skipped by native keys |
| With Dropdown | Real native More options trigger, Settings/Export/Archive and the three original dashboard panels |
| Vertical | Actual vertical tablist, bordered body and distinct longer Password/Notifications text |
| With Input and Button | Actual176DIP native Search input, opposite-edge Action,512DIP maximum and the source's untriggered Reports content declaration |

The Gallery uses public `navigation.TabsRoot`, `TabsList`, `TabsTrigger` and
`TabsContent`. A real native tablist owns keys, focus, selected state and
Controls/LabelledBy relationships. Only the active source panel is built.
Icon-only triggers keep useful native Home/Search/Settings names without
showing text. Existing on-demand Material SVG equivalents replace source icon
placeholders; no icon font or new asset is introduced.

Source examples have no extra note fields or search transport. The actual
Action button records an application-memory snapshot of query and tab; Enter
has no added submission handler. The More menu records only local intent.
Configuration/language menus and More follow Gallery's styled independent
Flyout or system-menu setting. The existing standalone program above retains
its separate demonstration controls and panel-note composition.

Unspecified Create wrappers use a400DIP native maximum. Source headers retain
a400/512DIP native minimum inside a real horizontal viewport; this preserves
the176DIP input and opposite-edge actions when the parent is narrow. Dashboard
lists also remain horizontally reachable, without wrapping the source tabs.
Font shaping and natural tab widths use native metrics. Source body and
comparison dimensions are retained; software images do not claim pixel parity
with every web style/font combination.

Both compound and legacy Tabs previously applied native automatic dim on top
of explicit opacity. The source requires one half-opacity layer. The fix keeps
explicit disabled at0.5, preserves inherited native dim and still rejects
pointer/keyboard input; high-contrast actual pixels are255 enabled and128 for
either explicit or ancestor-disabled. That check extends the existing native
orientation/disabled workflow rather than adding a second matrix.

`TestTabsCompositionsSourceWorkflow` is one concentrated public-input workflow
through all20 recipes, actual pointer/keyboard selection, axis/disabled
navigation, independent Roots, native header editing, local actions, both real
menu routes, untriggered content, narrow Hebrew and page/style retention. Set
`GALLERY_TABS_COMPOSITION_SNAPSHOT_DIR` for four representative software PNGs.
Application state contains only strings, maps, counters and public menu state;
no Context, Element, Tabs scope or frame callback survives a view. Real
OS/IME/AT, fonts, display scale and full platform/visual acceptance remain.
