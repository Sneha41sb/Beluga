package dsp

import (
	"math"
)

// GenerateSineWave produces a slice of float64 audio samples for a given frequency,
// duration (in seconds), sample rate, and amplitude.
func GenerateSineWave(freq float64, durationSec float64, sampleRate float64, amplitude float64) []float64 {
	numSamples := int(durationSec * sampleRate)
	samples := make([]float64, numSamples)

	for i := 0; i < numSamples; i++ {
		t := float64(i) / sampleRate
		samples[i] = amplitude * math.Sin(2 * math.Pi * freq * t)
	}

	return samples
}
