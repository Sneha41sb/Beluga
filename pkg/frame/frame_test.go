package frame

import (
	"bytes"
	"testing"
)

func TestFrameEncodeDecode(t *testing.T) {
	originalFrame := Frame{
		Type:      TypeMeshAlert,
		TTL:       5,
		MessageID: 1048576,
		Payload:   []byte("EMERGENCY ALERT: Water & supplies at Central Square!"),
	}

	// 1. Encode frame
	encodedBytes := Encode(originalFrame)

	// 2. Decode frame
	decodedFrame, err := Decode(encodedBytes)
	if err != nil {
		t.Fatalf("Failed to decode frame: %v", err)
	}

	// 3. Assert fields match
	if decodedFrame.Type != originalFrame.Type {
		t.Errorf("Type mismatch: expected %d, got %d", originalFrame.Type, decodedFrame.Type)
	}
	if decodedFrame.TTL != originalFrame.TTL {
		t.Errorf("TTL mismatch: expected %d, got %d", originalFrame.TTL, decodedFrame.TTL)
	}
	if decodedFrame.MessageID != originalFrame.MessageID {
		t.Errorf("MessageID mismatch: expected %d, got %d", originalFrame.MessageID, decodedFrame.MessageID)
	}
	if !bytes.Equal(decodedFrame.Payload, originalFrame.Payload) {
		t.Errorf("Payload mismatch: expected '%s', got '%s'", string(originalFrame.Payload), string(decodedFrame.Payload))
	}
}

func TestFrameCRCCorruptionDetection(t *testing.T) {
	originalFrame := Frame{
		Type:      TypeNote,
		TTL:       3,
		MessageID: 42,
		Payload:   []byte("Secret Note"),
	}

	encodedBytes := Encode(originalFrame)

	// Corrupt 1 byte in payload (simulate audio noise bitflip)
	encodedBytes[15] ^= 0xFF

	_, err := Decode(encodedBytes)
	if err == nil {
		t.Fatalf("Expected CRC verification error for corrupted payload, got nil")
	}

	if err != ErrCRCMismatch {
		t.Errorf("Expected ErrCRCMismatch, got %v", err)
	}
}
