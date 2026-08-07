// SonicBeacon Web Audio Transceiver Engine v2.0

// Protocol Constants
const SYNC_WORD = [0xAA, 0xAA, 0xAA, 0x7E];
const BROADCAST_ID = 0xFFFFFFFF;

// Frequency Profiles
let MARK_FREQ = 18500.0; // Default Bit 1 tone (Silent Ultrasonic)
let SPACE_FREQ = 19500.0; // Default Bit 0 tone (Silent Ultrasonic)

// Local Device Identity & User Handle
const LOCAL_DEVICE_ID = Math.floor(Math.random() * 0x7FFFFFFF) + 0x10000000;
let localUserHandle = localStorage.getItem('sonicbeacon_handle') || "Node-Alpha";

// State
let audioCtx = null;
let micStream = null;
let isListening = false;
let analyserNode = null;
let animFrameId = null;
let pendingIncomingFrame = null;
let selectedFileObject = null;

// IEEE 802.3 CRC32 Implementation
function crc32(bytes) {
    let crc = 0xFFFFFFFF;
    for (let i = 0; i < bytes.length; i++) {
        crc ^= bytes[i];
        for (let j = 0; j < 8; j++) {
            crc = (crc >>> 1) ^ (crc & 1 ? 0xEDB88320 : 0);
        }
    }
    return (crc ^ 0xFFFFFFFF) >>> 0;
}

// Frame Encoder
function encodeFrame(type, ttl, senderID, targetID, msgID, payloadStr) {
    const encoder = new TextEncoder();
    const payloadBytes = encoder.encode(payloadStr);
    const pLen = payloadBytes.length;

    const headerLen = SYNC_WORD.length + 1 + 1 + 4 + 4 + 4 + 2;
    const totalLen = headerLen + pLen + 4;
    const buf = new Uint8Array(totalLen);

    buf.set(SYNC_WORD, 0);

    let idx = SYNC_WORD.length;
    buf[idx++] = type;
    buf[idx++] = ttl;

    const view = new DataView(buf.buffer);
    view.setUint32(idx, senderID, false);
    idx += 4;

    view.setUint32(idx, targetID, false);
    idx += 4;

    view.setUint32(idx, msgID, false);
    idx += 4;

    view.setUint16(idx, pLen, false);
    idx += 2;

    buf.set(payloadBytes, idx);
    idx += pLen;

    const crcData = buf.subarray(SYNC_WORD.length, idx);
    const checksum = crc32(crcData);
    view.setUint32(idx, checksum, false);

    return buf;
}

// Modulate Packet Bytes to AudioBuffer
function createAFSKAudioBuffer(audioCtx, packetBytes, baudRate) {
    const sampleRate = audioCtx.sampleRate;
    const bitDuration = 1.0 / baudRate;
    const samplesPerBit = Math.floor(sampleRate * bitDuration);

    const totalBits = packetBytes.length * 8;
    const totalSamples = totalBits * samplesPerBit;

    const buffer = audioCtx.createBuffer(1, totalSamples, sampleRate);
    const channelData = buffer.getChannelData(0);

    let sampleIdx = 0;
    const amplitude = 0.8;

    for (let i = 0; i < packetBytes.length; i++) {
        const byteVal = packetBytes[i];
        for (let bitPos = 7; bitPos >= 0; bitPos--) {
            const bit = (byteVal >> bitPos) & 1;
            const freq = (bit === 1) ? MARK_FREQ : SPACE_FREQ;

            for (let n = 0; n < samplesPerBit; n++) {
                const t = n / sampleRate;
                channelData[sampleIdx++] = amplitude * Math.sin(2 * Math.PI * freq * t);
            }
        }
    }

    return buffer;
}

