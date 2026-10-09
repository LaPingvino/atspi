package atspi

import (
	"slices"

	"github.com/godbus/dbus/v5"
)

const (
	cachePath  = "/org/a11y/atspi/cache"
	ifaceCache = "org.a11y.atspi.Cache"
)

// cacheItem is one object as org.a11y.atspi.Cache describes it, so that a
// screen reader can fetch the whole tree in one call:
// (so)(so)(so)iiassusau.
type cacheItem struct {
	Object, App, Parent ref
	IndexInParent       int32
	ChildCount          int32
	Interfaces          []string
	Name                string
	Role                uint32
	Description         string
	States              []uint32
}

// cache is org.a11y.atspi.Cache.
type cache struct{ b *Bridge }

func (c *cache) GetItems() ([]cacheItem, *dbus.Error) {
	b := c.b
	b.mu.RLock()
	defer b.mu.RUnlock()
	items := []cacheItem{}
	index := func(parent uint64, id uint64) int32 {
		if parent == 0 {
			return int32(slices.Index(b.windows, id))
		}
		if p := b.nodes[parent]; p != nil {
			return int32(slices.Index(p.Children, id))
		}
		return -1
	}
	for id, n := range b.nodes {
		bits := []uint32{0, 0}
		for _, s := range n.States {
			bits[s/32] |= 1 << (s % 32)
		}
		if b.focused == id {
			bits[StateFocused/32] |= 1 << (StateFocused % 32)
		}
		items = append(items, cacheItem{
			Object: b.ref(id), App: b.ref(0), Parent: b.ref(n.Parent),
			IndexInParent: index(n.Parent, id), ChildCount: int32(len(n.Children)),
			Interfaces: interfacesOf(id, n), Name: n.Name, Role: uint32(n.Role),
			Description: n.Description, States: bits,
		})
	}
	return items, nil
}
