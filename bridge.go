package atspi

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/godbus/dbus/v5"
)

const (
	pathPrefix = "/org/a11y/atspi/accessible"
	rootPath   = pathPrefix + "/root"
	nullPath   = "/org/a11y/atspi/null"

	ifaceAccessible  = "org.a11y.atspi.Accessible"
	ifaceApplication = "org.a11y.atspi.Application"
	ifaceComponent   = "org.a11y.atspi.Component"
	ifaceAction      = "org.a11y.atspi.Action"
	ifaceText        = "org.a11y.atspi.Text"
	ifaceProperties  = "org.freedesktop.DBus.Properties"
)

// ref is an object as AT-SPI names it: (bus name, object path).
type ref struct {
	Name string
	Path dbus.ObjectPath
}

// Bridge is an application's connection to the accessibility bus.
type Bridge struct {
	conn    *dbus.Conn
	app     AppInfo
	handler Handler

	mu       sync.RWMutex
	nodes    map[uint64]*Node
	windows  []uint64 // the application's children
	registry ref      // the desktop: the application's parent
	appID    int32
	focused  uint64
	origins  map[uint64][2]int32 // where windows are on the screen, if known
}

// Enabled reports whether assistive technologies are in use (a screen
// reader running, or accessibility switched on in the desktop's settings).
func Enabled() bool {
	conn, err := dbus.SessionBus()
	if err != nil {
		return false
	}
	obj := conn.Object("org.a11y.Bus", "/org/a11y/bus")
	for _, p := range []string{"org.a11y.Status.IsEnabled", "org.a11y.Status.ScreenReaderEnabled"} {
		if v, err := obj.GetProperty(p); err == nil {
			if on, ok := v.Value().(bool); ok && on {
				return true
			}
		}
	}
	return false
}

// Connect joins the accessibility bus as app: its objects appear to screen
// readers once given with Update.
func Connect(app AppInfo, h Handler) (*Bridge, error) {
	session, err := dbus.SessionBus()
	if err != nil {
		return nil, err
	}
	var addr string
	if err := session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		return nil, fmt.Errorf("no accessibility bus: %w", err)
	}
	conn, err := dbus.Dial(addr)
	if err != nil {
		return nil, err
	}
	if err := conn.Auth(nil); err != nil {
		conn.Close()
		return nil, err
	}
	if err := conn.Hello(); err != nil {
		conn.Close()
		return nil, err
	}
	b := &Bridge{conn: conn, app: app, handler: h, nodes: map[uint64]*Node{}, origins: map[uint64][2]int32{}}
	for iface, v := range map[string]any{
		ifaceAccessible:  &accessible{b},
		ifaceApplication: &application{b},
		ifaceComponent:   &component{b},
		ifaceAction:      &action{b},
		ifaceText:        &text{b},
		ifaceProperties:  &properties{b},
	} {
		if err := conn.ExportSubtree(v, pathPrefix, iface); err != nil {
			conn.Close()
			return nil, err
		}
	}
	if err := conn.Export(&cache{b}, cachePath, ifaceCache); err != nil {
		conn.Close()
		return nil, err
	}
	// the registry embeds the application in the desktop
	var parent ref
	err = conn.Object("org.a11y.atspi.Registry", rootPath).Call("org.a11y.atspi.Socket.Embed", 0, b.ref(0)).Store(&parent)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("registering with the accessibility registry: %w", err)
	}
	b.mu.Lock()
	b.registry = parent
	b.mu.Unlock()
	return b, nil
}

// Close leaves the accessibility bus.
func (b *Bridge) Close() error {
	b.conn.Object("org.a11y.atspi.Registry", rootPath).Call("org.a11y.atspi.Socket.Unembed", 0, b.ref(0))
	return b.conn.Close()
}

// BusName is the bridge's name on the accessibility bus.
func (b *Bridge) BusName() string { return b.conn.Names()[0] }

// SetWindowOrigin says where window id is on the screen, for screen
// coordinates (unknown under Wayland: then they are the window's).
func (b *Bridge) SetWindowOrigin(id uint64, x, y int32) {
	b.mu.Lock()
	b.origins[id] = [2]int32{x, y}
	b.mu.Unlock()
}

// Update replaces the objects with nodes (all of them: those not given are
// gone) and tells assistive technologies what changed.
func (b *Bridge) Update(nodes []Node) {
	next := make(map[uint64]*Node, len(nodes))
	var windows []uint64
	for i := range nodes {
		n := &nodes[i]
		next[n.ID] = n
		if n.Parent == 0 {
			windows = append(windows, n.ID)
		}
	}
	b.mu.Lock()
	prev, prevWindows := b.nodes, b.windows
	b.nodes, b.windows = next, windows
	b.mu.Unlock()

	b.childrenChanged(0, prevWindows, windows)
	for id, n := range next {
		old, ok := prev[id]
		if !ok {
			continue
		}
		b.childrenChanged(id, old.Children, n.Children)
		if old.Name != n.Name {
			b.emit(id, "Object", "PropertyChange", "accessible-name", 0, 0, n.Name)
		}
		if old.Description != n.Description {
			b.emit(id, "Object", "PropertyChange", "accessible-description", 0, 0, n.Description)
		}
		for _, s := range []State{StateChecked, StateExpanded, StateSelected, StateEnabled, StateSensitive} {
			if was, is := slices.Contains(old.States, s), slices.Contains(n.States, s); was != is {
				b.emit(id, "Object", "StateChanged", stateName(s), boolInt(is), 0, int32(0))
			}
		}
	}
}