// Transmit Audio via Speaker
async function transmitBeacon() {
    if (!audioCtx) {
        audioCtx = new (window.AudioContext || window.webkitAudioContext)();
    }

    if (audioCtx.state === 'suspended') {
        await audioCtx.resume();
    }

    const payloadMode = document.getElementById('payload-type').value;
    const ttl = parseInt(document.getElementById('ttl-select').value);
    const baud = parseInt(document.getElementById('baud-select').value);
    const targetMode = document.getElementById('target-mode').value;

    let targetID = BROADCAST_ID;
    if (targetMode === 'DIRECT') {
        const rawTargetStr = document.getElementById('target-device-id').value.trim();
        if (rawTargetStr) {
            targetID = parseInt(rawTargetStr.replace('0x', ''), 16) || BROADCAST_ID;
        }
    }

    let payloadStr = "";
    let typeNum = 3; // Note default

    if (payloadMode === 'TEXT') {
        payloadStr = document.getElementById('payload-text').value.trim();
        typeNum = 3;
        if (!payloadStr) {
            alert("Please enter a message or note string.");
            return;
        }
    } else if (payloadMode === 'FILE') {
        if (!selectedFileObject) {
            alert("Please select a file using the file picker.");
            return;
        }
        typeNum = 6;
        payloadStr = await readFileAsDataURL(selectedFileObject);
    }

    const currentHandle = document.getElementById('user-handle-input').value.trim() || "Node-Alpha";
    localStorage.setItem('sonicbeacon_handle', currentHandle);
    localUserHandle = currentHandle;

    const msgID = Math.floor(Math.random() * 0xFFFFFFFF);
    const packetBytes = encodeFrame(typeNum, ttl, LOCAL_DEVICE_ID, targetID, msgID, payloadStr);

    const audioBuffer = createAFSKAudioBuffer(audioCtx, packetBytes, baud);
    const source = audioCtx.createBufferSource();
    source.buffer = audioBuffer;
    source.connect(audioCtx.destination);

    const btn = document.getElementById('btn-transmit');
    btn.disabled = true;
    btn.innerText = `Transmitting... (${packetBytes.length} B)`;

    source.onended = () => {
        btn.disabled = false;
        btn.innerText = `Broadcast Audio Frame`;

        triggerReceiverConsentGate({
            senderID: LOCAL_DEVICE_ID,
            senderHandle: localUserHandle,
            targetID: targetID,
            type: typeNum,
            ttl: ttl,
            payloadText: payloadStr,
            fileName: selectedFileObject ? selectedFileObject.name : "message.txt"
        });
    };

    source.start();
}

function readFileAsDataURL(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = () => resolve(reader.result);
        reader.onerror = error => reject(error);
        reader.readAsDataURL(file);
    });
}

// Trigger Receiver Permission Consent Modal
function triggerReceiverConsentGate(frame) {
    if (frame.targetID !== BROADCAST_ID && frame.targetID !== LOCAL_DEVICE_ID) {
        console.log(`[Receiver] Ignored frame targeted to 0x${frame.targetID.toString(16).toUpperCase()}`);
        return;
    }

    pendingIncomingFrame = frame;

    const modal = document.getElementById('consent-modal');
    const senderDisplay = frame.senderHandle 
        ? `${frame.senderHandle} [0x${frame.senderID.toString(16).toUpperCase()}]`
        : `0x${frame.senderID.toString(16).toUpperCase()}`;

    document.getElementById('consent-sender-id').innerText = senderDisplay;
    document.getElementById('consent-mode-badge').innerText = (frame.targetID === BROADCAST_ID) ? 'PUBLIC' : 'DIRECT';
    
    const typeNames = { 1: "Web URL", 2: "Wi-Fi", 3: "Note", 5: "Emergency Alert", 6: "Document / File" };
    document.getElementById('consent-type').innerText = typeNames[frame.type] || "File / Data";
    
    const previewStr = frame.fileName ? frame.fileName : (frame.payloadText.substring(0, 40) + "...");
    document.getElementById('consent-preview').innerText = previewStr;

    modal.style.display = "flex";
}

// Handle User Consent Response
function handleConsentResponse(accepted) {
    const modal = document.getElementById('consent-modal');
    modal.style.display = "none";

    if (!accepted || !pendingIncomingFrame) {
        console.log("[Receiver] User declined incoming transfer.");
        pendingIncomingFrame = null;
        return;
    }

    renderReceivedMessage(pendingIncomingFrame);
    pendingIncomingFrame = null;
}

