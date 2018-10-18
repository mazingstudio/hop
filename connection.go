package hop

import (
	"sync"

	"github.com/cenkalti/backoff"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

type connection struct {
	addr   string
	config *Config
	conn   *amqp.Connection
	mu     sync.Mutex
}

func newConnection(addr string, config *Config) *connection {
	return &connection{
		addr:   addr,
		config: config,
	}
}

func (c *connection) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	conn, err := c.dialWithBackOff()
	if err != nil {
		return errors.Wrap(err, "queue connection failed permanently")
	}
	c.conn = conn
	return nil
}

func (c *connection) dialWithBackOff() (*amqp.Connection, error) {
	var conn *amqp.Connection
	connect := func() error {
		var err error
		if c.config.AMQPConfig != nil {
			conn, err = amqp.DialConfig(c.addr, *c.config.AMQPConfig)
			return err
		}
		conn, err = amqp.Dial(c.addr)
		return err
	}
	b := backoff.NewExponentialBackOff()
	b.MaxElapsedTime = c.config.MaxConnectionRetry
	err := backoff.Retry(connect, b)
	if err != nil {
		return nil, errors.Wrap(err, "error dialing")
	}
	return conn, nil
}

func (c *connection) newChannel() interface{} {
	var (
		ch  *amqp.Channel
		err error
	)
	for err == nil {
		ch, err = c.conn.Channel()
		if err != nil {
			err = c.connect()
			continue
		}
		return ch
	}
	return errors.Wrap(err, "failed to initialize channel")
}

func (c *connection) Close() error {
	return c.conn.Close()
}
