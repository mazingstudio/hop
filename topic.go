package hop

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Topic represents a named "tube" from which jobs can be exclusively pulled and
// put into. Underneath, each topic corresponds to a queue bound to a direct
// exchange, and each topic instance manages its own channel.
type Topic struct {
	queue  *WorkQueue
	name   string
	config *Config
}

// GetTopic returns a "tube" handle, from which to pull and put jobs into.
// Thanks to AMQP protocol rules, Topic declarations are idempotent, meaning
// that Topics only get created if they don't already exist.
func (q *WorkQueue) GetTopic(name string) (*Topic, error) {
	ch, err := q.getChannel()
	if err != nil {
		return nil, errors.Wrap(err, "error getting channel")
	}
	defer q.putChannel(ch)
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
		queue:  q,
		name:   name,
		config: q.config,
	}, nil
}

// Name returns the Topic's name.
func (t *Topic) Name() string {
	return t.name
}

// Pull blocks until a Job is available to consume and returns it.
func (t *Topic) Pull() (*Job, error) {
	var (
		del amqp.Delivery
		ok  bool
		err error
	)
	ch, err := t.queue.getChannel()
	if err != nil {
		return nil, errors.Wrap(err, "error getting channel")
	}
	defer t.queue.putChannel(ch)
	for !ok {
		del, ok, err = ch.Get(t.name, false)
		if err != nil {
			return nil, errors.Wrap(err, "error getting message")
		}
	}
	return &Job{
		topic: t,
		del:   &del,
	}, nil
}

// Put places a Job in the queue with the given body. Use PutPublishing if you
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
	ch, err := t.queue.getChannel()
	if err != nil {
		return errors.Wrap(err, "error getting channel")
	}
	defer t.queue.putChannel(ch)
	err = ch.Publish(t.config.ExchangeName, t.name, false, false, *p)
	if err != nil {
		return errors.Wrap(err, "error publishing message")
	}
	return nil
}