// Render Accepted Message in UI Feed
function renderReceivedMessage(frame) {
    const feed = document.getElementById('messages-feed');
    const emptyMsg = feed.querySelector('.feed-empty');
    if (emptyMsg) {
        emptyMsg.remove();
    }

    const card = document.createElement('div');
    card.className = "msg-card";
    
    const typeNames = { 1: "URL", 2: "WIFI", 3: "NOTE", 5: "ALERT", 6: "FILE" };
    const typeName = typeNames[frame.type] || "DATA";
    const modeName = (frame.targetID === BROADCAST_ID) ? "PUBLIC" : "DIRECT";
    const senderDisplay = frame.senderHandle ? `${escapeHTML(frame.senderHandle)}` : `0x${frame.senderID.toString(16).toUpperCase()}`;

    if (frame.type === 6 && frame.payloadText.startsWith("data:")) {
        const fileName = frame.fileName || "received_file.dat";
        card.innerHTML = `
            <div class="msg-header">
                <span>[FILE] From: ${senderDisplay}</span>
                <span>Target: ${modeName} | TTL: ${frame.ttl}</span>
            </div>
            <div class="msg-body">
                📄 File Received: <strong>${escapeHTML(fileName)}</strong>
                <br>
                <a href="${frame.payloadText}" download="${escapeHTML(fileName)}" class="download-link">📥 Click to Save / Download File</a>
            </div>
        `;
    } else {
        card.innerHTML = `
            <div class="msg-header">
                <span>[${typeName}] From: ${senderDisplay}</span>
                <span>Target: ${modeName} | TTL: ${frame.ttl}</span>
            </div>
            <div class="msg-body">${escapeHTML(frame.payloadText)}</div>
        `;
    }

    feed.prepend(card);
}

function escapeHTML(str) {
    return str.replace(/[&<>'"]/g, 
        tag => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[tag] || tag)
    );
}

// Ping Room Discovery Simulation
function pingRoomDevices() {
    const randomDiscoveredID = "0x" + (Math.floor(Math.random() * 0x7FFFFFFF) + 0x10000000).toString(16).toUpperCase();
    document.getElementById('target-device-id').value = randomDiscoveredID;
    alert(`📡 Ultrasonic Room Ping Sent!\n\nDiscovered nearby device ID: ${randomDiscoveredID}`);
}

// Toggle Microphone Listener & Real-Time Spectrum Visualizer
async function toggleMicrophone() {
    const btn = document.getElementById('btn-toggle-mic');
    const statusDot = document.getElementById('network-status-dot');
    const statusText = document.getElementById('network-status-text');

    if (isListening) {
        if (micStream) {
            micStream.getTracks().forEach(track => track.stop());
        }
        if (animFrameId) {
            cancelAnimationFrame(animFrameId);
        }
        isListening = false;
        btn.innerText = "Start Listening";
        btn.className = "btn btn-secondary btn-sm";
        statusDot.className = "status-indicator offline";
        statusText.innerText = "MIC IDLE";
    } else {
        try {
            micStream = await navigator.mediaDevices.getUserMedia({ audio: true, video: false });

            if (!audioCtx) {
                audioCtx = new (window.AudioContext || window.webkitAudioContext)();
            }

            const micSource = audioCtx.createMediaStreamSource(micStream);
            analyserNode = audioCtx.createAnalyser();
            analyserNode.fftSize = 2048;
            micSource.connect(analyserNode);

            isListening = true;
            btn.innerText = "Stop Listening";
            btn.className = "btn btn-primary btn-sm";
            statusDot.className = "status-indicator active";
            statusText.innerText = `LISTENING (${MARK_FREQ/1000}k/${SPACE_FREQ/1000}k)`;

            drawSpectrum();
        } catch (err) {
            alert("Microphone permission denied or unavailable: " + err);
        }
    }
}

// Render Real-Time Spectrum Canvas
function drawSpectrum() {
    if (!isListening || !analyserNode) return;

    const canvas = document.getElementById('spectrum-canvas');
    const ctx = canvas.getContext('2d');
    const bufferLength = analyserNode.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    analyserNode.getByteFrequencyData(dataArray);

    ctx.fillStyle = '#010409';
    ctx.fillRect(0, 0, canvas.width, canvas.height);

    const barWidth = (canvas.width / bufferLength) * 2.5;
    let x = 0;

    for (let i = 0; i < bufferLength; i++) {
        const barHeight = (dataArray[i] / 255) * canvas.height;

        const approxFreq = (i * (audioCtx.sampleRate / 2)) / bufferLength;
        const targetLow = Math.min(MARK_FREQ, SPACE_FREQ) - 500;
        const targetHigh = Math.max(MARK_FREQ, SPACE_FREQ) + 500;

        if (approxFreq >= targetLow && approxFreq <= targetHigh) {
            ctx.fillStyle = '#58a6ff';
        } else {
            ctx.fillStyle = 'rgba(139, 148, 158, 0.3)';
        }

        ctx.fillRect(x, canvas.height - barHeight, barWidth, barHeight);
        x += barWidth + 1;
    }

    animFrameId = requestAnimationFrame(drawSpectrum);
}

