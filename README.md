# 🔊 SonicBeacon & UFTP (Ultrasonic File Transfer Protocol)

> **A 100% Offline, Long-Range Acoustic Mesh Network & Encrypted File Transfer Engine**  
> *Transfer files, documents, emergency alerts, URLs, and Wi-Fi credentials across devices using high-frequency sound waves—without Cellular Data, Wi-Fi Routers, or Internet.*

---

## ❓ 1. WHY (The Problem Statement)

Modern digital file sharing relies heavily on centralized infrastructure (cellular towers, Wi-Fi routers, and cloud servers). However, existing solutions fail in critical scenarios:

1. **Network Blackouts & Disasters:** During natural disasters, protests, or emergency power outages, cell towers and internet services are often shut down or congested.
2. **Access Point (AP) Isolation on Public Wi-Fi:** Coffee shop, hotel, airport, and university Wi-Fi networks block local device-to-device IP traffic (`192.168.x.x`), rendering local network tools (like LocalSend or Snapdrop) useless.
3. **Cross-OS Ecosystem Lock-In:** Apple’s AirDrop is locked to iOS/Mac; Android Quick Share is locked to Android/Windows. Linux has no native ecosystem compatibility.
4. **Air-Gap Security Risks:** Transferring sensitive secrets (2FA recovery keys, SSH keys, passwords, seed phrases) over cloud services (Slack, WhatsApp, Email) leaves traces in cloud logs and clipboard memory.

**SonicBeacon (UFTP)** solves every single one of these problems by using **sound waves in the physical air** as the transmission medium.

---

## 💡 2. WHAT (Project Overview)

**SonicBeacon** is a full-stack, dual-surface application (Go Backend Engine + Progressive Web App) that implements **UFTP (Ultrasonic File Transfer Protocol)**.

### Key Capabilities:
- **Zero-Setup Acoustic Pairing:** No Bluetooth pairing, no typing IP addresses, no cloud account logins.
- **Dual Frequency Profiles:**
  - **`Silent Ultrasonic` (18.5 kHz / 19.5 kHz):** Completely silent to human ears, ideal for quiet indoor rooms (3–5m range).
  - **`Long-Range Chirp` (2.4 kHz / 3.2 kHz):** Short audible digital chime that cuts through outdoor crowd noise and travels up to **50+ meters**.
- **Real File & Data Transfer:** Supports sending real files (images, PDFs, documents, ZIPs), URLs, Wi-Fi credentials, and text notes.
- **Dual Transmission Modes:** Public Broadcast (`0xFFFFFFFF`) vs. Direct Targeted Device (`TargetID`).
- **Receiver Permission Gate:** Prompts the receiving user with an **Accept / Decline** consent modal before downloading any file to disk.

---

## 🏛️ 3. HOW (Technical Architecture & Engineering)

UFTP is engineered as a layered network protocol stack:

```text
┌─────────────────────────────────────────────────────────────┐
│             SonicBeacon Web App & Go Transceiver            │
├─────────────────────────────────────────────────────────────┤
│  pkg/crypto - AES-256-GCM Authenticated Encryption          │
├─────────────────────────────────────────────────────────────┤
│  pkg/fec    - Reed-Solomon Erasure Self-Healing (RS)        │
├─────────────────────────────────────────────────────────────┤
│  pkg/mesh   - Multi-Hop Acoustic Mesh Relay (sync.Mutex)    │
├─────────────────────────────────────────────────────────────┤
│  pkg/frame  - Preamble (0xAA 0x7E), TargetID, CRC32 Checksum │
├─────────────────────────────────────────────────────────────┤
│  pkg/dsp    - AFSK Modulator, Goertzel Detector, Demodulator │
└─────────────────────────────────────────────────────────────┘
```

### Core Engineering Components:

1. **Audio Engine (`pkg/dsp`)**:
   - **Sine Generation (`sine.go`):** Computes pure mathematical float64 audio sample buffers ($x[n] = A \cdot \sin(2\pi f t)$).
   - **AFSK Modulator (`afsk.go`):** Converts binary bits into AFSK tones (18.5 kHz for Bit 1 / 19.5 kHz for Bit 0).
   - **Goertzel Filter (`goertzel.go`):** Single-frequency Discrete Fourier Transform (DFT) detecting tone energy in $O(N)$ time complexity.
   - **Demodulator (`demodulate.go`):** Slides bit windows over incoming audio samples, compares Goertzel power values, and packs bits into bytes via `(val << 1) | bit`.

2. **Protocol Layer (`pkg/frame`)**:
   - **Preamble (`0xAA 0xAA 0xAA 0x7E`):** Aligns bit clocks and detects frame start.
   - **Header Fields:** `Type`, `TTL` (Hop limit), `SenderID`, `TargetID`, `MessageID`, `PayloadLength`.
   - **IEEE 802.3 CRC32 Checksum:** Validates payload integrity and discards noise-corrupted frames.

3. **Multi-Hop Mesh Relay (`pkg/mesh`)**:
   - Implements **Controlled Flooding / Epidemic Routing**.
   - Features `sync.Mutex` thread-safe deduplication cache (`seenMessages map[uint32]bool`) and automatic TTL hop decay (`TTL = TTL - 1`) to prevent audio feedback loops while extending physical coverage across crowds.

