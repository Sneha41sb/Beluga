package dsp

import (
	"math"
)

// GoertzelPower calculates the energy power level of a specific target frequency 
// within a slice of audio samples using the Goertzel algorithm.
func GoertzelPower(samples []float64, targetFreq float64, sampleRate float64) float64 {
	N := float64(len(samples))
	k := math.Round(N * targetFreq / sampleRate)
	omega := (2.0 * math.Pi * k) / N
	coeff := 2.0 * math.Cos(omega)

	var q0, q1, q2 float64

	for _, sample := range samples {
		q0 = coeff*q1 - q2 + sample
		q2 = q1
		q1 = q0
	}

	power := (q1 * q1) + (q2 * q2) - (q1 * q2 * coeff)
	return power
}