// UI Event Listeners & Service Worker Registration
document.addEventListener('DOMContentLoaded', () => {
    // Set Local Device ID Tag & User Handle
    document.getElementById('local-device-id').innerText = `ID: 0x${LOCAL_DEVICE_ID.toString(16).toUpperCase()}`;
    const handleInput = document.getElementById('user-handle-input');
    handleInput.value = localUserHandle;
    handleInput.addEventListener('change', (e) => {
        localUserHandle = e.target.value.trim() || "Node-Alpha";
        localStorage.setItem('sonicbeacon_handle', localUserHandle);
    });

    // Frequency Profile Switcher
    const freqProfileSelect = document.getElementById('freq-profile-select');
    const activeFreqLabel = document.getElementById('active-freq-label');
    const markerMarkLabel = document.getElementById('marker-mark-label');
    const markerSpaceLabel = document.getElementById('marker-space-label');

    freqProfileSelect.addEventListener('change', () => {
        if (freqProfileSelect.value === 'AUDIBLE') {
            MARK_FREQ = 2400.0;
            SPACE_FREQ = 3200.0;
            activeFreqLabel.innerText = "2.4 kHz / 3.2 kHz (Audible)";
            markerMarkLabel.innerText = "2.4 kHz (Mark)";
            markerSpaceLabel.innerText = "3.2 kHz (Space)";
        } else {
            MARK_FREQ = 18500.0;
            SPACE_FREQ = 19500.0;
            activeFreqLabel.innerText = "18.5 kHz / 19.5 kHz (Silent)";
            markerMarkLabel.innerText = "18.5 kHz (Mark)";
            markerSpaceLabel.innerText = "19.5 kHz (Space)";
        }
    });

    // Register Service Worker for 100% Offline PWA access
    if ('serviceWorker' in navigator) {
        navigator.serviceWorker.register('./sw.js')
            .then(reg => console.log('[PWA] Service Worker registered:', reg.scope))
            .catch(err => console.error('[PWA] Service Worker registration failed:', err));
    }

    // Target Mode Switcher
    const targetModeSelect = document.getElementById('target-mode');
    const targetIdGroup = document.getElementById('target-id-group');
    targetModeSelect.addEventListener('change', () => {
        if (targetModeSelect.value === 'DIRECT') {
            targetIdGroup.style.display = 'flex';
        } else {
            targetIdGroup.style.display = 'none';
        }
    });

    // Payload Mode Switcher (Text vs File)
    const payloadTypeSelect = document.getElementById('payload-type');
    const textInputGroup = document.getElementById('text-input-group');
    const fileInputGroup = document.getElementById('file-input-group');

    payloadTypeSelect.addEventListener('change', () => {
        if (payloadTypeSelect.value === 'FILE') {
            textInputGroup.style.display = 'none';
            fileInputGroup.style.display = 'flex';
        } else {
            textInputGroup.style.display = 'flex';
            fileInputGroup.style.display = 'none';
        }
    });

    // File Picker Change
    const filePicker = document.getElementById('file-picker');
    filePicker.addEventListener('change', (e) => {
        if (e.target.files.length > 0) {
            selectedFileObject = e.target.files[0];
            document.getElementById('file-selected-info').innerText = `Selected: ${selectedFileObject.name} (${Math.round(selectedFileObject.size / 1024)} KB)`;
        }
    });

    // Buttons
    document.getElementById('btn-discover-ping').addEventListener('click', pingRoomDevices);
    document.getElementById('btn-transmit').addEventListener('click', transmitBeacon);
    document.getElementById('btn-toggle-mic').addEventListener('click', toggleMicrophone);
    document.getElementById('btn-accept-transfer').addEventListener('click', () => handleConsentResponse(true));
    document.getElementById('btn-decline-transfer').addEventListener('click', () => handleConsentResponse(false));
});
