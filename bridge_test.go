package atspi

import (
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
)

// client connects to the accessibility bus as a screen reader would.
func client(t *testing.T) *dbus.Conn {
	t.Helper()
	session, err := dbus.SessionBus()
	if err != nil {
		t.Skip("no session bus")
	}
	var addr string
	if err := session.Object("org.a11y.Bus", "/org/a11y/bus").Call("org.a11y.Bus.GetAddress", 0).Store(&addr); err != nil {
		t.Skip("no accessibility bus")
	}
	c, err := dbus.Dial(addr)
	if err != nil {
		t.Skip(err)
	}
	if err := c.Auth(nil); err != nil {
		t.Fatal(err)
	}
	if err := c.Hello(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func sampleNodes() []Node {
	return []Node{
		{ID: 1, Role: RoleFrame, Name: "Accolade", States: []State{StateActive, StateShowing, StateVisible, StateEnabled},
			Children: []uint64{2, 3, 4}, Bounds: Rect{0, 0, 800, 600}},
		{ID: 2, Parent: 1, Role: RoleLabel, Name: "Title", States: []State{StateShowing, StateVisible},
			Bounds: Rect{10, 10, 100, 20}, Text: &Text{Content: "Title", Caret: -1}},
		{ID: 3, Parent: 1, Role: RoleButton, Name: "Save", States: []State{StateFocusable, StateShowing, StateVisible, StateEnabled, StateSensitive},
			Bounds: Rect{10, 40, 80, 30}, Actions: []string{"click"}},
		{ID: 4, Parent: 1, Role: RoleText, Name: "Script", Bounds: Rect{10, 80, 700, 500},
			States: []State{StateEditable, StateMultiLine, StateFocusable, StateShowing, StateVisible, StateEnabled, StateSensitive},
			Text:   &Text{Content: "INT. BARN - DAY\n\nRain falls on the roof.", Caret: 3}},
	}
}

// The application appears on the desktop, and a screen reader can walk it
// and read it; focus changes reach it as events.
func TestOnTheAccessibilityBus(t *testing.T) {
	reader := client(t)
	clicked := make(chan uint64, 1)
	b, err := Connect(AppInfo{Name: "atspi-test", ToolkitName: "test", ToolkitVersion: "0"},
		Handler{DoAction: func(id uint64, i int) bool { clicked <- id; return true }})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	b.Update(sampleNodes())

	// the desktop lists the application
	var apps []ref
	desktop := reader.Object("org.a11y.atspi.Registry", rootPath)
	if err := desktop.Call(ifaceAccessible+".GetChildren", 0).Store(&apps); err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(apps, func(r ref) bool { return r.Name == b.BusName() }) {
		t.Fatalf("not among the desktop's %d applications", len(apps))
	}

	app := reader.Object(b.BusName(), rootPath)
	name, err := app.GetProperty(ifaceAccessible + ".Name")
	if err != nil || name.Value() != "atspi-test" {
		t.Errorf("application name %v %v", name, err)
	}
	var windows []ref
	if err := app.Call(ifaceAccessible+".GetChildren", 0).Store(&windows); err != nil || len(windows) != 1 {
		t.Fatalf("windows %v %v", windows, err)
	}
	win := reader.Object(windows[0].Name, windows[0].Path)
	var role uint32
	win.Call(ifaceAccessible+".GetRole", 0).Store(&role)
	if Role(role) != RoleFrame {
		t.Errorf("window role %d", role)
	}
	var kids []ref
	win.Call(ifaceAccessible+".GetChildren", 0).Store(&kids)
	if len(kids) != 3 {
		t.Fatalf("window children %v", kids)
	}

	// the button: its name, role, states and action
	button := reader.Object(kids[1].Name, kids[1].Path)
	if v, _ := button.GetProperty(ifaceAccessible + ".Name"); v.Value() != "Save" {
		t.Errorf("button name %v", v)
	}
	var roleName string
	button.Call(ifaceAccessible+".GetRoleName", 0).Store(&roleName)
	if roleName != "push button" {
		t.Errorf("button role name %q", roleName)
	}
	var states []uint32
	button.Call(ifaceAccessible+".GetState", 0).Store(&states)
	if len(states) != 2 || states[0]&(1<<StateFocusable) == 0 {
		t.Errorf("button states %b", states)
	}
	var ifaces []string
	button.Call(ifaceAccessible+".GetInterfaces", 0).Store(&ifaces)
	if !slices.Contains(ifaces, ifaceAction) || slices.Contains(ifaces, ifaceText) {
		t.Errorf("button interfaces %v", ifaces)
	}
	var ok bool
	if err := button.Call(ifaceAction+".DoAction", 0, int32(0)).Store(&ok); err != nil || !ok {
		t.Errorf("DoAction %v %v", ok, err)
	}
	select {
	case id := <-clicked:
		if id != 3 {
			t.Errorf("action on %d", id)
		}
	case <-time.After(time.Second):
		t.Error("the action did not reach the handler")
	}
	var ext extents
	button.Call(ifaceComponent+".GetExtents", 0, uint32(1)).Store(&ext)
	if ext != (extents{10, 40, 80, 30}) {
		t.Errorf("button extents %v", ext)
	}

	// the script: read by word and line
	script := reader.Object(kids[2].Name, kids[2].Path)
	if v, _ := script.GetProperty(ifaceText + ".CaretOffset"); v.Value() != int32(3) {
		t.Errorf("caret %v", v)
	}
	var s string
	var start, end int32
	script.Call(ifaceText+".GetStringAtOffset", 0, int32(2), granularityWord).Store(&s, &start, &end)
	if s != "INT" || start != 0 || end != 3 {
		t.Errorf("word at 2: %q %d %d", s, start, end)
	}
	script.Call(ifaceText+".GetStringAtOffset", 0, int32(20), granularityLine).Store(&s, &start, &end)
	if s != "Rain falls on the roof." {
		t.Errorf("line at 20: %q %d %d", s, start, end)
	}
	script.Call(ifaceText+".GetTextAtOffset", 0, int32(5), boundaryLineStart).Store(&s, &start, &end)
	if s != "INT. BARN - DAY\n" {
		t.Errorf("line (boundary) at 5: %q", s)
	}

	// events: the focus moving to the script
	if err := reader.AddMatchSignal(dbus.WithMatchInterface("org.a11y.atspi.Event.Object"), dbus.WithMatchMember("StateChanged")); err != nil {
		t.Fatal(err)
	}
	signals := make(chan *dbus.Signal, 10)
	reader.Signal(signals)
	b.Focused(4)
	deadline := time.After(2 * time.Second)
	for {
		select {
		case sig := <-signals:
			if sig.Name == "org.a11y.atspi.Event.Object.StateChanged" && strings.HasSuffix(string(sig.Path), "/4") &&
				sig.Body[0] == "focused" && sig.Body[1] == int32(1) {
				return
			}
		case <-deadline:
			t.Fatal("no focus event")
		}
	}
}
