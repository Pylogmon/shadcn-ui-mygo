# Native tooltip composition

Run `go run ./examples/tooltip-composition` or `make tooltips-app` on macOS.
The eight fixed source configurations include six logical/physical sides,
icons, long text, a wrapper around a disabled control, a real Kbd, a link and
formatted paragraphs. Hover or focus a trigger; Escape dismisses it. Update
status changes real child content and its semantic description. Platform theme
switches to NativeTooltipWith, which uses the independent window's native theme.

TooltipWith content must build with its child Context and must contain inert
visuals. Description supplies the trigger's accessibility description; Revision
refreshes content changed by application state. NativeTooltipWith is a
platform-themed Flyout, not an OS tooltip widget. A backend that cannot report
actual window geometry omits the styled arrow. Source data and style references
are archived in `pkg/desktop/testdata/shadcn-tooltip`.
