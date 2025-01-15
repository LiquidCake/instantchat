package util

const ServerStatusOnline = "online"
const ServerStatusShuttingDown = "shutting_down"
const ServerStatusRestarting = "restarting"

// main limit is anyway determined by websocket message size limit
const MaxMessageLength = 10000

const HTTP_STATUS_UNAUTHORIZED int = 401
