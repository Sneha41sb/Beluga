package dsp

import (
	"testing"
)

func TestModulateAFSK(t *testing.T) {
	config := AFSKConfig{
		SampleRate: 44100.0,
		MarkFreq:   18500.0,
		SpaceFreq:  19500.0,
		BaudRate:   100,
		Amplitude:  0.8,
	}

	data := []byte("HI") // 2 bytes = 16 bits

	samples := ModulateAFSK(data, config)

	// 1 bit duration = 1/100 = 0.01s -> 441 samples per bit
	// 16 bits * 441 = 7056 samples total
	expectedSamplesPerBit := int(config.SampleRate / float64(config.BaudRate))
	expectedTotalSamples := len(data) * 8 * expectedSamplesPerBit

	if len(samples) != expectedTotalSamples {
		t.Fatalf("expected total samples %d, got %d", expectedTotalSamples, len(samples))
	}
}
