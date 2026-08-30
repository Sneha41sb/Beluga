// Vercel Serverless Relay Endpoint
//
// IMPORTANT: Vercel Serverless Functions do not guarantee that repeated
// invocations hit the same process. A plain in-memory array (the old
// implementation) is only visible to the single warm instance that wrote
// it -- a second device polling this endpoint can easily land on a
// different instance and see an empty store, so it silently never
// receives anything that was "sent". That was the root cause of
// transfers not showing up on a second device when this project is
// deployed to Vercel.
//
// Fix: if UPSTASH_REDIS_REST_URL / UPSTASH_REDIS_REST_TOKEN are present
// (a free Upstash Redis database connected via the Vercel Marketplace),
// messages are persisted there so every instance sees the same data.
// Without those env vars we fall back to the old in-memory behavior,
// which still works for local `vercel dev` / low-traffic single-instance
// testing, but is best-effort only on real multi-instance deployments.

const RETENTION_MS = 60000; // Keep messages for 60 seconds
const MAX_STORED = 200; // Hard cap so the in-memory fallback can't grow unbounded
const MAX_PAYLOAD_BYTES = 4 * 1024 * 1024; // Stay under Vercel's ~4.5MB request body limit

const REDIS_URL = process.env.UPSTASH_REDIS_REST_URL;
const REDIS_TOKEN = process.env.UPSTASH_REDIS_REST_TOKEN;
const REDIS_KEY = 'sonicbeacon:messages';

// In-memory fallback store (persists only within a single warm instance).
global.__sonicbeaconMemoryStore = global.__sonicbeaconMemoryStore || [];

async function redisCommand(command) {
    const res = await fetch(REDIS_URL, {
        method: 'POST',
        headers: {
            Authorization: `Bearer ${REDIS_TOKEN}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(command)
    });
    if (!res.ok) {
        throw new Error(`Upstash error: HTTP ${res.status}`);
    }
    return res.json();
}

async function loadMessages() {
    if (REDIS_URL && REDIS_TOKEN) {
        const result = await redisCommand(['LRANGE', REDIS_KEY, '0', String(MAX_STORED - 1)]);
        const raw = result && result.result ? result.result : [];
        return raw.map((s) => {
            try { return JSON.parse(s); } catch (e) { return null; }
        }).filter(Boolean);
    }
    return global.__sonicbeaconMemoryStore;
}

async function saveMessage(msg) {
    if (REDIS_URL && REDIS_TOKEN) {
        await redisCommand(['LPUSH', REDIS_KEY, JSON.stringify(msg)]);
        await redisCommand(['LTRIM', REDIS_KEY, '0', String(MAX_STORED - 1)]);
        await redisCommand(['EXPIRE', REDIS_KEY, '120']);
        return;
    }
    global.__sonicbeaconMemoryStore.push(msg);
    if (global.__sonicbeaconMemoryStore.length > MAX_STORED) {
        global.__sonicbeaconMemoryStore.splice(0, global.__sonicbeaconMemoryStore.length - MAX_STORED);
    }
}

module.exports = async (req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

    if (req.method === 'OPTIONS') {
        return res.status(200).end();
    }

    const now = Date.now();

    if (req.method === 'POST') {
        let body = req.body;
        if (typeof body === 'string') {
            try { body = JSON.parse(body); } catch (e) { body = null; }
        }

        if (!body || !body.senderID) {
            return res.status(400).json({ error: 'Invalid frame payload' });
        }

        const approxBytes = Buffer.byteLength(JSON.stringify(body), 'utf8');
        if (approxBytes > MAX_PAYLOAD_BYTES) {
            return res.status(413).json({
                error: 'Payload too large for the cloud relay',
                detail: 'This relay endpoint is capped well below the platform request-size limit. Large files should be sent over the self-hosted Go relay server (`go run cmd/uftp/main.go server`) on the same Wi-Fi/LAN instead.'
            });
        }

        body.timestamp = now;

        try {
            await saveMessage(body);
        } catch (e) {
            return res.status(502).json({ error: 'Relay storage unavailable', detail: String(e) });
        }

        return res.status(200).json({ status: 'ok', msgID: body.msgID });
    }

    if (req.method === 'GET') {
        const since = parseInt(req.query.since || '0', 10);

        let messages;
        try {
            messages = await loadMessages();
        } catch (e) {
            return res.status(502).json({ error: 'Relay storage unavailable', detail: String(e) });
        }

        const freshMessages = messages.filter((m) => (now - m.timestamp) < RETENTION_MS && m.timestamp > since);

        if (!REDIS_URL) {
            global.__sonicbeaconMemoryStore = global.__sonicbeaconMemoryStore.filter((m) => (now - m.timestamp) < RETENTION_MS);
        }

        return res.status(200).json({
            now: now,
            persistent: Boolean(REDIS_URL && REDIS_TOKEN),
            messages: freshMessages
        });
    }

    return res.status(405).json({ error: 'Method not allowed' });
};
