# Input OTP source compositions

Gallery's Forms & controls page exposes **Input OTP configuration** with all
10 documentation examples and eight Create sections at shadcn/ui revision
`8ab27715521cbd4f87ae6886cd989cbb678f6de7`. The 12 original files, byte lengths,
SHA-256 values, Git blobs and pinned URLs are recorded in
[`source-map.json`](source-map.json).

| Configuration | Original material retained |
| --- | --- |
| Six populated slots | Docs default `123456`, six contiguous slots |
| Digits Only | Original Field label, fit-content width and ASCII digit pattern |
| Three pairs | Three empty pairs with two separators |
| Disabled populated code | Disabled native editor, `123456`, groups of three |
| Controlled value | Empty draft and exact “Enter your one-time password.” / “You entered: …” readout |
| Invalid three pairs | `000000`, invalid slots and editable native input |
| Four digits | Empty four-slot digit pattern |
| Alphanumeric code | Empty groups of three and ASCII letters/numbers pattern |
| Verify your login (docs) | Full card, email emphasis, 44×48 slots, wider separators, required input, Resend/Verify and both links |
| Translated verification code | Verbatim English/Arabic/Hebrew labels, Arabic default and populated six slots |
| Create: Verify your login | Full card, explicit Vega 48×64 slots and eight-DIP footer gap |
| Create: Simple | Simple Field label and empty groups of three |
| Create: Digits Only | Digits Only Field and six empty contiguous slots |
| Create: With Separator | With Separator label and `123456` in three pairs |
| Create: Alphanumeric | Original preceding “Accepts both letters and numbers.” description and ASCII pattern |
| Create: Disabled | Ordinary enabled label with disabled populated native editor |
| Create: 4 Digits | Original label and preceding PIN description |
| Create: Invalid State | Original normal-color label, preceding description, `000000` three pairs and explicit “Invalid code. Please try again.” error |

Every code uses the public `control.InputOTP`: one actual native editor owns
text, selection, replacement, clipboard, undo and IME. The visible slot and
separator parts remain inert and add no keyboard/AX targets. The source patterns
gate the entire candidate; the native scrubbing presets are not substituted.
In particular, the existing legacy `Alphanumeric` preset accepts Unicode
letters, while these source examples explicitly accept only ASCII letters and
digits. The unrestricted defaults remain unrestricted.

The ordinary slots use Vega's 36×36 DIPs. The docs form explicitly uses 44×48;
its separator's eight-DIP horizontal margins combine with the ordinary gap to
make 48 DIPs between groups. Create's style-specific form classes explicitly
use 48×64 in Vega. Native `SlotContent` retains the source 20-DIP form text.
The form card uses the source 448-DIP maximum; fixed slots retain their sizes
in a real horizontal native viewport when the form is narrower. Other
unspecified Create wrapper widths use a 320-DIP native preview maximum.

Neither original form has a validation or transport handler. The actual Verify
button records a local application snapshot; Resend Code and the two real
native links record local intent. No extra six-character validation, implicit
Enter submit, email delivery or account change is implemented. Required is
associated with the actual editor without inventing a visible asterisk. The
original `href="#"` links run through a public local Router: native pointer and
Enter activate once, Space does not, and no OS browser opens. MyGo consumes the
path/query and drops the fragment; HTML anchor scrolling/history is not claimed.

The configuration and language selectors follow Gallery's Flyout/Native mode.
Each recipe owns an independent string, selection integers, change counters and
local snapshots. No Context, Element, FieldScope or frame callback survives a
view. No font, icon font, image asset or dependency is added; the existing
on-demand Material SVG refresh symbol supplies the source Resend icon.

`TestOTPCompositionsSourceWorkflow` is one concentrated public-input flow:
actual configuration selection, native select-all/replacement/paste/undo,
anchored patterns, four-slot bounds, controlled readout, disabled/invalid,
Field associations, full form dimensions/actions/links, Hebrew/narrow layout,
page return and style changes. Set `GALLERY_OTP_COMPOSITION_SNAPSHOT_DIR` to
export four representative software PNGs.

Software images and injected selector windows do not prove real OS/AT/IME,
font, multi-display or full platform/visual acceptance. Native indices use
runes rather than browser UTF-16 units. The existing public SDK does not place
an IME candidate window at each painted slot, and its empty-clipboard-to-text
paste-transform transaction cannot guarantee one-step native undo. These
examples do not expand or claim to resolve those documented SDK boundaries.
