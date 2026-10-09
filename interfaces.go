package atspi

import (
	"os"
	"slices"
	"strconv"

	"github.com/godbus/dbus/v5"
)

// The D-Bus interfaces, each serving every object under pathPrefix (the
// message's path says which).

// accessible is org.a11y.atspi.Accessible.
type accessible struct{ b *Bridge }

func (a *accessible) GetChildAtIndex(m dbus.Message, i int32) (ref, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil {
		return ref{}, err
	}
	c := a.b.children(id, n)
	if i < 0 || int(i) >= len(c) {
		return ref{a.b.BusName(), nullPath}, nil
	}
	return a.b.ref(c[i]), nil
}

func (a *accessible) GetChildren(m dbus.Message) ([]ref, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil {
		return nil, err
	}
	var out []ref
	for _, c := range a.b.children(id, n) {
		out = append(out, a.b.ref(c))
	}
	return out, nil
}

func (a *accessible) GetIndexInParent(m dbus.Message) (int32, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil || id == 0 {
		return -1, err
	}
	var siblings []uint64
	if n.Parent == 0 {
		siblings = a.b.children(0, nil)
	} else {
		a.b.mu.RLock()
		p := a.b.nodes[n.Parent]
		a.b.mu.RUnlock()
		if p != nil {
			siblings = p.Children
		}
	}
	return int32(slices.Index(siblings, id)), nil
}

// relation is a relation to other objects (none yet).
type relation struct {
	Type    uint32
	Targets []ref
}

func (a *accessible) GetRelationSet(m dbus.Message) ([]relation, *dbus.Error) {
	if _, _, err := a.b.lookup(m); err != nil {
		return nil, err
	}
	return []relation{}, nil
}

func (a *accessible) GetRole(m dbus.Message) (uint32, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil {
		return 0, err
	}
	if id == 0 {
		return uint32(RoleApplication), nil
	}
	return uint32(n.Role), nil
}

func (a *accessible) GetRoleName(m dbus.Message) (string, *dbus.Error) {
	r, err := a.GetRole(m)
	return roleNames[Role(r)], err
}

func (a *accessible) GetLocalizedRoleName(m dbus.Message) (string, *dbus.Error) {
	return a.GetRoleName(m)
}

func (a *accessible) GetState(m dbus.Message) ([]uint32, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil {
		return nil, err
	}
	bits := []uint32{0, 0}
	if id != 0 {
		for _, s := range n.States {
			bits[s/32] |= 1 << (s % 32)
		}
		a.b.mu.RLock()
		focused := a.b.focused == id
		a.b.mu.RUnlock()
		if focused {
			bits[StateFocused/32] |= 1 << (StateFocused % 32)
		}
	}
	return bits, nil
}

func (a *accessible) GetAttributes(m dbus.Message) (map[string]string, *dbus.Error) {
	id, _, err := a.b.lookup(m)
	if err != nil {
		return nil, err
	}
	if id == 0 {
		return map[string]string{"toolkit": a.b.app.ToolkitName}, nil
	}
	return map[string]string{}, nil
}

func (a *accessible) GetApplication(m dbus.Message) (ref, *dbus.Error) {
	if _, _, err := a.b.lookup(m); err != nil {
		return ref{}, err
	}
	return a.b.ref(0), nil
}

func (a *accessible) GetInterfaces(m dbus.Message) ([]string, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil {
		return nil, err
	}
	return interfacesOf(id, n), nil
}

func interfacesOf(id uint64, n *Node) []string {
	if id == 0 {
		return []string{ifaceAccessible, ifaceApplication}
	}
	out := []string{ifaceAccessible, ifaceComponent}
	if len(n.Actions) > 0 {
		out = append(out, ifaceAction)
	}
	if n.Text != nil {
		out = append(out, ifaceText)
	}
	return out
}

// application is org.a11y.atspi.Application (the root only).
type application struct{ b *Bridge }

func (a *application) GetLocale(m dbus.Message, lctype uint32) (string, *dbus.Error) {
	for _, v := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if l := os.Getenv(v); l != "" {
			return l, nil
		}
	}
	return "C", nil
}

func (a *application) GetApplicationBusAddress(m dbus.Message) (string, *dbus.Error) {
	return "", nil // no peer-to-peer bus
}

// component is org.a11y.atspi.Component: where objects are.
type component struct{ b *Bridge }

type extents struct{ X, Y, Width, Height int32 }

func (c *component) bounds(m dbus.Message, coordType uint32) (Rect, uint64, *dbus.Error) {
	id, n, err := c.b.lookup(m)
	if err != nil || id == 0 {
		return Rect{}, id, err
	}
	return c.b.screen(n.Bounds, c.b.windowOf(id), coordType), id, nil
}

