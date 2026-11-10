# Desktop Combobox

Run `go run ./examples/desktop-combobox`. The small parent window contains two
independent `ComboboxState[project]` models: typed single selection and typed
multiple selection. Open either trigger and search a project name or stack
(for example, `go`). The editor and ordinary collection live in a real,
independent Flyout window, so they can extend beyond the parent window. The
parent shows the selected business IDs; callbacks only update local status.

The platform-theme switch chooses `desktop.NativeCombobox`; the other setting
uses `desktop.Combobox` and the shared shadcn theme. NativeCombobox uses MyGo's
platform theme and native base controls in a plain Flyout. It is not an
operating-system ComboBox or a static OS menu. Both keep a real editable search
field and the same typed value model. Application and popup fonts inherit their
respective themes.

Enable **Cancel popup close**, open a popup, type a query and click outside it.
The current `OnOpenChange` handler cancels the proposed close and preserves the
editor. This switch cancels all close requests, including Escape and a single
selection's automatic close; turn it off to dismiss. Selected values still
update. `Host` is bound explicitly to `parentWindow`, and the parent view calls
`BeginFrame`/`EndFrame`. Each provider receives its current parent or child
context and builds handlers there; it captures no parent context, element or
combobox scope across windows.

```sh
go run ./examples/desktop-combobox -native
go run ./examples/desktop-combobox -snapshot /tmp/desktop-combobox.png
```

`-snapshot` renders only the closed parent window. Opening a child requires the
desktop host; the snapshot does not draw a fake child popup on the parent canvas.
