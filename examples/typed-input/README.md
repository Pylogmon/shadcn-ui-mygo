# Typed input examples

Run `go run ./examples/typed-input`. Choose Input Types, Basic and Field states, With controls, or Form. Date/time use the user's native regional formats. File/Picture opens the system picker after activation. The form validates email; URL validation appears in Input Types.

Use `-section types|states|composition|form`, `-dark`, or `-snapshot /path/input.png` for a reproducible native render.

NumberInput separates editable text from committed `float64` state, with optional Min/Max, Step and typed callbacks. DateInput and TimeInput bind `time.Time`; FileInput binds a path slice. Field and InputGroup own labels, descriptions and errors. TimeInput's Handle and error relations target its native group; internal segment focus remains native. Email/tel/URL remain string inputs with application validation. See the fixed source manifest and native boundary notes in `pkg/control/testdata/shadcn-input`.
