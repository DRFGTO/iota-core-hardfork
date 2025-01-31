package mock

import (
	"crypto/ed25519"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/crypto/identity"
	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/hive.go/runtime/workerpool"
	"github.com/iotaledger/iota-core/pkg/model"
	"github.com/iotaledger/iota-core/pkg/network"
	"github.com/iotaledger/iota-core/pkg/protocol"
	"github.com/iotaledger/iota-core/pkg/protocol/engine"
	"github.com/iotaledger/iota-core/pkg/protocol/engine/blocks"
	"github.com/iotaledger/iota-core/pkg/protocol/engine/filter"
	"github.com/iotaledger/iota-core/pkg/protocol/engine/notarization"
	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/iota.go/v4/builder"
)

type Node struct {
	Testing *testing.T

	Name   string
	Weight int64

	privateKey ed25519.PrivateKey
	pubKey     ed25519.PublicKey
	AccountID  iotago.AccountID
	PeerID     network.PeerID

	Endpoint *Endpoint
	Workers  *workerpool.Group

	optsProtocolOptions []options.Option[protocol.Protocol]

	Protocol *protocol.Protocol
}

func NewNode(t *testing.T, net *Network, partition string, name string, weight int64, opts ...options.Option[protocol.Protocol]) *Node {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		panic(err)
	}

	accountID := iotago.AccountID(*iotago.Ed25519AddressFromPubKey(pub))
	accountID.RegisterAlias(name)

	peerID := network.PeerID(pub)
	identity.RegisterIDAlias(peerID, name)

	return &Node{
		Testing: t,

		Name:       name,
		Weight:     weight,
		pubKey:     pub,
		privateKey: priv,
		AccountID:  accountID,
		PeerID:     peerID,

		Endpoint:            net.Join(peerID, partition),
		Workers:             workerpool.NewGroup(name),
		optsProtocolOptions: opts,
	}
}

func (n *Node) Initialize(opts ...options.Option[protocol.Protocol]) {
	n.Protocol = protocol.New(n.Workers.CreateGroup("Protocol"),
		n.Endpoint,
		append(opts, n.optsProtocolOptions...)...,
	)

	n.Protocol.Run()
}

func (n *Node) HookLogging() {
	events := n.Protocol.Events

	n.attachEngineLogs(n.Protocol.MainEngineInstance())

	events.Network.BlockReceived.Hook(func(block *model.Block, source identity.ID) {
		fmt.Printf("%s > Network.BlockReceived: from %s %s - %d\n", n.Name, source, block.ID(), block.ID().Index())
	})

	events.Network.BlockRequestReceived.Hook(func(blockID iotago.BlockID, source identity.ID) {
		fmt.Printf("%s > Network.BlockRequestReceived: from %s %s\n", n.Name, source, blockID)
	})

	// events.Network.AttestationsReceived.Hook(func(event *network.AttestationsReceivedEvent) {
	// 	fmt.Printf("%s > Network.AttestationsReceived: from %s for %s\n", n.Name, event.Source, event.Commitment.ID())
	// })
	//
	// events.Network.AttestationsRequestReceived.Hook(func(event *network.AttestationsRequestReceivedEvent) {
	// 	fmt.Printf("%s > Network.AttestationsRequestReceived: from %s %s -> %d\n", n.Name, event.Source, event.Commitment.ID(), event.EndIndex)
	// })
	//
	// events.Network.SlotCommitmentReceived.Hook(func(event *network.SlotCommitmentReceivedEvent) {
	// 	fmt.Printf("%s > Network.SlotCommitmentReceived: from %s %s\n", n.Name, event.Source, event.Commitment.ID())
	// })
	//
	// events.Network.SlotCommitmentRequestReceived.Hook(func(event *network.SlotCommitmentRequestReceivedEvent) {
	// 	fmt.Printf("%s > Network.SlotCommitmentRequestReceived: from %s %s\n", n.Name, event.Source, event.CommitmentID)
	// })

	events.Network.Error.Hook(func(err error, id identity.ID) {
		fmt.Printf("%s > Network.Error: from %s %s\n", n.Name, id, err)
	})
}

