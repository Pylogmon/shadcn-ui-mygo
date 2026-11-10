package control

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

// comboLabelFilter preserves Base UI's fixed UTF-16 window before comparison.
// A variable-length substring search would also match expansions such as ß/ss
// and composed/decomposed forms that the source's slice boundaries exclude.
type comboLabelFilter struct {
	width      int
	pattern    *search.Pattern
	zero       *search.Pattern
	emptyQuery bool
}

func newComboLabelFilter(query, locale string) comboLabelFilter {
	matcher := search.New(language.Make(locale), search.Loose)
	text := comboFilterText(query)
	zero := matcher.CompileString("")
	return comboLabelFilter{
		width:      len(utf16.Encode([]rune(query))),
		pattern:    matcher.CompileString(text),
		zero:       zero,
		emptyQuery: comboCollationEmpty(zero, text),
	}
}

func (f comboLabelFilter) contains(label string) bool {
	if f.width == 0 {
		return true
	}
	units := utf16.Encode([]rune(label))
	for start := 0; start+f.width <= len(units); start++ {
		end := start + f.width
		// JavaScript slices may contain an isolated surrogate. It cannot equal
		// a valid Go query; do not turn it into U+FFFD during UTF-16 decoding.
		if units[start] >= 0xDC00 && units[start] <= 0xDFFF || units[end-1] >= 0xD800 && units[end-1] <= 0xDBFF {
			continue
		}
		window := comboFilterText(string(utf16.Decode(units[start:end])))
		if f.emptyQuery {
			if comboCollationEmpty(f.zero, window) {
				return true
			}
			continue
		}
		begin, matchedEnd := f.pattern.IndexString(window, search.Anchor)
		if begin == 0 && matchedEnd == len(window) {
			return true
		}
	}
	return false
}

func comboCollationEmpty(zero *search.Pattern, text string) bool {
	// The search API does not report an empty input as an anchored match.
	// Nonempty inputs may also consist entirely of zero-weight collation
	// elements (for example ZWJ or a combining accent at base sensitivity).
	// Consume only prefixes accepted by the public empty search pattern; a
	// later non-ignorable element must still make the comparison unequal.
	for text != "" {
		begin, end := zero.IndexString(text, search.Anchor)
		if begin != 0 || end <= 0 {
			return false
		}
		text = text[end:]
	}
	return true
}

func comboFilterText(text string) string {
	// Intl ignorePunctuation ignores punctuation and spacing, retaining
	// symbols such as + and $. The existing search collator provides locale
	// search tailoring, case/diacritic/width equivalence and default ignorables.
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) || unicode.IsSpace(r) {
			return -1
		}
		return r
	}, text)
}

// ECMAScript String.trim includes FEFF but excludes the NEL control U+0085.
// Keep the editor's raw query; use this only for the source filtering policy.
func comboTrimQuery(query string) string {
	return strings.TrimFunc(query, func(r rune) bool {
		switch r {
		case '\t', '\n', '\v', '\f', '\r', '\u2028', '\u2029', '\uFEFF':
			return true
		}
		return unicode.Is(unicode.Zs, r)
	})
}
