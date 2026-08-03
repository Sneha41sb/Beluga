package dsp

import (
	"testing"
)

func TestGoertzelPower(t *testing.T) {
	sampleRate := 44100.0
	markFreq := 18500.0  // Bit 1 tone
	spaceFreq := 19500.0 // Bit 0 tone

	// Generate 1 bit duration (100 Baud -> 0.01 seconds -> 441 samples) of 18.5 kHz sine wave
	bitSamples := GenerateSineWave(markFreq, 0.01, sampleRate, 0.8)

	// Measure power at 18.5 kHz vs 19.5 kHz
	powerMark := GoertzelPower(bitSamples, markFreq, sampleRate)
	powerSpace := GoertzelPower(bitSamples, spaceFreq, sampleRate)

	t.Logf("18.5 kHz Signal -> Power at 18.5kHz: %.2f | Power at 19.5kHz: %.2f", powerMark, powerSpace)

	// The power at 18.5 kHz should be significantly higher than at 19.5 kHz
	if powerMark <= powerSpace*10.0 {
		t.Fatalf("Expected power at mark freq (%.2f) to be significantly higher than space freq (%.2f)", powerMark, powerSpace)
	}
}
