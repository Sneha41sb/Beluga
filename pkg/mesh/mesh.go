package mesh

import (
	"sync"
	"uftp/pkg/frame"
)

// MeshNode represents an active peer in the acoustic mesh network
type MeshNode struct {
	ID           string
	mu           sync.Mutex
	seenMessages map[uint32]bool
}

// NewMeshNode initializes a new MeshNode
func NewMeshNode(nodeID string) *MeshNode {
	return &MeshNode{
		ID:           nodeID,
		seenMessages: make(map[uint32]bool),
	}
}

// ProcessIncomingFrame processes an incoming frame from microphone demodulation.
// Returns (relayFrame, shouldRelay):
// - shouldRelay is true if this node should re-broadcast the message to extend range.
// - relayFrame contains the updated frame with decremented TTL.
func (n *MeshNode) ProcessIncomingFrame(f frame.Frame) (frame.Frame, bool) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// 1. Deduplication: Check if we have already processed/relayed this MessageID
	if n.seenMessages[f.MessageID] {
		return frame.Frame{}, false // Ignore duplicate
	}

	// 2. Mark as seen
	n.seenMessages[f.MessageID] = true

	// 3. Check TTL for relay eligibility
	if f.TTL <= 1 {
		return f, false // Do not relay if TTL expired
	}

	// 4. Create relay frame with decremented TTL
	relayFrame := frame.Frame{
		Type:      f.Type,
		TTL:       f.TTL - 1,
		MessageID: f.MessageID,
		Payload:   f.Payload,
	}

	return relayFrame, true
}