func (c *component) Contains(m dbus.Message, x, y int32, coordType uint32) (bool, *dbus.Error) {
	r, _, err := c.bounds(m, coordType)
	return x >= r.X && y >= r.Y && x < r.X+r.Width && y < r.Y+r.Height, err
}

func (c *component) GetAccessibleAtPoint(m dbus.Message, x, y int32, coordType uint32) (ref, *dbus.Error) {
	id, n, err := c.b.lookup(m)
	if err != nil {
		return ref{}, err
	}
	// the deepest child containing the point
	window := c.b.windowOf(id)
	found := uint64(0)
	var walk func(ids []uint64)
	walk = func(ids []uint64) {
		for _, cid := range ids {
			c.b.mu.RLock()
			cn := c.b.nodes[cid]
			c.b.mu.RUnlock()
			if cn == nil {
				continue
			}
			r := c.b.screen(cn.Bounds, window, coordType)
			if x >= r.X && y >= r.Y && x < r.X+r.Width && y < r.Y+r.Height {
				found = cid
				walk(cn.Children)
				return
			}
		}
	}
	walk(c.b.children(id, n))
	if found == 0 {
		return ref{c.b.BusName(), nullPath}, nil
	}
	return c.b.ref(found), nil
}

func (c *component) GetExtents(m dbus.Message, coordType uint32) (extents, *dbus.Error) {
	r, _, err := c.bounds(m, coordType)
	return extents{r.X, r.Y, r.Width, r.Height}, err
}

func (c *component) GetPosition(m dbus.Message, coordType uint32) (int32, int32, *dbus.Error) {
	r, _, err := c.bounds(m, coordType)
	return r.X, r.Y, err
}

func (c *component) GetSize(m dbus.Message) (int32, int32, *dbus.Error) {
	r, _, err := c.bounds(m, 1)
	return r.Width, r.Height, err
}

func (c *component) GetLayer(m dbus.Message) (uint32, *dbus.Error) {
	id, n, err := c.b.lookup(m)
	if err != nil {
		return 0, err
	}
	if id != 0 && n.Parent == 0 {
		return 7, nil // ATSPI_LAYER_WINDOW
	}
	return 3, nil // ATSPI_LAYER_WIDGET
}

func (c *component) GetMDIZOrder(m dbus.Message) (int16, *dbus.Error) { return 0, nil }

func (c *component) GrabFocus(m dbus.Message) (bool, *dbus.Error) {
	id, _, err := c.b.lookup(m)
	if err != nil || c.b.handler.GrabFocus == nil {
		return false, err
	}
	return c.b.handler.GrabFocus(id), nil
}

func (c *component) GetAlpha(m dbus.Message) (float64, *dbus.Error) { return 1, nil }

func (c *component) SetExtents(m dbus.Message, x, y, w, h int32, coordType uint32) (bool, *dbus.Error) {
	return false, nil
}

func (c *component) SetPosition(m dbus.Message, x, y int32, coordType uint32) (bool, *dbus.Error) {
	return false, nil
}

func (c *component) SetSize(m dbus.Message, w, h int32) (bool, *dbus.Error) { return false, nil }

func (c *component) ScrollTo(m dbus.Message, kind uint32) (bool, *dbus.Error) { return false, nil }

func (c *component) ScrollToPoint(m dbus.Message, coordType uint32, x, y int32) (bool, *dbus.Error) {
	return false, nil
}

// action is org.a11y.atspi.Action.
type action struct{ b *Bridge }

func (a *action) name(m dbus.Message, i int32) (string, *dbus.Error) {
	_, n, err := a.b.lookup(m)
	if err != nil || n == nil || i < 0 || int(i) >= len(n.Actions) {
		return "", err
	}
	return n.Actions[i], nil
}

func (a *action) GetDescription(m dbus.Message, i int32) (string, *dbus.Error)   { return a.name(m, i) }
func (a *action) GetName(m dbus.Message, i int32) (string, *dbus.Error)          { return a.name(m, i) }
func (a *action) GetLocalizedName(m dbus.Message, i int32) (string, *dbus.Error) { return a.name(m, i) }
func (a *action) GetKeyBinding(m dbus.Message, i int32) (string, *dbus.Error)    { return "", nil }

type actionInfo struct{ Name, Description, KeyBinding string }

func (a *action) GetActions(m dbus.Message) ([]actionInfo, *dbus.Error) {
	_, n, err := a.b.lookup(m)
	if err != nil || n == nil {
		return nil, err
	}
	var out []actionInfo
	for _, s := range n.Actions {
		out = append(out, actionInfo{s, s, ""})
	}
	return out, nil
}

func (a *action) DoAction(m dbus.Message, i int32) (bool, *dbus.Error) {
	id, n, err := a.b.lookup(m)
	if err != nil || n == nil || a.b.handler.DoAction == nil || i < 0 || int(i) >= len(n.Actions) {
		return false, err
	}
	return a.b.handler.DoAction(id, int(i)), nil
}

