package fec

import (
	"bytes"
	"errors"
	"github.com/klauspost/reedsolomon"
)

type FECConfig struct {
	DataShards   int // e.g. 4 data blocks
	ParityShards int // e.g. 2 parity blocks (can heal up to 2 missing shards)
}

type FEC struct {
	config  FECConfig
	encoder reedsolomon.Encoder
}

// NewFEC initializes a new Reed-Solomon FEC encoder/decoder
func NewFEC(config FECConfig) (*FEC, error) {
	enc, err := reedsolomon.New(config.DataShards, config.ParityShards)
	if err != nil {
		return nil, err
	}
	return &FEC{config: config, encoder: enc}, nil
}

// Encode splits data into data shards and generates parity shards
func (f *FEC) Encode(data []byte) ([][]byte, error) {
	shards, err := f.encoder.Split(data)
	if err != nil {
		return nil, err
	}

	if err := f.encoder.Encode(shards); err != nil {
		return nil, err
	}

	return shards, nil
}

// Reconstruct auto-heals missing/nil shards and returns original data
func (f *FEC) Reconstruct(shards [][]byte, originalLen int) ([]byte, error) {
	if err := f.encoder.Reconstruct(shards); err != nil {
		return nil, err
	}

	ok, err := f.encoder.Verify(shards)
	if !ok || err != nil {
		return nil, errors.New("fec verification failed after reconstruction")
	}

	var buf bytes.Buffer
	err = f.encoder.Join(&buf, shards, originalLen)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
