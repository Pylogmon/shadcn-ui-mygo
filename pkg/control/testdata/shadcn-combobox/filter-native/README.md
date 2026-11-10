# Native source filter

The already archived, unchanged Base UI 1.6.0 `internals/filter.js` uses
`Intl.Collator` with search usage, base sensitivity and ignored punctuation.
It compares windows of exactly `query.length` UTF-16 code units. The archived
`AriaCombobox.js` first applies ECMAScript `String.trim` to the default query.
`cases.json` records both original SHA-256 values and 64 reference vectors.

Run `node oracle.cjs` in this directory, using the recorded Node v26.12.0,
ICU 78.3 / CLDR 48 runtime, to reproduce the JSON. The script executes the
unmodified official filter in a VM. Its two import adapters provide String
identity for the plain string labels used here and locale-to-cache-key text;
they replace neither the source matching algorithm nor `Intl.Collator`.
The Go test needs only the committed JSON, not a JavaScript installation.

The native implementation retains the pinned `golang.org/x/text/search`
search collator, comparing each fixed-length window as a whole. It removes
punctuation and spacing before comparison, retaining symbols such as `+`
and `$`. A generic sort collator is not equivalent: German search `ä/a`
and Persian `ک/ك` differ even at base sensitivity. UTF-16 windows that split
a surrogate pair cannot equal a valid Go query and are excluded rather than
decoded into U+FFFD. The real replacement character remains matchable.
Literal-empty and collation-empty windows compare through the public empty
search pattern, including combining marks and default-ignorable characters.
An ignored prefix cannot hide a later non-ignorable letter or symbol.

The vectors cover canonical-length asymmetry, ligatures/expansions, Turkish
and Azerbaijani casing, Swedish/German/Persian tailoring, Greek/Russian/Indic
text, kana/width/voicing, punctuation versus symbols, ECMAScript BOM/NEL trim
and supplementary-character boundaries. One native flow exercises typed
search, keyboard selection, locale replacement and blank multi-query policy.
Custom Filter callbacks still receive the original query and the actual
editor keeps its original text.

The pinned Go search tables derive from CLDR 23 / Unicode 6.2; the reference
Intl runtime uses newer data. These verified cases and the source window/trim
algorithm do not prove every locale or newer Unicode collation identical to
all browser/OS versions. The public [search API](https://pkg.go.dev/golang.org/x/text@v0.43.0/search)
and the fixed dependency's `search/search.go` document its search behavior.
Actual IME candidate windows and screen-reader acceptance remain separate.

Original files and their MIT notice remain in the existing parent archive;
this folder adds an oracle and native adaptation record, not another npm
runtime dependency or a modified official source file.
