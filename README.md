# atspi

Linux screen reader support (AT-SPI2) for Go GUIs, without C.

A toolkit, or an application, hands the package a snapshot of its accessible
objects (`[]atspi.Node`: roles, names, states, bounds, actions, text) whenever
they change, and tells it what happens: the focus moving, a window becoming
active, text being typed. The `Bridge` answers screen readers (Orca) from the
latest snapshot on its own goroutines, never waiting for the UI thread, and
passes their requests (an action, moving the caret) back through a `Handler`.

It knows no toolkit; a Fyne bridge is being written in a fork of Fyne.

    go run ./cmd/atspi-demo   # a small window on the accessibility bus

The role and state numbers are generated from at-spi2-core's
`atspi-constants.h`; the method signatures follow its `xml/*.xml`.
