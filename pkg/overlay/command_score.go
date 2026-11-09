// Copyright (c) 2022 Paco Coursey
// SPDX-License-Identifier: MIT
// The scoring rules are ported from cmdk 1.1.1's command-score.ts (MIT).
// Pinned reference, license and offline differential fixtures: testdata/cmdk.
package overlay

import (
	"math"
	"strings"
	"unicode/utf16"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CommandScore returns cmdk's fuzzy match score in [0, 1]. Zero hides a result;
// larger scores prefer exact, consecutive and word-boundary matches, with small
// case/distance penalties and support for adjacent transpositions/repeated keys.
// Keywords are appended as aliases, as in the upstream command-score function.
// Command uses the displayed Label and includes a distinct Value as an alias.
func CommandScore(value, query string, keywords ...string) float64 {
	if len(keywords) > 0 {
		value += " " + strings.Join(keywords, " ")
	}
	original, abbreviation := commandUnits(value), commandUnits(query)
	if len(abbreviation) == 0 {
		if len(original) == 0 {
			return 1
		}
		return .99
	}
	lower, lowerQuery := commandFoldUnits(value), commandFoldUnits(query)
	type position struct{ value, query int }
	memo := make(map[position]float64)
	var score func(int, int) float64
	score = func(start, qi int) float64 {
		// Case expansion can let the upstream transposition branch overshoot
		// the original UTF-16 query length and recurse forever. Treat a
		// completed/overshot query as terminal instead (see oracle failures).
		if qi >= len(abbreviation) {
			if start == len(original) {
				return 1
			}
			return .99
		}
		key := position{start, qi}
		if result, found := memo[key]; found {
			return result
		}
		high := 0.0
		needle := commandUnitAt(lowerQuery, qi)
		for i := start; i < len(lower); i++ {
			if int32(lower[i]) != needle {
				continue
			}
			result := score(i+1, qi+1)
			if result > high {
				switch {
				case i == start:
				case commandGap(commandUnitAt(original, i-1)):
					result *= .8
					if start > 0 {
						result *= math.Pow(.999, float64(commandBreaks(original, start, i-1, commandGap)))
					}
				case commandSpace(commandUnitAt(original, i-1)):
					result *= .9
					if start > 0 {
						result *= math.Pow(.999, float64(commandBreaks(original, start, i-1, commandSpace)))
					}
				default:
					result *= .17
					if start > 0 {
						result *= math.Pow(.999, float64(i-start))
					}
				}
				if commandUnitAt(original, i) != commandUnitAt(abbreviation, qi) {
					result *= .9999
				}
			}
			previous, next := commandUnitAt(lower, i-1), commandUnitAt(lowerQuery, qi+1)
			if result < .1 && previous == next || next == needle && previous != needle {
				result = max(result, score(i+1, qi+2)*.1)
			}
			high = max(high, result)
		}
		memo[key] = high
		return high
	}
	return score(0, 0)
}

// JavaScript's algorithm indexes UTF-16 units, not bytes or Unicode scalars.
// Preserve that convention for supplementary characters and case expansion.
func commandUnits(value string) []uint16 { return utf16.Encode([]rune(value)) }

func commandFoldUnits(value string) []uint16 {
	// ASCII commands are the common path and do not need a Unicode transformer.
	ascii := true
	for i := range len(value) {
		if value[i] >= 0x80 {
			ascii = false
			break
		}
	}
	var units []uint16
	if ascii {
		units = make([]uint16, len(value))
		for i := range value {
			ch := value[i]
			if ch >= 'A' && ch <= 'Z' {
				ch += 'a' - 'A'
			}
			units[i] = uint16(ch)
		}
	} else {
		// Full, language-independent lower casing includes dotted I expansion
		// and contextual final sigma; strings.ToLower does neither.
		units = commandUnits(cases.Lower(language.Und).String(value))
	}
	for i, unit := range units {
		if commandSpace(int32(unit)) {
			units[i] = ' '
		}
	}
	return units
}

func commandUnitAt(units []uint16, index int) int32 {
	if index < 0 || index >= len(units) {
		return -1
	}
	return int32(units[index])
}

func commandGap(unit int32) bool {
	switch unit {
	case '\\', '/', '_', '+', '.', '#', '"', '@', '[', '(', '{', '&':
		return true
	}
	return false
}

// ECMAScript \s includes BOM and excludes NEL, unlike unicode.IsSpace.
func commandSpace(unit int32) bool {
	switch unit {
	case '-', ' ', '\t', '\n', '\v', '\f', '\r', 0xA0, 0x1680, 0x2028, 0x2029, 0x202F, 0x205F, 0x3000, 0xFEFF:
		return true
	}
	return unit >= 0x2000 && unit <= 0x200A
}

func commandBreaks(units []uint16, start, end int, isBreak func(int32) bool) int {
	count := 0
	for _, unit := range units[min(start, len(units)):max(min(start, len(units)), min(end, len(units)))] {
		if isBreak(int32(unit)) {
			count++
		}
	}
	return count
}

func commandDefaultScore(item CommandItem, query string) float64 {
	aliases := item.Keywords
	if item.Label != item.Value {
		aliases = make([]string, 0, len(item.Keywords)+1)
		aliases = append(aliases, item.Value)
		aliases = append(aliases, item.Keywords...)
	}
	return CommandScore(item.Label, query, aliases...)
}
