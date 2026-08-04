package frame

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
)

// Payload Types
const (
	TypeURL       byte = 1
	TypeWiFi      byte = 2
	TypeNote      byte = 3
	TypeCode      byte = 4
	TypeMeshAlert byte = 5
)

// SyncWord is our 4-byte preamble for frame synchronization: 0xAA 0xAA 0xAA 0x7E
var SyncWord = []byte{0xAA, 0xAA, 0xAA, 0x7E}

// Frame represents a structured UFTP network packet
type Frame struct {
	Type      byte   // Payload type indicator
	TTL       byte   // Time-To-Live hop counter for mesh relay
	MessageID uint32 // Unique message identifier for deduplication
	Payload   []byte // Raw payload data
}

var (
	ErrSyncWordNotFound = errors.New("preamble sync word not found")
	ErrTruncatedFrame   = errors.New("frame is truncated or incomplete")
	ErrCRCMismatch      = errors.New("crc32 checksum verification failed")
)

// Encode packs a Frame into a binary byte slice ready for AFSK modulation.
// Format: SyncWord (4B) | Type (1B) | TTL (1B) | MessageID (4B) | Length (2B) | Payload (NB) | CRC32 (4B)
func Encode(f Frame) []byte {
	payloadLen := uint16(len(f.Payload))
	headerLen := len(SyncWord) + 1 + 1 + 4 + 2
	totalLen := headerLen + int(payloadLen) + 4 // +4 for CRC32

	buf := make([]byte, totalLen)

	// 1. Write SyncWord
	copy(buf[0:4], SyncWord)

	// 2. Write Header (Type, TTL, MessageID, Length)
	buf[4] = f.Type
	buf[5] = f.TTL
	binary.BigEndian.PutUint32(buf[6:10], f.MessageID)
	binary.BigEndian.PutUint16(buf[10:12], payloadLen)

	// 3. Write Payload
	copy(buf[12:12+int(payloadLen)], f.Payload)

	// 4. Calculate CRC32 over Header (excluding SyncWord) + Payload
	crcData := buf[4 : 12+int(payloadLen)]
	checksum := crc32.ChecksumIEEE(crcData)

	// 5. Append CRC32 at the end
	binary.BigEndian.PutUint32(buf[12+int(payloadLen):], checksum)

	return buf
}

// Decode extracts and validates a Frame from raw bytes.
func Decode(data []byte) (*Frame, error) {
	if len(data) < len(SyncWord)+1+1+4+2+4 {
		return nil, ErrTruncatedFrame
	}

	// 1. Locate SyncWord in stream
	syncIdx := -1
	for i := 0; i <= len(data)-len(SyncWord); i++ {
		if data[i] == SyncWord[0] && data[i+1] == SyncWord[1] &&
			data[i+2] == SyncWord[2] && data[i+3] == SyncWord[3] {
			syncIdx = i
			break
		}
	}

	if syncIdx == -1 {
		return nil, ErrSyncWordNotFound
	}

	frameStart := syncIdx
	headerStart := frameStart + len(SyncWord)

	if len(data) < headerStart+1+1+4+2+4 {
		return nil, ErrTruncatedFrame
	}

	pType := data[headerStart]
	ttl := data[headerStart+1]
	msgID := binary.BigEndian.Uint32(data[headerStart+2 : headerStart+6])
	pLen := binary.BigEndian.Uint16(data[headerStart+6 : headerStart+8])

	payloadStart := headerStart + 8
	payloadEnd := payloadStart + int(pLen)
	crcEnd := payloadEnd + 4

	if len(data) < crcEnd {
		return nil, ErrTruncatedFrame
	}

	payload := make([]byte, pLen)
	copy(payload, data[payloadStart:payloadEnd])

	// Verify CRC32
	crcData := data[headerStart:payloadEnd]
	expectedCRC := binary.BigEndian.Uint32(data[payloadEnd:crcEnd])
	actualCRC := crc32.ChecksumIEEE(crcData)

	if expectedCRC != actualCRC {
		return nil, ErrCRCMismatch
	}

	return &Frame{
		Type:      pType,
		TTL:       ttl,
		MessageID: msgID,
		Payload:   payload,
	}, nil
}
