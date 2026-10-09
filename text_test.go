package atspi

import "testing"

// Words, lines (wrapped ones too) and sentences as a screen reader asks
// for them, in characters (é is one).
func TestTextUnits(t *testing.T) {
	x := newTextOf(&Text{Content: "Café au lait.\nJosé sits. He waits."})
	for _, c := range []struct {
		off         int
		granularity uint32
		want        string
		start, end  int32
	}{
		{1, granularityChar, "a", 1, 2},
		{2, granularityWord, "Café", 0, 4},
		{15, granularityWord, "José", 14, 18},
		{16, granularityLine, "José sits. He waits.", 14, 34},
		{3, granularityLine, "Café au lait.", 0, 13},
		{25, granularitySentence, "He waits.", 25, 34},
	} {
		s, a, b := x.stringAt(c.off, c.granularity)
		if s != c.want || a != c.start || b != c.end {
			t.Errorf("%d/%d: %q %d-%d, want %q %d-%d", c.off, c.granularity, s, a, b, c.want, c.start, c.end)
		}
	}
	// a wrapped line ends where the next starts
	w := newTextOf(&Text{Content: "one two three four", Lines: []int{0, 8, 14}})
	if s, _, _ := w.stringAt(10, granularityLine); s != "three" {
		t.Errorf("wrapped line: %q", s)
	}
	if s, _, _ := w.atOffset(10, boundaryLineStart, -1); s != "one two " {
		t.Errorf("line before: %q", s)
	}
	if s, _, _ := w.atOffset(10, boundaryLineStart, 1); s != "four" {
		t.Errorf("line after: %q", s)
	}
	if s, _, _ := newTextOf(&Text{}).stringAt(0, granularityLine); s != "" {
		t.Errorf("empty: %q", s)
	}
}
