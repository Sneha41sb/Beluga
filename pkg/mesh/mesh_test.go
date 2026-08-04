package mesh

import (
	"testing"
	"uftp/pkg/frame"
)

func TestMeshMultiHopPropagation(t *testing.T) {
	nodeA := NewMeshNode("NodeA")
	nodeB := NewMeshNode("NodeB")
	nodeC := NewMeshNode("NodeC")

	_ = nodeA // Host Node A

	initialFrame := frame.Frame{
		Type:      frame.TypeMeshAlert,
		TTL:       3,
		MessageID: 42,
		Payload:   []byte("EMERGENCY ALERT: Safe meeting point at North Gate."),
	}

	// 1. Hop 1: Node A sends to Node B
	relayB, shouldRelayB := nodeB.ProcessIncomingFrame(initialFrame)
	if !shouldRelayB {
		t.Fatalf("Node B should relay the frame")
	}
	if relayB.TTL != 2 {
		t.Errorf("Expected TTL=2 at Node B relay, got %d", relayB.TTL)
	}

	// 2. Hop 2: Node B sends to Node C
	relayC, shouldRelayC := nodeC.ProcessIncomingFrame(relayB)
	if !shouldRelayC {
		t.Fatalf("Node C should relay the frame")
	}
	if relayC.TTL != 1 {
		t.Errorf("Expected TTL=1 at Node C relay, got %d", relayC.TTL)
	}

	// 3. Duplicate Detection: Node B receives Node C's relay again
	_, shouldRelayDuplicate := nodeB.ProcessIncomingFrame(relayC)
	if shouldRelayDuplicate {
		t.Fatalf("Node B should ignore duplicate MessageID 42")
	}

	t.Logf("Successfully verified multi-hop mesh relay & duplicate suppression!")
}
