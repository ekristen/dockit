package common

type ContextKey string

// Common Constants, mostly for annotations
const ContextReqIDKey ContextKey = "reqid"

// ContextKeyDb is used to set context value for db
const ContextKeyDb ContextKey = "db"

// ContextKeyNode is used to set context value for node/snowflake
const ContextKeyNode ContextKey = "node"

// ContextKeyWebsocket is used to set context value for passing websocket
const ContextKeyWebsocket ContextKey = "websocket"
