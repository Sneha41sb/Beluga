# 🔊 SonicBeacon & UFTP (Ultrasonic File Transfer Protocol)

> **A 100% Offline, Long-Range Acoustic Mesh Network & Encryption Protocol**  
> *Transfer files, emergency alerts, URLs, and credentials across devices using sound waves—without Wi-Fi, Cellular Data, or Internet.*

---

## 🌟 Features & Highlights

- **🔊 Acoustic Audio Transceiver (`pkg/dsp`)**: Modulates binary data into high-frequency ultrasonic audio tones (18.5 kHz Mark / 19.5 kHz Space) using Audio Frequency-Shift Keying (AFSK) and $O(N)$ Goertzel single-frequency DFT power detection.
- **📦 Reliable Packet Framing (`pkg/frame`)**: Wraps payloads with Sync Preambles (`0xAA 0x7E`), `SenderID`, `TargetID` (Public Broadcast vs. Private Targeted), hop counts, and IEEE 802.3 CRC32 checksum error detection.
- **🌐 Long-Range Multi-Hop Mesh Relay (`pkg/mesh`)**: Bypasses acoustic range limits through peer-to-peer relaying across devices with thread-safe `sync.Mutex` message deduplication and TTL decay safeguards.
- **🛡️ AES-256-GCM Payload Encryption (`pkg/crypto`)**: Authenticated symmetric encryption to secure private targeted file transfers against eavesdropping.
- **🏥 Reed-Solomon Forward Error Correction (`pkg/fec`)**: Self-healing erasure coding that reconstructs corrupted audio data even if up to 30% of audio shards are destroyed by room noise.
- **💻 Interactive Command-Line Interface (`cmd/uftp`)**: CLI tool for packet sending and multi-hop acoustic mesh propagation testing.
- **🌐 Web Audio Progressive Web App (`web/`)**: Zero-install web dashboard with real file uploads, FFT spectrum canvas, receiver permission consent gate, and Service Worker (`sw.js`) for 100% offline access.

---

## 🏛️ System Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│                 SonicBeacon Web App & CLI                   │
├─────────────────────────────────────────────────────────────┤
│  pkg/crypto - AES-256-GCM Payload Encryption                │
├─────────────────────────────────────────────────────────────┤
│  pkg/fec    - Reed-Solomon Forward Error Correction (RS)    │
├─────────────────────────────────────────────────────────────┤
│  pkg/mesh   - Multi-Hop Acoustic Mesh Relay (sync.Mutex)    │
├─────────────────────────────────────────────────────────────┤
│  pkg/frame  - Preamble (0xAA 0x7E), TargetID, CRC32 Checksum │
├─────────────────────────────────────────────────────────────┤
│  pkg/dsp    - AFSK Modulator, Goertzel Detector, Demodulator │
└─────────────────────────────────────────────────────────────┘
```

---

## 🧪 Unit Test Suite

Run the full Go test suite across all packages:

```bash
go test -v ./...
```

### Output:
```text
=== RUN   TestAESGCMEncryptDecrypt
--- PASS: TestAESGCMEncryptDecrypt (0.00s)
=== RUN   TestAFSKLoopback
--- PASS: TestAFSKLoopback (0.01s)
=== RUN   TestGoertzelPower
--- PASS: TestGoertzelPower (0.00s)
=== RUN   TestReedSolomonSelfHealing
--- PASS: TestReedSolomonSelfHealing (0.00s)
=== RUN   TestFrameEncodeDecode
--- PASS: TestFrameEncodeDecode (0.00s)
=== RUN   TestMeshMultiHopPropagation
--- PASS: TestMeshMultiHopPropagation (0.00s)
PASS
```

---

## 🚀 Running the Web Dashboard

Start the local HTTP server:

```bash
python3 -m http.server 8080 --directory web
```

Open `http://localhost:8080` in your web browser!
