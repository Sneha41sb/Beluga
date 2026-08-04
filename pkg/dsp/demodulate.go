package dsp

// DemodulateAFSK converts raw audio float64 samples back into raw bytes.
func DemodulateAFSK(samples []float64, config AFSKConfig) []byte {
	samplesPerBit := int(config.SampleRate / float64(config.BaudRate))
	var decodedBytes []byte

	var currentByte byte
	bitCount := 0

	for i := 0; i+samplesPerBit <= len(samples); i += samplesPerBit {
		bitSamples := samples[i : i+samplesPerBit]

		markPower := GoertzelPower(bitSamples, config.MarkFreq, config.SampleRate)
		spacePower := GoertzelPower(bitSamples, config.SpaceFreq, config.SampleRate)

		var bit byte
		if markPower > spacePower {
			bit = 1
		} else {
			bit = 0
		}

		currentByte = (currentByte << 1) | bit
		bitCount++

		if bitCount == 8 {
			decodedBytes = append(decodedBytes, currentByte)
			currentByte = 0
			bitCount = 0
		}
	}

	return decodedBytes
}