// text is org.a11y.atspi.Text.
type text struct{ b *Bridge }

func (t *text) of(m dbus.Message) (uint64, textOf, *dbus.Error) {
	id, n, err := t.b.lookup(m)
	if err != nil {
		return 0, textOf{}, err
	}
	if n == nil || n.Text == nil {
		return id, textOf{}, dbus.MakeFailedError(errNoText)
	}
	return id, newTextOf(n.Text), nil
}

func (t *text) GetStringAtOffset(m dbus.Message, offset int32, granularity uint32) (string, int32, int32, *dbus.Error) {
	_, x, err := t.of(m)
	if err != nil {
		return "", 0, 0, err
	}
	s, a, b := x.stringAt(int(offset), granularity)
	return s, a, b, nil
}

func (t *text) GetText(m dbus.Message, start, end int32) (string, *dbus.Error) {
	_, x, err := t.of(m)
	if err != nil {
		return "", err
	}
	if end < 0 {
		end = int32(len(x.r))
	}
	return x.slice(int(start), int(end)), nil
}

func (t *text) SetCaretOffset(m dbus.Message, offset int32) (bool, *dbus.Error) {
	id, _, err := t.of(m)
	if err != nil || t.b.handler.SetCaret == nil {
		return false, err
	}
	return t.b.handler.SetCaret(id, int(offset)), nil
}

func (t *text) GetTextBeforeOffset(m dbus.Message, offset int32, kind uint32) (string, int32, int32, *dbus.Error) {
	return t.around(m, offset, kind, -1)
}

func (t *text) GetTextAtOffset(m dbus.Message, offset int32, kind uint32) (string, int32, int32, *dbus.Error) {
	return t.around(m, offset, kind, 0)
}

func (t *text) GetTextAfterOffset(m dbus.Message, offset int32, kind uint32) (string, int32, int32, *dbus.Error) {
	return t.around(m, offset, kind, 1)
}

func (t *text) around(m dbus.Message, offset int32, kind uint32, n int) (string, int32, int32, *dbus.Error) {
	_, x, err := t.of(m)
	if err != nil {
		return "", 0, 0, err
	}
	s, a, b := x.atOffset(int(offset), kind, n)
	return s, a, b, nil
}

func (t *text) GetCharacterAtOffset(m dbus.Message, offset int32) (int32, *dbus.Error) {
	_, x, err := t.of(m)
	if err != nil || offset < 0 || int(offset) >= len(x.r) {
		return 0, err
	}
	return x.r[offset], nil
}

func (t *text) GetAttributeValue(m dbus.Message, offset int32, name string) (string, *dbus.Error) {
	return "", nil
}

func (t *text) GetAttributes(m dbus.Message, offset int32) (map[string]string, int32, int32, *dbus.Error) {
	_, x, err := t.of(m)
	return map[string]string{}, 0, int32(len(x.r)), err
}

func (t *text) GetDefaultAttributes(m dbus.Message) (map[string]string, *dbus.Error) {
	return map[string]string{}, nil
}

func (t *text) GetCharacterExtents(m dbus.Message, offset int32, coordType uint32) (int32, int32, int32, int32, *dbus.Error) {
	id, n, err := t.b.lookup(m)
	if err != nil || n == nil || n.Text == nil || n.Text.CharacterBounds == nil {
		return 0, 0, 0, 0, err
	}
	r := t.b.screen(n.Text.CharacterBounds(int(offset)), t.b.windowOf(id), coordType)
	return r.X, r.Y, r.Width, r.Height, nil
}

func (t *text) GetOffsetAtPoint(m dbus.Message, x, y int32, coordType uint32) (int32, *dbus.Error) {
	return -1, nil
}

func (t *text) GetNSelections(m dbus.Message) (int32, *dbus.Error) {
	id, n, err := t.b.lookup(m)
	if err != nil || id == 0 || n.Text == nil || n.Text.SelectionStart == n.Text.SelectionEnd {
		return 0, err
	}
	return 1, nil
}

func (t *text) GetSelection(m dbus.Message, num int32) (int32, int32, *dbus.Error) {
	id, n, err := t.b.lookup(m)
	if err != nil || id == 0 || n.Text == nil || num != 0 {
		return 0, 0, err
	}
	return int32(n.Text.SelectionStart), int32(n.Text.SelectionEnd), nil
}

func (t *text) AddSelection(m dbus.Message, start, end int32) (bool, *dbus.Error) {
	return t.SetSelection(m, 0, start, end)
}

