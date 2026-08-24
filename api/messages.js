let messageStore = [];
const RETENTION_MS = 60000; // Keep messages for 60 seconds

module.exports = (req, res) => {
    res.setHeader('Access-Control-Allow-Origin', '*');
    res.setHeader('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.setHeader('Access-Control-Allow-Headers', 'Content-Type');

    if (req.method === 'OPTIONS') {
        return res.status(200).end();
    }

    const now = Date.now();
    messageStore = messageStore.filter(m => (now - m.timestamp) < RETENTION_MS);

    if (req.method === 'POST') {
        let body = req.body;
        if (typeof body === 'string') {
            try { body = JSON.parse(body); } catch(e) {}
        }
        if (body && body.senderID) {
            body.timestamp = now;
            messageStore.push(body);
            return res.status(200).json({ status: 'ok', msgID: body.msgID });
        }
        return res.status(400).json({ error: 'Invalid frame payload' });
    }

    if (req.method === 'GET') {
        const since = parseInt(req.query.since || '0', 10);
        const newMessages = messageStore.filter(m => m.timestamp > since);
        return res.status(200).json({
            now: now,
            messages: newMessages
        });
    }

    return res.status(405).json({ error: 'Method not allowed' });
};
