package ctxkey

// ContextKey represents keys used in context value
// store specific for single http request.
type ContextKey uint

const (
	// DebugKey represents key used to store
	// information about debug mode.
	DebugKey ContextKey = iota
)