// childrenChanged reports the children of parent added and removed.
func (b *Bridge) childrenChanged(parent uint64, old, now []uint64) {
	for i, c := range old {
		if !slices.Contains(now, c) {
			b.emit(parent, "Object", "ChildrenChanged", "remove", int32(i), 0, b.ref(c))
		}
	}
	for i, c := range now {
		if !slices.Contains(old, c) {
			b.emit(parent, "Object", "ChildrenChanged", "add", int32(i), 0, b.ref(c))
		}
	}
}

// Focused says that object id has the keyboard focus (0: nothing has).
// Screen readers speak what gets the focus: without this they are silent.
func (b *Bridge) Focused(id uint64) {
	b.mu.Lock()
	prev := b.focused
	b.focused = id
	b.mu.Unlock()
	if prev == id {
		return
	}
	if prev != 0 {
		b.emit(prev, "Object", "StateChanged", "focused", 0, 0, int32(0))
	}
	if id != 0 {
		b.emit(id, "Object", "StateChanged", "focused", 1, 0, int32(0))
		b.emit(id, "Focus", "Focus", "", 0, 0, int32(0))
	}
}

// WindowActivated says that window id became the active window (or no
// longer is). Screen readers follow the active window.
func (b *Bridge) WindowActivated(id uint64, active bool) {
	name := ""
	b.mu.RLock()
	if n := b.nodes[id]; n != nil {
		name = n.Name
	}
	b.mu.RUnlock()
	b.emit(id, "Object", "StateChanged", "active", boolInt(active), 0, int32(0))
	if active {
		b.emit(id, "Window", "Activate", "", 0, 0, name)
	} else {
		b.emit(id, "Window", "Deactivate", "", 0, 0, name)
	}
}

// TextInserted says that text was inserted at offset in object id; give
// the updated node (Update) first. Screen readers echo typing from this.
func (b *Bridge) TextInserted(id uint64, offset int, inserted string) {
	b.emit(id, "Object", "TextChanged", "insert", int32(offset), int32(len([]rune(inserted))), inserted)
}

// TextDeleted says that the text deleted was removed at offset.
func (b *Bridge) TextDeleted(id uint64, offset int, deleted string) {
	b.emit(id, "Object", "TextChanged", "delete", int32(offset), int32(len([]rune(deleted))), deleted)
}

// CaretMoved says that the caret of object id is at offset now (after a
// TextInserted or TextDeleted, if text changed).
func (b *Bridge) CaretMoved(id uint64, offset int) {
	b.emit(id, "Object", "TextCaretMoved", "", int32(offset), 0, int32(0))
}

// SelectionChanged says that the text selected in object id changed.
func (b *Bridge) SelectionChanged(id uint64) {
	b.emit(id, "Object", "TextSelectionChanged", "", 0, 0, int32(0))
}

// Announce asks screen readers to say message (a status: "Saved").
func (b *Bridge) Announce(id uint64, message string) {
	b.emit(id, "Object", "Announcement", message, 0, 0, message) // politeness: 0 (as the toolkit chooses)
}

func (b *Bridge) emit(id uint64, iface, member, kind string, d1, d2 int32, data any) {
	_ = b.conn.Emit(b.path(id), "org.a11y.atspi.Event."+iface+"."+member, kind, d1, d2, dbus.MakeVariant(data), map[string]dbus.Variant{})
}

func boolInt(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

func stateName(s State) string {
	switch s {
	case StateChecked:
		return "checked"
	case StateExpanded:
		return "expanded"
	case StateSelected:
		return "selected"
	case StateEnabled:
		return "enabled"
	case StateSensitive:
		return "sensitive"
	}
	return ""
}

func (b *Bridge) path(id uint64) dbus.ObjectPath {
	if id == 0 {
		return rootPath
	}
	return dbus.ObjectPath(pathPrefix + "/" + strconv.FormatUint(id, 10))
}

func (b *Bridge) ref(id uint64) ref { return ref{b.BusName(), b.path(id)} }

var errUnknown = dbus.NewError("org.freedesktop.DBus.Error.UnknownObject", []any{"no such object"})

// lookup is the object a message is for (nil, ok for the application).
func (b *Bridge) lookup(m dbus.Message) (uint64, *Node, *dbus.Error) {
	p, _ := m.Headers[dbus.FieldPath].Value().(dbus.ObjectPath)
	last := string(p)[strings.LastIndex(string(p), "/")+1:]
	if last == "root" {
		return 0, nil, nil
	}
	id, err := strconv.ParseUint(last, 10, 64)
	if err != nil {
		return 0, nil, errUnknown
	}
	b.mu.RLock()
	n := b.nodes[id]
	b.mu.RUnlock()
	if n == nil {
		return 0, nil, errUnknown
	}
	return id, n, nil
}

// children are the children of object id.
func (b *Bridge) children(id uint64, n *Node) []uint64 {
	if id == 0 {
		b.mu.RLock()
		defer b.mu.RUnlock()
		return slices.Clone(b.windows)
	}
	return n.Children
}

func (b *Bridge) parentRef(id uint64, n *Node) ref {
	if id == 0 {
		b.mu.RLock()
		defer b.mu.RUnlock()
		return b.registry
	}
	return b.ref(n.Parent)
}

// screen is where a rectangle of window w is on the screen (coordType 0)
// or in the window (1, 2).
func (b *Bridge) screen(r Rect, window uint64, coordType uint32) Rect {
	if coordType == 0 {
		b.mu.RLock()
		o := b.origins[window]
		b.mu.RUnlock()
		r.X += o[0]
		r.Y += o[1]
	}
	return r
}

// windowOf is the window an object is in.
func (b *Bridge) windowOf(id uint64) uint64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for i := 0; i < 1000; i++ {
		n := b.nodes[id]
		if n == nil || n.Parent == 0 {
			return id
		}
		id = n.Parent
	}
	return id
}

var errNoText = errors.New("not text")
