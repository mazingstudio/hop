package hop

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// WorkQueue is an abstraction over an AMQP connection that automatically
// performs management and configuration to provide easy work queue semantics.
type WorkQueue struct {
	conn        *connection
	config      *Config
	channelPool sync.Pool
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
	conn := newConnection(addr, config)
	err := conn.connect()
	if err != nil {
		return nil, errors.Wrap(err, "error connecting")
	}
	return newQueue(conn, config)
}

func newQueue(conn *connection, config *Config) (*WorkQueue, error) {
	q := &WorkQueue{
		conn:   conn,
		config: config,
		channelPool: sync.Pool{
			New: conn.newChannel,
		},
	}
	if err := q.init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize queue")
	}
	return q, nil
}

func (q *WorkQueue) init() error {
	ch, err := q.getChannel()
	if err != nil {
		return errors.Wrap(err, "error getting management channel")
	}
	defer q.putChannel(ch)

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

// --- Channels ---

func (q *WorkQueue) getChannel() (*amqp.Channel, error) {
	res := q.channelPool.Get()
	switch v := res.(type) {
	case error:
		return nil, errors.Wrap(v, "channel retrieval failed permanently")
	case *amqp.Channel:
		return v, nil
	}
	return nil, nil
}

func (q *WorkQueue) putChannel(ch *amqp.Channel) {
	q.channelPool.Put(ch)
}
