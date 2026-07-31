package dsp

import (
	"math"
	"testing"
)

func TestGenerateSineWave(t *testing.T) {
	freq := 18500.0      // 18.5 kHz ultrasonic carrier
	durationSec := 1.0   // 1 second
	sampleRate := 44100.0
	amplitude := 0.8

	samples := GenerateSineWave(freq, durationSec, sampleRate, amplitude)

	// Test 1: Check slice length
	expectedLength := int(durationSec * sampleRate)
	if len(samples) != expectedLength {
		t.Fatalf("expected length %d, got %d", expectedLength, len(samples))
	}

	// Test 2: Check amplitude bounds
	for i, sample := range samples {
		if math.Abs(sample) > amplitude+1e-9 {
			t.Errorf("sample at index %d exceeded amplitude limit: %f > %f", i, math.Abs(sample), amplitude)
		}
	}

	// Test 3: Check first sample at t=0 is 0 (sin(0) = 0)
	if math.Abs(samples[0]) > 1e-9 {
		t.Errorf("expected sample[0] to be ~0, got %f", samples[0])
	}
}
