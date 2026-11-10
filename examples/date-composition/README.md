# Calendar and editable dates

Run `go run ./examples/date-composition`. This native application combines a
Persian/Gregorian calendar, an exact-date editor, and an application-defined
relative-date editor. The fixed reference date is June 10, 2025 in Asia/Tehran.
Changing the calendar model or browsing a month preserves the selected instant;
choosing a day preserves the local clock.

| Fixed source example | Native operation |
| --- | --- |
| `calendar-hijri.tsx` | Persian/Jalali model, original Vazirmatn font, faIR digits, RTL and month/year menus; toggle Gregorian without changing the selected date. |
| `date-picker-input.tsx` | Edit an ISO or long English date, open with ArrowDown, browse and choose a day. |
| `date-picker-natural-language.tsx` | Parse `tomorrow`, `next week`, `in N days`, or an exact date using the application's explicit callback. |

Keep `DatePickerInputState` at a stable address, initialize `Text` explicitly,
and supply `Parse`. `Format` controls the text published after calendar
selection. Invalid text remains editable and reports `state.Error` while the
last valid date and month remain available. `parts.Editor` is the real native
editor and supports `OnSubmit`; `parts.Root` composes with `control.Field`.
`parts.ValueChanged` reports accepted date changes, while the root's native
`Changed` also reports raw text edits. Callbacks are read from the current build.

The relative parser is deliberately small application code, not a replacement
for the source's complete `chrono-node` grammar. Applications can connect a
different parser without changing the editor or calendar. Persian conversion
follows the pinned source's arithmetic solar model, rather than a lunar Hijri
or astronomical calendar.

For a native render, use `-snapshot /tmp/dates.png`; add `-dark` and/or `-open`
to show dark appearance or the exact-date popup. The tests cover editing,
parsing, month navigation, selection feedback, native composition and focus.
Real Cocoa cancellation and other-platform runtime/assistive-technology behavior
remain separate platform checks.
