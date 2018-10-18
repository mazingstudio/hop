package hop

import (
	"time"

	"github.com/streadway/amqp"
)

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
	// MaxConnectionRetry indicates for how long dialing should be retried
	// during connection recovery
	MaxConnectionRetry time.Duration
	// AMQPConfig is optionally used to set up the internal connection to the
	// AMQP broker. If nil, default values (determined by the underlying
	// library) are used.
	AMQPConfig *amqp.Config
}

// DefaultConfig are the default values used when initializing the default
// queue.
var DefaultConfig = &Config{
	ExchangeName:       "hop.exchange",
	Persistent:         false,
	MaxConnectionRetry: 15 * time.Minute,
}
