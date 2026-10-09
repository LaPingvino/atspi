package atspi

import (
	"sort"
	"unicode"
)

// Units of text (AtspiTextGranularity), for GetStringAtOffset.
const (
	granularityChar uint32 = iota
	granularityWord
	granularitySentence
	granularityLine
	granularityParagraph
)

// Boundaries of text (AtspiTextBoundaryType), for GetTextAtOffset and the
// text before and after an offset.
const (
	boundaryChar uint32 = iota
	boundaryWordStart
	boundaryWordEnd
	boundarySentenceStart
	boundarySentenceEnd
	boundaryLineStart
	boundaryLineEnd
)

// textOf is a Text's characters with what a screen reader asks of them.
type textOf struct {
	r []rune
	t *Text
}

func newTextOf(t *Text) textOf { return textOf{r: []rune(t.Content), t: t} }

func (x textOf) clamp(off int) int { return min(max(off, 0), len(x.r)) }

// slice is the text of [start, end), clamped.
func (x textOf) slice(start, end int) string {
	start, end = x.clamp(start), x.clamp(end)
	if end < start {
		return ""
	}
	return string(x.r[start:end])
}

func isWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’' }

// boundaries are the offsets at which units of a kind start, from 0 to
// the end (both included), sorted.
func (x textOf) boundaries(kind uint32) []int {
	n := len(x.r)
	b := []int{0}
	add := func(i int) {
		if i > b[len(b)-1] && i < n {
			b = append(b, i)
		}
	}
	switch kind {
	case boundaryChar:
		for i := 1; i < n; i++ {
			add(i)
		}
	case boundaryWordStart:
		for i := 1; i < n; i++ {
			if isWord(x.r[i]) && !isWord(x.r[i-1]) {
				add(i)
			}
		}
	case boundaryWordEnd:
		for i := 1; i < n; i++ {
			if !isWord(x.r[i]) && isWord(x.r[i-1]) {
				add(i)
			}
		}
	case boundarySentenceStart, boundarySentenceEnd:
		for i := 1; i < n; i++ {
			end := (x.r[i-1] == '.' || x.r[i-1] == '!' || x.r[i-1] == '?' || x.r[i-1] == '\n') && unicode.IsSpace(x.r[i])
			if kind == boundarySentenceEnd && end {
				add(i)
			}
			if kind == boundarySentenceStart && !unicode.IsSpace(x.r[i]) {
				j := i - 1
				for j > 0 && unicode.IsSpace(x.r[j]) && x.r[j] != '\n' {
					j--
				}
				if x.r[j] == '.' || x.r[j] == '!' || x.r[j] == '?' || x.r[j] == '\n' || (x.r[i-1] == '\n') {
					add(i)
				}
			}
		}
	case boundaryLineStart:
		if x.t.Lines != nil {
			for _, l := range x.t.Lines {
				add(l)
			}
		} else {
			for i := 1; i < n; i++ {
				if x.r[i-1] == '\n' {
					add(i)
				}
			}
		}
	case boundaryLineEnd:
		if x.t.Lines != nil {
			for _, l := range x.t.Lines[min(1, len(x.t.Lines)):] {
				// a wrapped line ends where the next starts; one ended by a
				// line break, at the break
				if l > 0 && l <= n && x.r[l-1] == '\n' {
					add(l - 1)
				} else {
					add(l)
				}
			}
		} else {
			for i := 0; i < n; i++ {
				if x.r[i] == '\n' {
					add(i)
				}
			}
		}
	}
	return append(b, n)
}

// unit is the unit [start, end) of boundary kind around off: n units
// before it (n < 0), the one at it (0) or after it (n > 0).
func (x textOf) unit(kind uint32, off, n int) (int, int) {
	b := x.boundaries(kind)
	off = x.clamp(off)
	i := sort.Search(len(b), func(i int) bool { return b[i] > off }) - 1 // b[i] <= off
	i = min(max(i, 0), len(b)-2)
	i += n
	if i < 0 || i >= len(b)-1 {
		return off, off
	}
	return b[i], b[i+1]
}

// atOffset is GetTextAtOffset (n = 0) and the text before (-1) and after
// (1) an offset.
func (x textOf) atOffset(off int, kind uint32, n int) (string, int32, int32) {
	if len(x.r) == 0 {
		return "", 0, 0
	}
	if kind > boundaryLineEnd {
		kind = boundaryChar
	}
	start, end := x.unit(kind, off, n)
	return x.slice(start, end), int32(start), int32(end)
}

// stringAt is GetStringAtOffset: the character, word, sentence, line or
// paragraph at off, without the spaces or line break around it.
func (x textOf) stringAt(off int, granularity uint32) (string, int32, int32) {
	if len(x.r) == 0 {
		return "", 0, 0
	}
	var start, end int
	switch granularity {
	case granularityWord:
		start, end = x.unit(boundaryWordStart, off, 0)
		for end > start && !isWord(x.r[end-1]) {
			end--
		}
	case granularitySentence:
		start, end = x.unit(boundarySentenceStart, off, 0)
	case granularityLine:
		start, end = x.unit(boundaryLineStart, off, 0)
	case granularityParagraph:
		start, end = textOf{r: x.r, t: &Text{}}.unit(boundaryLineStart, off, 0)
	default:
		start, end = x.unit(boundaryChar, off, 0)
		return x.slice(start, end), int32(start), int32(end)
	}
	for end > start && (x.r[end-1] == '\n' || (granularity != granularityWord && unicode.IsSpace(x.r[end-1]))) {
		end--
	}
	return x.slice(start, end), int32(start), int32(end)
}