func (n *Node) attachEngineLogs(instance *engine.Engine) {
	engineName := fmt.Sprintf("%s - %s", lo.Cond(n.Protocol.MainEngineInstance() != instance, "Candidate", "Main"), instance.Name()[:8])
	events := instance.Events

	events.BlockDAG.BlockAttached.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] BlockDAG.BlockAttached: %s\n", n.Name, engineName, block.ID())
	})

	events.BlockDAG.BlockSolid.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] BlockDAG.BlockSolid: %s\n", n.Name, engineName, block.ID())
	})

	events.BlockDAG.BlockInvalid.Hook(func(block *blocks.Block, err error) {
		fmt.Printf("%s > [%s] BlockDAG.BlockInvalid: %s - %s\n", n.Name, engineName, block.ID(), err)
	})

	events.BlockDAG.BlockMissing.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] BlockDAG.BlockMissing: %s\n", n.Name, engineName, block.ID())
	})

	events.BlockDAG.MissingBlockAttached.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] BlockDAG.MissingBlockAttached: %s\n", n.Name, engineName, block.ID())
	})

	events.Booker.BlockBooked.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] Booker.BlockBooked: %s\n", n.Name, engineName, block.ID())
	})

	events.Clock.AcceptedTimeUpdated.Hook(func(newTime time.Time) {
		fmt.Printf("%s > [%s] Clock.AcceptedTimeUpdated: %s\n", n.Name, engineName, newTime)
	})

	events.Filter.BlockAllowed.Hook(func(block *model.Block) {
		fmt.Printf("%s > [%s] Filter.BlockAllowed: %s\n", n.Name, engineName, block.ID())
	})

	events.Filter.BlockFiltered.Hook(func(event *filter.BlockFilteredEvent) {
		fmt.Printf("%s > [%s] Filter.BlockFiltered: %s - %s\n", n.Name, engineName, event.Block.ID(), event.Reason.Error())
		n.Testing.Fatal("no blocks should be filtered")
	})

	events.BlockRequester.Tick.Hook(func(blockID iotago.BlockID) {
		fmt.Printf("%s > [%s] BlockRequester.Tick: %s\n", n.Name, engineName, blockID)
	})

	events.BlockProcessed.Hook(func(blockID iotago.BlockID) {
		fmt.Printf("%s > [%s] Engine.BlockProcessed: %s\n", n.Name, engineName, blockID)
	})

	events.Error.Hook(func(err error) {
		fmt.Printf("%s > [%s] Engine.Error: %s\n", n.Name, engineName, err.Error())
	})

	events.Notarization.SlotCommitted.Hook(func(details *notarization.SlotCommittedDetails) {
		fmt.Printf("%s > [%s] NotarizationManager.SlotCommitted: %s %s\n", n.Name, engineName, details.Commitment.ID(), details.Commitment)
	})

	events.BlockGadget.BlockAccepted.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] Consensus.BlockGadget.BlockAccepted: %s %s\n", n.Name, engineName, block.ID(), block.Block().SlotCommitment.MustID())
	})

	events.BlockGadget.BlockConfirmed.Hook(func(block *blocks.Block) {
		fmt.Printf("%s > [%s] Consensus.BlockGadget.BlockConfirmed: %s %s\n", n.Name, engineName, block.ID(), block.Block().SlotCommitment.MustID())
	})

	events.SlotGadget.SlotFinalized.Hook(func(slotIndex iotago.SlotIndex) {
		fmt.Printf("%s > [%s] Consensus.SlotGadget.SlotConfirmed: %s\n", n.Name, engineName, slotIndex)
	})
}

func (n *Node) Wait() {
	n.Workers.WaitChildren()
}

func (n *Node) Shutdown() {
	n.Protocol.Shutdown()
	n.Workers.Shutdown()
}

// TODO: the block Issuance should be improved once the protocol has a better way to issue blocks.

