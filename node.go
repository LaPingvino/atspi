// Package atspi makes a Go application's user interface visible to Linux
// screen readers (Orca) and other assistive technologies, over AT-SPI2: the
// accessibility bus of the freedesktop desktops.
//
// It knows nothing of any GUI toolkit. A toolkit (or an application) hands
// it a snapshot of its accessible objects, as a list of Nodes, whenever they
// change, and tells it what happens: the focus moving, a window becoming
// active, text being typed. The Bridge answers the screen reader's questions
// from the latest snapshot, on its own goroutines, without waiting for the
// toolkit's UI thread; requests that change something (an action, moving
// the caret) are passed back through the Handler.
//
// Text offsets are in characters (runes), as AT-SPI counts them.
package atspi

// Node is one accessible object at one moment.
type Node struct {
	// ID identifies the object for as long as it exists: the same object
	// keeps its ID from one snapshot to the next, and an ID is not reused
	// for another object (a screen reader holds on to them). 0 is the
	// application itself; IDs given here must not be 0.
	ID uint64
	// Parent is the ID of the object containing this one; 0 for a window
	// (its parent is the application).
	Parent   uint64
	Children []uint64

	Role        Role
	Name        string
	Description string
	States      []State

	// Bounds is where the object is, in pixels of its window (0, 0 the
	// window's top left).
	Bounds Rect

	// Actions are the names of what the object can do ("click", "press");
	// a screen reader asks for one by index (Handler.DoAction).
	Actions []string

	// Text, if set, is the object's text, which a screen reader reads by
	// character, word and line: a label's, an entry's, an editor's.
	Text *Text
}

// Rect is a rectangle in pixels.
type Rect struct {
	X, Y, Width, Height int32
}

// Text is an object's text, with the caret and the selection, in runes.
type Text struct {
	Content string
	// Caret is where the caret is (-1: the object has none).
	Caret int
	// SelectionStart and SelectionEnd are the selected characters
	// [start, end); equal when nothing is selected.
	SelectionStart, SelectionEnd int
	// Lines are the offsets at which the lines start as they are shown
	// (wrapped lines included), the first being 0. Nil: lines end at
	// line breaks.
	Lines []int
	// CharacterBounds, if set, is where a character is drawn, in pixels
	// of the window (for magnifiers and for the caret's position).
	CharacterBounds func(offset int) Rect
}

// AppInfo describes the application.
type AppInfo struct {
	Name           string // the application's name, as a screen reader says it
	ToolkitName    string // "Fyne"
	ToolkitVersion string
}

// Handler receives the requests of assistive technologies that change
// something. Its methods are called on the bridge's goroutines: hand the
// work to the UI thread. Any of them may be nil.
type Handler struct {
	// DoAction performs action index of Node.Actions on object id.
	DoAction func(id uint64, index int) bool
	// GrabFocus moves the keyboard focus to object id.
	GrabFocus func(id uint64) bool
	// SetCaret moves the caret of object id to offset.
	SetCaret func(id uint64, offset int) bool
	// SetSelection selects [start, end) in object id.
	SetSelection func(id uint64, start, end int) bool
}
