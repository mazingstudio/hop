package hop

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// WorkQueue is an abstraction over an AMQP connection that automatically
// performs management and configuration to provide easy work queue semantics.
type WorkQueue struct {
	conn   *amqp.Connection
	config *Config
}

const exchangeKind = "direct"

// DefaultQueue creates a queue with the default configuration and connection
// parameters.
func DefaultQueue(addr string) (*WorkQueue, error) {
	return ConnectQueue(addr, DefaultConfig)
}

// ConnectQueue dials using the provided address string and creates a queue
// with the passed configuration. If you need to manage your own connection,
// use NewQueue.
func ConnectQueue(addr string, config *Config) (*WorkQueue, error) {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}
	return NewQueue(conn, config)
}

// NewQueue allows you to manually construct a WorkQueue with the provided
// parameters. Due to the slight verbosity of this process, DefaultQueue and
// ConnectQueue are generally recommended instead.
func NewQueue(conn *amqp.Connection, config *Config) (*WorkQueue, error) {
	q := &WorkQueue{
		conn:   conn,
		config: config,
	}
	if err := q.init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize queue")
	}
	return q, nil
}

func (q *WorkQueue) init() error {
	ch, err := q.conn.Channel()
	if err != nil {
		return errors.Wrap(err, "error getting management channel")
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(q.config.ExchangeName, exchangeKind,
		false, false, false, false, nil)
	if err != nil {
		return errors.Wrap(err, "error declaring exchange")
	}
	return nil
}

// Close gracefully closes the connection to the message broker. DO NOT use this
// if you're managing your AMQP connection manually.
func (q *WorkQueue) Close() error {
	err := q.conn.Close()
	if err != nil {
		return errors.Wrap(err, "error closing connection")
	}
	return nil
}
