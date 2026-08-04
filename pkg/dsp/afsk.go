package dsp

type AFSKConfig struct {
	SampleRate float64
	MarkFreq   float64
	SpaceFreq  float64
	BaudRate   int
	Amplitude  float64
}

// ModulateAFSK turns raw byte data into continuous AFSK audio float64 samples.
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