func (n *Node) IssueBlock() iotago.BlockID {
	references := n.Protocol.TipManager.Tips(iotago.BlockMaxParents)
	block, err := builder.NewBlockBuilder().
		StrongParents(references[model.StrongParentType]).
		WeakParents(references[model.WeakParentType]).
		ShallowLikeParents(references[model.ShallowLikeParentType]).
		SlotCommitment(n.Protocol.MainEngineInstance().Storage.Settings().LatestCommitment().Commitment()).
		LatestFinalizedSlot(n.Protocol.MainEngineInstance().Storage.Settings().LatestFinalizedSlot()).
		Payload(&iotago.TaggedData{
			Tag: []byte("ACTIVITY"),
		}).
		Sign(n.AccountID, n.privateKey).
		Build()

	if err != nil {
		panic(err)
	}

	modelBlock, err := model.BlockFromBlock(block, n.Protocol.API())
	if err != nil {
		panic(err)
	}

	err = n.Protocol.ProcessBlock(modelBlock, n.PeerID)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Issued block: %s - commitment %s %d - latest finalized slot %d\n", modelBlock.ID(), modelBlock.Block().SlotCommitment.MustID(), modelBlock.Block().SlotCommitment.Index, modelBlock.Block().LatestFinalizedSlot)

	return modelBlock.ID()
}

func (n *Node) IssueBlockAtSlot(alias string, slot iotago.SlotIndex, parents ...iotago.BlockID) *model.Block {
	slotTimeProvider := n.Protocol.MainEngineInstance().Storage.Settings().API().SlotTimeProvider()
	issuingTime := time.Unix(slotTimeProvider.GenesisUnixTime()+int64(slot-1)*slotTimeProvider.Duration(), 0)
	require.True(n.Testing, issuingTime.Before(time.Now()), "issued block is in the current or future slot")

	block, err := builder.NewBlockBuilder().
		StrongParents(parents).
		IssuingTime(issuingTime).
		SlotCommitment(n.Protocol.MainEngineInstance().Storage.Settings().LatestCommitment().Commitment()).
		LatestFinalizedSlot(n.Protocol.MainEngineInstance().Storage.Settings().LatestFinalizedSlot()).
		Payload(&iotago.TaggedData{
			Tag: []byte("ACTIVITY"),
		}).
		Sign(n.AccountID, n.privateKey).
		Build()

	if err != nil {
		panic(err)
	}

	modelBlock, err := model.BlockFromBlock(block, n.Protocol.API())
	if err != nil {
		panic(err)
	}

	err = n.Protocol.ProcessBlock(modelBlock, n.PeerID)
	if err != nil {
		panic(err)
	}

	modelBlock.ID().RegisterAlias(alias)
	fmt.Printf("Issued block: %s - commitment %s %d - latest finalized slot %d\n", modelBlock.ID(), modelBlock.Block().SlotCommitment.MustID(), modelBlock.Block().SlotCommitment.Index, modelBlock.Block().LatestFinalizedSlot)

	return modelBlock
}

func (n *Node) IssueActivity(duration time.Duration, wg *sync.WaitGroup) {
	go func() {
		defer wg.Done()

		start := time.Now()
		fmt.Println(n.Name, "> Starting activity")
		var counter int
		for {
			if references := n.Protocol.TipManager.Tips(iotago.BlockMaxParents); len(references[model.StrongParentType]) > 0 {
				if !n.issueActivityBlock(fmt.Sprintf("activity %s.%d", n.Name, counter), references[model.StrongParentType]...) {
					fmt.Println(n.Name, "> Stopped activity due to block not being issued")
					return
				}
				counter++
				time.Sleep(1 * time.Second)
				if duration > 0 && time.Since(start) > duration {
					fmt.Println(n.Name, "> Stopped activity after", time.Since(start))
					return
				}
			} else {
				fmt.Println(n.Name, "> Skipped activity due lack of strong parents")
			}
		}
	}()
}

func (n *Node) issueActivityBlock(alias string, parents ...iotago.BlockID) bool {
	if !n.Protocol.MainEngineInstance().WasStopped() {

		block, err := builder.NewBlockBuilder().
			StrongParents(parents).
			SlotCommitment(n.Protocol.MainEngineInstance().Storage.Settings().LatestCommitment().Commitment()).
			LatestFinalizedSlot(n.Protocol.MainEngineInstance().Storage.Settings().LatestFinalizedSlot()).
			Payload(&iotago.TaggedData{
				Tag: []byte("ACTIVITY"),
			}).
			Sign(n.AccountID, n.privateKey).
			Build()

		if err != nil {
			panic(err)
		}

		modelBlock, err := model.BlockFromBlock(block, n.Protocol.API())
		if err != nil {
			panic(err)
		}

		err = n.Protocol.ProcessBlock(modelBlock, n.PeerID)
		if err != nil {
			panic(err)
		}

		modelBlock.ID().RegisterAlias(alias)

		return true
	}

	return false
}
