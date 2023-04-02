package booker

import (
	"github.com/iotaledger/goshimmer/packages/core/votes/conflicttracker"
	"github.com/iotaledger/goshimmer/packages/core/votes/sequencetracker"
	"github.com/iotaledger/goshimmer/packages/core/votes/slottracker"
	"github.com/iotaledger/goshimmer/packages/protocol/markers"
	"github.com/iotaledger/hive.go/runtime/event"
	"github.com/iotaledger/iota-core/pkg/interfaces/ledger/utxo"
)

type Events struct {
	BlockBooked         *event.Event1[*BlockBookedEvent]
	AttachmentCreated   *event.Event1[*Block]
	AttachmentOrphaned  *event.Event1[*Block]
	BlockConflictAdded  *event.Event1[*BlockConflictAddedEvent]
	MarkerConflictAdded *event.Event1[*MarkerConflictAddedEvent]
	Error               *event.Event1[error]

	SequenceEvicted *event.Event1[markers.SequenceID]

	VirtualVoting *VirtualVotingEvents

	event.Group[Events, *Events]
}

// NewEvents contains the constructor of the Events object (it is generated by a generic factory).
var NewEvents = event.CreateGroupConstructor(func() (newEvents *Events) {
	return &Events{
		BlockBooked:         event.New1[*BlockBookedEvent](),
		AttachmentCreated:   event.New1[*Block](),
		AttachmentOrphaned:  event.New1[*Block](),
		BlockConflictAdded:  event.New1[*BlockConflictAddedEvent](),
		MarkerConflictAdded: event.New1[*MarkerConflictAddedEvent](),
		Error:               event.New1[error](),

		SequenceEvicted: event.New1[markers.SequenceID](),
		VirtualVoting:   NewVirtualVotingEvents(),
	}
})

type BlockConflictAddedEvent struct {
	Block             *Block
	ConflictID        utxo.TransactionID
	ParentConflictIDs utxo.TransactionIDs
}

type MarkerConflictAddedEvent struct {
	Marker            markers.Marker
	ConflictID        utxo.TransactionID
	ParentConflictIDs utxo.TransactionIDs
}

type BlockBookedEvent struct {
	Block       *Block
	ConflictIDs utxo.TransactionIDs
}

type VirtualVotingEvents struct {
	BlockTracked *event.Event1[*Block]

	ConflictTracker *conflicttracker.Events[utxo.TransactionID]
	SequenceTracker *sequencetracker.Events
	SlotTracker     *slottracker.Events

	event.Group[VirtualVotingEvents, *VirtualVotingEvents]
}

// NewVirtualVotingEvents contains the constructor of the VirtualVotingEvents object (it is generated by a generic factory).
var NewVirtualVotingEvents = event.CreateGroupConstructor(func() (newEvents *VirtualVotingEvents) {
	return &VirtualVotingEvents{
		BlockTracked:    event.New1[*Block](),
		ConflictTracker: conflicttracker.NewEvents[utxo.TransactionID](),
		SequenceTracker: sequencetracker.NewEvents(),
		SlotTracker:     slottracker.NewEvents(),
	}
})
