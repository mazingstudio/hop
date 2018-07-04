package hop

// Config represents a WorkQueue configuration.
type Config struct {
	// ExchangeName is used when naming the underlying exchange for the work
	// queue. Please DO NOT use a pre-existing exchange name, as this would not
	// guarantee that the pre-existing exchange has the necessary properties
	// to provide work queue semantics.
	ExchangeName string
	// Persist indicates whether the underlying queues and messages should be
	// saved to disk to survive server restarts and crashes
	Persistent bool
}

var DefaultConfig = &Config{
	ExchangeName: "hop.exchange",
	Persistent:   false,
}
