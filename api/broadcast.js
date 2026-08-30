// Alias of /api/messages for the client's `/api/broadcast` POST calls (the
// self-hosted Go server exposes a separate SSE-push /api/broadcast route;
// this file makes the same client code work unmodified when deployed to
// Vercel, instead of silently 404ing).
module.exports = require('./messages.js');