func (t *text) RemoveSelection(m dbus.Message, num int32) (bool, *dbus.Error) {
	id, n, err := t.b.lookup(m)
	if err != nil || id == 0 || n.Text == nil || t.b.handler.SetSelection == nil {
		return false, err
	}
	return t.b.handler.SetSelection(id, n.Text.Caret, n.Text.Caret), nil
}

func (t *text) SetSelection(m dbus.Message, num, start, end int32) (bool, *dbus.Error) {
	id, _, err := t.b.lookup(m)
	if err != nil || id == 0 || t.b.handler.SetSelection == nil {
		return false, err
	}
	return t.b.handler.SetSelection(id, int(start), int(end)), nil
}

func (t *text) GetRangeExtents(m dbus.Message, start, end int32, coordType uint32) (int32, int32, int32, int32, *dbus.Error) {
	x, y, w, h, err := t.GetCharacterExtents(m, start, coordType)
	return x, y, w, h, err
}

type rangeOf struct {
	Start, End int32
	Content    dbus.Variant
}

func (t *text) GetBoundedRanges(m dbus.Message, x, y, w, h int32, coordType, xClip, yClip uint32) ([]rangeOf, *dbus.Error) {
	return []rangeOf{}, nil
}

func (t *text) GetAttributeRun(m dbus.Message, offset int32, includeDefaults bool) (map[string]string, int32, int32, *dbus.Error) {
	return t.GetAttributes(m, offset)
}

func (t *text) GetDefaultAttributeSet(m dbus.Message) (map[string]string, *dbus.Error) {
	return map[string]string{}, nil
}

func (t *text) ScrollSubstringTo(m dbus.Message, start, end int32, kind uint32) (bool, *dbus.Error) {
	return false, nil
}

func (t *text) ScrollSubstringToPoint(m dbus.Message, start, end int32, coordType uint32, x, y int32) (bool, *dbus.Error) {
	return false, nil
}

// properties is org.freedesktop.DBus.Properties for every object.
type properties struct{ b *Bridge }

func (p *properties) Get(m dbus.Message, iface, name string) (dbus.Variant, *dbus.Error) {
	all, err := p.GetAll(m, iface)
	if err != nil {
		return dbus.Variant{}, err
	}
	v, ok := all[name]
	if !ok {
		return dbus.Variant{}, dbus.NewError("org.freedesktop.DBus.Error.UnknownProperty", []any{iface + "." + name})
	}
	return v, nil
}

func (p *properties) GetAll(m dbus.Message, iface string) (map[string]dbus.Variant, *dbus.Error) {
	id, n, err := p.b.lookup(m)
	if err != nil {
		return nil, err
	}
	v := dbus.MakeVariant
	switch iface {
	case ifaceAccessible:
		name, desc := p.b.app.Name, ""
		if id != 0 {
			name, desc = n.Name, n.Description
		}
		return map[string]dbus.Variant{
			"Name": v(name), "Description": v(desc), "Parent": v(p.b.parentRef(id, n)),
			"ChildCount": v(int32(len(p.b.children(id, n)))), "Locale": v(""),
			"AccessibleId": v(strconv.FormatUint(id, 10)), "HelpText": v(""), "version": v(uint32(1)),
		}, nil
	case ifaceApplication:
		p.b.mu.RLock()
		appID := p.b.appID
		p.b.mu.RUnlock()
		return map[string]dbus.Variant{
			"ToolkitName": v(p.b.app.ToolkitName), "Version": v(p.b.app.ToolkitVersion),
			"ToolkitVersion": v(p.b.app.ToolkitVersion), "AtspiVersion": v("2.1"),
			"InterfaceVersion": v(uint32(1)), "Id": v(appID), "version": v(uint32(1)),
		}, nil
	case ifaceText:
		if n == nil || n.Text == nil {
			return map[string]dbus.Variant{}, nil
		}
		return map[string]dbus.Variant{
			"CharacterCount": v(int32(len([]rune(n.Text.Content)))), "CaretOffset": v(int32(n.Text.Caret)),
			"version": v(uint32(1)),
		}, nil
	case ifaceAction:
		count := 0
		if n != nil {
			count = len(n.Actions)
		}
		return map[string]dbus.Variant{"NActions": v(int32(count)), "version": v(uint32(1))}, nil
	case ifaceComponent:
		return map[string]dbus.Variant{"version": v(uint32(1))}, nil
	}
	return map[string]dbus.Variant{}, nil
}

func (p *properties) Set(m dbus.Message, iface, name string, value dbus.Variant) *dbus.Error {
	if iface == ifaceApplication && name == "Id" {
		if id, ok := value.Value().(int32); ok {
			p.b.mu.Lock()
			p.b.appID = id
			p.b.mu.Unlock()
		}
		return nil
	}
	return dbus.NewError("org.freedesktop.DBus.Error.PropertyReadOnly", []any{iface + "." + name})
}