4. **Cryptography (`pkg/crypto`)**:
   - **AES-256-GCM Encryption:** Authenticated encryption using SHA-256 derived keys and random 12-byte Nonces to secure private directed transfers.

5. **Forward Error Correction (`pkg/fec`)**:
   - **Reed-Solomon Erasure Coding:** Appends parity shards to payload blocks, allowing the receiver to reconstruct 100% of original data even if up to 30% of audio shards are destroyed by room noise.

6. **Web Audio Transceiver (`web/`)**:
   - Browser Web Audio API (`AudioContext`, `AnalyserNode`) for speaker playback and live FFT spectrum visualizer canvas.
   - Includes Service Worker (`sw.js`) and Web App Manifest (`manifest.json`) for 100% offline Progressive Web App (PWA) installation.

---

## 📍 4. WHERE (Use Cases & Applications)

- **Emergency Blackout Zones:** Communicate news, meeting points, and documents during internet shutdowns or protests.
- **Air-Gapped Cybersecurity:** Transfer 2FA recovery codes, SSH keys, or `.env` files phone-to-laptop without cloud traces.
- **Public & Corporate Wi-Fi:** Transfer files on coffee shop or university networks where router AP Isolation blocks local IP traffic.
- **Zero-Friction Link Sharing:** Share Wi-Fi credentials or slides links to everyone in a classroom or conference room without group chats.

---

## 🚀 5. HOW TO USE (Installation & Execution)

### Prerequisites:
- **Go 1.20+** installed
- **Python 3** (or any static HTTP web server)

---

### Step 1: Run the Go Unit Test Suite

Verify that all DSP, Framing, Mesh, Cryptography, and FEC packages pass unit tests:

```bash
go test -v ./...
```

**Expected Output:**
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

### Step 2: Run the Go CLI Mesh Simulator

Run a multi-hop acoustic mesh relay simulation in your terminal:

```bash
go run cmd/uftp/main.go simulate-mesh
```

---

### Step 3: Launch the SonicBeacon Web Dashboard

**Recommended — Go relay server (required for cross-device transfer):**

```bash
go run cmd/uftp/main.go server 8080
```

This serves the web app **and** the `/api/events` (SSE) + `/api/broadcast` relay
endpoints on the same port. Two or more devices on the same Wi-Fi/LAN opening
`http://<your-computer's-LAN-IP>:8080` in a browser will reliably see each
other's transfers over the network relay, in addition to (or instead of)
acoustic transfer.

> ⚠️ **Do not use a plain static file server** (e.g. `python3 -m http.server`)
> if you need transfers to reach a *different* device. It only serves the
> static HTML/JS/CSS with no backend, so `/api/events` and `/api/broadcast`
> 404, and the receiving device silently never gets anything except acoustic
> audio it can actually demodulate. This was the cause of "sender shows sent,
> receiver never fetches it" — the frame was never reaching the relay at all.
> A static server is fine only for previewing the UI on a single device.

Open **`http://localhost:8080`** in your browser!

#### Using the Web App:
1. **Transmitting:**
   - Set your **User Handle** in the header (e.g. `Node-Alpha`).
   - Select Frequency Profile: **`Silent Ultrasonic`** or **`Long-Range Chirp`**.
   - Select Target Mode: **`Public Broadcast`** or **`Direct Targeted Person`**.
   - Select Payload Mode: **`Message`** or **`Real File`** (Image, PDF, Doc).
   - Click **Broadcast Audio Frame**.

2. **Receiving:**
   - Click **Start Listening** to activate microphone demodulation and the live FFT spectrum canvas.
   - When an incoming transfer arrives, a **Permission Consent Request Banner** displays.
   - Click **Accept & Download File** to save the file directly to your disk!

The header shows a **`RELAY:`** badge (`LAN`, `CLOUD`, `CLOUD (best-effort)`,
or `OFFLINE`) so you can see at a glance whether the network relay is actually
reachable, instead of transfers silently vanishing.

---

## ☁️ 6. Deploying to Vercel

`vercel.json` + `api/messages.js` let you deploy SonicBeacon as a static site
with two serverless relay endpoints (`/api/messages`, `/api/broadcast`), no Go
server required. By default this relay stores messages **in memory**, which
only survives within a single warm serverless instance — on real traffic,
Vercel may route your two devices' requests to different instances, so a
transfer can appear "sent" but never reach the other device.

**For guaranteed delivery**, connect a free Upstash Redis database to the
project (Vercel dashboard → Storage → Marketplace → Upstash → Redis), which
sets `UPSTASH_REDIS_REST_URL` and `UPSTASH_REDIS_REST_TOKEN` automatically —
`api/messages.js` detects them and persists messages there instead. Without
it, the relay still works for casual/low-traffic testing, and the `RELAY:`
badge will show `CLOUD (best-effort)` as a reminder.

Large file transfers (`Real File` payload mode) are capped around 3–4MB on
the cloud relay to stay under Vercel's request body limit — the app warns
before sending an oversized file and recommends the self-hosted Go relay
server (no such limit) for bigger files.

---

## 📜 License
MIT License &bull; Developed as an open-source acoustic networking protocol.
