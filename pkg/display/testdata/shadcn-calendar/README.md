# Fixed calendar and editable-date references

`manifest.json` records 28 byte-exact upstream files and their SHA256 values.
The shadcn examples and documentation come from commit
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The npm versions and integrity
values match that commit's lockfile: `react-day-picker@9.8.1` and
`date-fns-jalali@4.1.0-0`. The original MIT licenses are included.

Despite its filename, `calendar-hijri.tsx` imports
`react-day-picker/persian`. Its model is the Persian/Jalali **solar** calendar.
The pinned package uses the arithmetic 33-year algorithm in `_lib/jalali.js`,
rather than an astronomical or Islamic lunar calendar. The Go conversion is
adapted from that file and supports complete positive Persian years 1–275138
within the source JavaScript Date range. Values remain real `time.Time` dates;
local time, location, existing Gregorian selection and DST handling remain in
the calendar's shared path.

The source Persian wrapper defaults to `faIR`, RTL, Saturday-first weeks and
Eastern Arabic-Indic digits. The native model uses those defaults and lets the
caller override direction and the first weekday. The runnable application
registers the original Vazirmatn family explicitly through `pkg/fonts` and
scopes it to the Persian calendar, matching the web example's wrapper. The
component itself keeps application font inheritance and native accessibility.

The two editable-date recipes keep text, selected date, displayed month and
open state in the application. Their parser is application code (`Date` or
`chrono-node` in the originals). `DatePickerInput` exposes the same combination
through explicit `Parse`/`Format` callbacks and a real native editor. Invalid
drafts retain the last accepted date/month. The runnable example demonstrates
a small relative-date grammar; the kit does not embed an NLP engine.

`oracle.json` is derived test data, not an upstream original. It contains 432
conversion/month-length vectors and three faIR label/week/month-clamp cases,
produced by the pinned npm implementation. To reproduce it, verify and extract
the complete npm tarball using the integrity in the manifest, then run:

```sh
TZ=UTC node oracle-generation.mjs /path/to/date-fns-jalali/package > oracle.json
```

The native regressions exercise real editing, marked text, submit, selection,
calendar browsing and focus. They do not establish real Cocoa cancellation,
screen-reader announcements, or native execution on other operating systems.
