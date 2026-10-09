// Command atspi-demo puts a small, fixed user interface on the
// accessibility bus (a window with a label, a button and a text field) and
// keeps it there, for trying screen readers and inspection tools against
// the atspi package.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/LaPingvino/atspi"
)

func main() {
	nodes := []atspi.Node{
		{ID: 1, Role: atspi.RoleFrame, Name: "atspi demo", Children: []uint64{2, 3, 4},
			States: []atspi.State{atspi.StateActive, atspi.StateShowing, atspi.StateVisible, atspi.StateEnabled},
			Bounds: atspi.Rect{Width: 800, Height: 600}},
		{ID: 2, Parent: 1, Role: atspi.RoleLabel, Name: "Scene", Bounds: atspi.Rect{X: 10, Y: 10, Width: 100, Height: 20},
			States: []atspi.State{atspi.StateShowing, atspi.StateVisible}},
		{ID: 3, Parent: 1, Role: atspi.RoleButton, Name: "Save", Actions: []string{"click"},
			States: []atspi.State{atspi.StateFocusable, atspi.StateShowing, atspi.StateVisible, atspi.StateEnabled, atspi.StateSensitive},
			Bounds: atspi.Rect{X: 10, Y: 40, Width: 80, Height: 30}},
		{ID: 4, Parent: 1, Role: atspi.RoleText, Name: "Script",
			States: []atspi.State{atspi.StateEditable, atspi.StateMultiLine, atspi.StateFocusable, atspi.StateShowing,
				atspi.StateVisible, atspi.StateEnabled, atspi.StateSensitive},
			Bounds: atspi.Rect{X: 10, Y: 80, Width: 700, Height: 500},
			Text:   &atspi.Text{Content: "INT. BARN - DAY\n\nRain falls on the roof.", Caret: 0}},
	}
	b, err := atspi.Connect(atspi.AppInfo{Name: "atspi demo", ToolkitName: "atspi", ToolkitVersion: "0"},
		atspi.Handler{DoAction: func(id uint64, i int) bool { fmt.Println("action", i, "on", id); return true }})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer b.Close()
	b.Update(nodes)
	b.WindowActivated(1, true)
	b.Focused(4)
	fmt.Println("on the accessibility bus as", b.BusName())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	select {
	case <-stop:
	case <-time.After(10 * time.Minute):
	}
}
