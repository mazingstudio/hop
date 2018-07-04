package hop

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Topic represents a named "tube" from which jobs can be exclusively pulled and
// put into. Underneath, each topic corresponds to a queue bound to a direct
// exchange, and each topic instance manages its own channel.
type Topic struct {
	name   string
	config *Config
	ch     *amqp.Channel
}

// GetTopic returns a "tube" handle, from which to pull and put jobs into.
func (q *WorkQueue) GetTopic(name string) (*Topic, error) {
	ch, err := q.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "error starting new channel")
	}
	_, err = ch.QueueDeclare(name, q.config.Persistent,
		false, false, false, nil)
	if err != nil {
		return nil, errors.Wrap(err, "error declaring queue")
	}
	err = ch.QueueBind(name, name, q.config.ExchangeName, true, nil)
	if err != nil {
		return nil, errors.Wrap(err, "unable to bind queue")
	}
	return &Topic{
		name:   name,
		config: q.config,
		ch:     ch,
	}, nil
}

// Name returns the topic's name.
func (t *Topic) Name() string {
	return t.name
}

// Pull blocks until a job is available to consume and returns it.
func (t *Topic) Pull() (*Job, error) {
	var (
		del amqp.Delivery
		ok  bool
		err error
	)
	for !ok {
		del, ok, err = t.ch.Get(t.name, false)
		if err != nil {
			return nil, errors.Wrap(err, "error getting message")
		}
	}
	return &Job{
		topic: t,
		del:   &del,
	}, nil
}

// Put places a job in the queue with the given body. Use PutPublishing if you
// need more control over your messages.
func (t *Topic) Put(body []byte) error {
	deliveryMode := amqp.Transient
	if t.config.Persistent {
		deliveryMode = amqp.Persistent
	}
	p := &amqp.Publishing{
		DeliveryMode: deliveryMode,
		Body:         body,
	}
	return t.PutPublishing(p)
}

// PutPublishing puts an AMQP message into the queue. Use this if you need more
// granularity in your message parameters.
func (t *Topic) PutPublishing(p *amqp.Publishing) error {
	err := t.ch.Publish(t.config.ExchangeName, t.name, false, false, *p)
	if err != nil {
		return errors.Wrap(err, "error publishing message")
	}
	return nil
}

// Close manually closes the topic's underlying channel. Since closing the
// topic's parent queue also closes the topic's channel, that approach is
// generally preferred.
func (t *Topic) Close() error {
	err := t.ch.Close()
	if err != nil {
		return errors.Wrap(err, "error closing channel")
	}
	return nil
}
