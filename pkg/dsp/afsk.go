package dsp

// AFSKConfig holds parameters for Audio Frequency-Shift Keying.
type AFSKConfig struct {
	SampleRate float64 // e.g. 44100.0
	MarkFreq   float64 // e.g. 18500.0 (Bit 1)
	SpaceFreq  float64 // e.g. 19500.0 (Bit 0)
	BaudRate   int     // e.g. 100 bits/sec
	Amplitude  float64 // e.g. 0.8
}

// ModulateAFSK converts raw byte data into continuous AFSK audio float64 samples.
func ModulateAFSK(data []byte, config AFSKConfig) []float64 {
	bitDuration := 1.0 / float64(config.BaudRate)
	var audioStream []float64

	for _, b := range data {
		for i := 7; i >= 0; i-- {
			bit := (b >> i) & 1

			freq := config.SpaceFreq
			if bit == 1 {
				freq = config.MarkFreq
			}

			bitSamples := GenerateSineWave(freq, bitDuration, config.SampleRate, config.Amplitude)
			audioStream = append(audioStream, bitSamples...)
		}
	}

	return audioStream
}
