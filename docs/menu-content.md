# Menu content geometry and selection

`overlay.MenuItem` is shared by styled desktop menus, context menus and menubars. `Inset` changes a row's logical start padding from 8 to 32 pixels; `GroupInset` independently changes the group caption introduced by that item. These values follow the pinned Vega `px-2` and `data-inset:pl-8` source and mirror with the current direction. Defaults remain 8 pixels.

Checkbox and radio rows reserve logical end padding 32 pixels. The selected state uses a decorative 16-pixel shared Check SVG at logical end 8 pixels. An independent `Icon` stays at the start of the row, including checked and radio rows. NativeOS menu presentation ignores visual inset flags and uses its platform images/state rendering.

Styled desktop action rows close after activation. Checkbox and radio rows remain open by default, following Base UI 1.6.0 `closeOnClick=false`. `CloseOnSelect *bool` can override an individual row. `MenuContentProps.KeepOpenOnSelect` (and corresponding Menu/Menubar props) takes priority over any request to close. `desktop.Select` explicitly closes its internal radio-backed options when it commits a value.

NativeOS menus and legacy `overlay.InWindow*` menus retain MyGo's menu-selection tracking; `CloseOnSelect` does not override that lifecycle. Legacy styled rows share the inset and indicator geometry. Business actions should update application-owned data through the current callback. Persist models and public handles; never retain a frame Context or Element.

The source defaults and originals are recorded in `pkg/desktop/testdata/shadcn-menu/close-on-click/manifest.json`. This API adds no font dependency.
