package tipmanager

import (
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/iota-core/pkg/protocol/engine/blockdag"
)

// Events represents events happening on the TipManager.
type Events struct {
	// Fired when a tip is added.
	TipAdded *event.Event1[*blockdag.Block]

	// Fired when a tip is removed.
	TipRemoved *event.Event1[*blockdag.Block]

	event.Group[Events, *Events]
}

// NewEvents contains the constructor of the Events object (it is generated by a generic factory).
var NewEvents = event.CreateGroupConstructor(func() (newEvents *Events) {
	return &Events{
		TipAdded:   event.New1[*blockdag.Block](),
		TipRemoved: event.New1[*blockdag.Block](),
	}
})
