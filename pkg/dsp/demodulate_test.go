package dsp

import (
	"testing"
)

func TestAFSKLoopback(t *testing.T) {
	config := AFSKConfig{
		SampleRate: 44100.0,
		MarkFreq:   18500.0,
		SpaceFreq:  19500.0,
		BaudRate:   100,
		Amplitude:  0.8,
	}

	originalMessage := "Hello, Ultrasonic World!"
	data := []byte(originalMessage)

	// 1. Modulate bytes into audio samples
	audioSamples := ModulateAFSK(data, config)

	// 2. Demodulate audio samples back into bytes
	decodedBytes := DemodulateAFSK(audioSamples, config)
	decodedMessage := string(decodedBytes)

	t.Logf("Original:  %s", originalMessage)
	t.Logf("Decoded:   %s", decodedMessage)

	// 3. Verify exact match
	if originalMessage != decodedMessage {
		t.Fatalf("Loopback failed! Expected '%s', got '%s'", originalMessage, decodedMessage)
	}
}
