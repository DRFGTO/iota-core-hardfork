package ledger

import (
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/iota-core/pkg/interfaces/ledger/mempool"
)

// Events is a container that acts as a dictionary for the existing events of a MemPool.
type Events struct {
	// MemPool contains all mempool related events.
	MemPool *mempool.Events

	event.Group[Events, *Events]
}

// NewEvents contains the constructor of the Events object (it is generated by a generic factory).
var NewEvents = event.CreateGroupConstructor(func() (newEvents *Events) {
	return &Events{
		MemPool: mempool.NewEvents(),
	}
})
