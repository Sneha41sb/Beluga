package fec

import (
	"bytes"
	"testing"
)

func TestReedSolomonSelfHealing(t *testing.T) {
	config := FECConfig{
		DataShards:   4,
		ParityShards: 2, // Can recover from 2 completely missing/corrupted shards
	}

	fecEngine, err := NewFEC(config)
	if err != nil {
		t.Fatalf("Failed to initialize FEC: %v", err)
	}

	originalMessage := []byte("HIGHLY SENSITIVE EMERGENCY ULTRASONIC PAYLOAD DATA OVER AFSK MESH")
	originalLen := len(originalMessage)

	// 1. Encode into 4 Data Shards + 2 Parity Shards (Total 6 Shards)
	shards, err := fecEngine.Encode(originalMessage)
	if err != nil {
		t.Fatalf("Encoding failed: %v", err)
	}

	if len(shards) != 6 {
		t.Fatalf("Expected 6 shards total, got %d", len(shards))
	}

	// 2. Simulate Severe Acoustic Corruption: Destroy 2 Shards (Data Shard 1 & Parity Shard 4)
	shards[1] = nil
	shards[4] = nil

	t.Logf("Simulated acoustic damage: Shard 1 and Shard 4 wiped to nil")

	// 3. Reconstruct original data from remaining 4 shards
	reconstructedData, err := fecEngine.Reconstruct(shards, originalLen)
	if err != nil {
		t.Fatalf("Reconstruction failed: %v", err)
	}

	// 4. Assert 100% data recovery match
	if !bytes.Equal(reconstructedData, originalMessage) {
		t.Fatalf("Reconstructed data mismatch!\nExpected: '%s'\nGot:      '%s'", string(originalMessage), string(reconstructedData))
	}

	t.Logf("Reed-Solomon successfully reconstructed 100%% of original data despite 2 destroyed shards!")
}
