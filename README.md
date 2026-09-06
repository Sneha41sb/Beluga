# SonicBeacon & UFTP (Ultrasonic File Transfer Protocol)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](#license)
[![Go Version](https://img.shields.io/badge/Go-1.20%2B-00ADD8?logo=go)](#installation--execution)
[![PWA](https://img.shields.io/badge/PWA-Offline%20Ready-5A0FC8)](#web-audio-transceiver)

> **A 100% offline, long-range acoustic mesh network and encrypted file transfer engine.**
> Transfer files, documents, emergency alerts, URLs, and Wi-Fi credentials between devices using high-frequency sound waves — no cellular data, Wi-Fi routers, or internet connection required.

---

## Table of Contents

- [Problem Statement](#problem-statement)
- [Overview](#overview)
- [Key Capabilities](#key-capabilities)
- [Architecture](#architecture)
- [Core Engineering Components](#core-engineering-components)
- [Use Cases](#use-cases)
- [Installation & Execution](#installation--execution)
- [Using the Web App](#using-the-web-app)
- [License](#license)

---

## Problem Statement

Modern digital file sharing relies heavily on centralized infrastructure — cellular towers, Wi-Fi routers, and cloud servers. This infrastructure fails in several common, high-stakes scenarios:

1. **Network Blackouts & Disasters** — Cell towers and internet services are frequently shut down or overwhelmed during natural disasters, protests, or power outages.
2. **Access Point (AP) Isolation on Public Wi-Fi** — Coffee shops, hotels, airports, and universities routinely block device-to-device IP traffic (`192.168.x.x`), rendering local-network tools like LocalSend or Snapdrop unusable.
3. **Cross-OS Ecosystem Lock-In** — AirDrop is restricted to iOS/macOS, and Quick Share to Android/Windows. Linux has no native equivalent.
4. **Air-Gap Security Risks** — Transferring sensitive secrets (2FA recovery keys, SSH keys, passwords, seed phrases) over cloud services such as Slack, WhatsApp, or email leaves traces in cloud logs and clipboard memory.

**SonicBeacon (UFTP)** addresses all four problems by using sound waves traveling through open air as the transmission medium.

---

## Overview

SonicBeacon is a full-stack, dual-surface application — a Go backend engine paired with a Progressive Web App — that implements **UFTP (Ultrasonic File Transfer Protocol)**.

## Key Capabilities

- **Zero-Setup Acoustic Pairing** — No Bluetooth pairing, no manual IP entry, no cloud account required.
- **Dual Frequency Profiles:**
  - **Silent Ultrasonic (18.5 kHz / 19.5 kHz)** — Inaudible to the human ear; ideal for quiet indoor environments (3–5 m range).
  - **Long-Range Chirp (2.4 kHz / 3.2 kHz)** — A short, audible digital chime that cuts through outdoor crowd noise and travels 50+ meters.
- **Real File & Data Transfer** — Supports files (images, PDFs, documents, ZIPs), URLs, Wi-Fi credentials, and plain-text notes.
- **Dual Transmission Modes** — Public Broadcast (`0xFFFFFFFF`) or Direct Targeted Device (`TargetID`).
- **Receiver Permission Gate** — Every incoming file requires explicit Accept/Decline consent before it touches disk.

---

## Architecture

UFTP is engineered as a layered network protocol stack:

```text
┌─────────────────────────────────────────────────────────────┐
│             SonicBeacon Web App & Go Transceiver             │
├─────────────────────────────────────────────────────────────┤
│  pkg/crypto - AES-256-GCM Authenticated Encryption           │
├─────────────────────────────────────────────────────────────┤
│  pkg/fec    - Reed-Solomon Erasure Self-Healing (RS)         │
├─────────────────────────────────────────────────────────────┤
│  pkg/mesh   - Multi-Hop Acoustic Mesh Relay (sync.Mutex)      │
├─────────────────────────────────────────────────────────────┤
│  pkg/frame  - Preamble (0xAA 0x7E), TargetID, CRC32 Checksum  │
├─────────────────────────────────────────────────────────────┤
│  pkg/dsp    - AFSK Modulator, Goertzel Detector, Demodulator  │
└─────────────────────────────────────────────────────────────┘
```

## Core Engineering Components

1. **Audio Engine (`pkg/dsp`)**
   - **Sine Generation (`sine.go`)** — Computes pure mathematical float64 audio sample buffers: `x[n] = A · sin(2πft)`.
   - **AFSK Modulator (`afsk.go`)** — Converts binary bits into AFSK tones (18.5 kHz for bit 1, 19.5 kHz for bit 0).
   - **Goertzel Filter (`goertzel.go`)** — A single-frequency Discrete Fourier Transform that detects tone energy in O(N) time.
   - **Demodulator (`demodulate.go`)** — Slides bit windows over incoming audio samples, compares Goertzel power values, and packs bits into bytes via `(val << 1) | bit`.

2. **Protocol Layer (`pkg/frame`)**
   - **Preamble (`0xAA 0xAA 0xAA 0x7E`)** — Aligns bit clocks and detects frame start.
   - **Header Fields** — `Type`, `TTL` (hop limit), `SenderID`, `TargetID`, `MessageID`, `PayloadLength`.
   - **IEEE 802.3 CRC32 Checksum** — Validates payload integrity and discards noise-corrupted frames.

3. **Multi-Hop Mesh Relay (`pkg/mesh`)**
   - Implements controlled flooding / epidemic routing.
   - Uses a thread-safe (`sync.Mutex`) deduplication cache (`seenMessages map[uint32]bool`) with automatic TTL hop decay (`TTL = TTL - 1`) to prevent audio feedback loops while extending physical coverage across crowds.

4. **Cryptography (`pkg/crypto`)**
   - **AES-256-GCM Encryption** — Authenticated encryption using SHA-256-derived keys and random 12-byte nonces to secure directed transfers.

5. **Forward Error Correction (`pkg/fec`)**
   - **Reed-Solomon Erasure Coding** — Appends parity shards to payload blocks, allowing full data reconstruction even if up to 30% of audio shards are lost to room noise.

6. **Web Audio Transceiver (`web/`)**
   - Uses the browser Web Audio API (`AudioContext`, `AnalyserNode`) for speaker playback and a live FFT spectrum visualizer canvas.
   - Ships with a Service Worker (`sw.js`) and Web App Manifest (`manifest.json`) for fully offline Progressive Web App installation.

---

## Use Cases

- **Emergency Blackout Zones** — Share news, meeting points, and documents during internet shutdowns or protests.
- **Air-Gapped Cybersecurity** — Move 2FA recovery codes, SSH keys, or `.env` files phone-to-laptop without cloud traces.
- **Public & Corporate Wi-Fi** — Transfer files on coffee shop or university networks where AP isolation blocks local IP traffic.
- **Zero-Friction Link Sharing** — Distribute Wi-Fi credentials or slide links to an entire classroom or conference room without group chats.

---

## Installation & Execution

### Prerequisites

- **Go 1.20+**
- **Python 3** (or any static HTTP web server, for local preview only)

### Step 1 — Run the Go Unit Test Suite

Verify that all DSP, framing, mesh, cryptography, and FEC packages pass their unit tests:

```bash
go test -v ./...
```

**Expected output:**

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

### Step 2 — Run the Go CLI Mesh Simulator

```bash
go run cmd/uftp/main.go simulate-mesh
```

This runs a multi-hop acoustic mesh relay simulation directly in your terminal.

### Step 3 — Launch the SonicBeacon Web Dashboard

**Recommended — Go relay server (required for cross-device transfer):**

```bash
go run cmd/uftp/main.go server 8080
```

This serves the web app **and** the `/api/events` (SSE) and `/api/broadcast` relay endpoints on the same port. Any two devices on the same Wi-Fi/LAN opening `http://<your-computer's-LAN-IP>:8080` will reliably see each other's transfers over the network relay, in addition to (or instead of) acoustic transfer.

> **Do not use a plain static file server** (e.g. `python3 -m http.server`) if you need transfers to reach a *different* device. A static server only serves the HTML/JS/CSS with no backend, so `/api/events` and `/api/broadcast` return 404s, and the receiving device never gets anything except acoustic audio it can demodulate directly. This was the root cause of "sender shows sent, receiver never fetches it" — the frame never reached the relay. A static server is fine for previewing the UI on a single device only.

Open **`http://localhost:8080`** in your browser.

---

## Using the Web App

### Transmitting

1. Set your **User Handle** in the header (e.g. `Node-Alpha`).
2. Select a **Frequency Profile** — `Silent Ultrasonic` or `Long-Range Chirp`.
3. Select a **Target Mode** — `Public Broadcast` or `Direct Targeted Person`.
4. Select a **Payload Mode** — `Message` or `Real File` (image, PDF, document).
5. Click **Broadcast Audio Frame**.

### Receiving

1. Click **Start Listening** to activate microphone demodulation and the live FFT spectrum canvas.
2. When an incoming transfer arrives, a **Permission Consent Request** banner appears.
3. Click **Accept & Download File** to save the file directly to disk.

The header displays a **`RELAY:`** badge (`LAN`, `CLOUD`, `CLOUD (best-effort)`, or `OFFLINE`) so you can see at a glance whether the network relay is reachable, instead of transfers silently failing.

---

## License

MIT License — developed as an open-source acoustic networking protocol.
