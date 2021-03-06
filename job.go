package hop

import (
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
)

// Job represents a work unit taken from the work queue.
type Job struct {
	topic *Topic
	del   *amqp.Delivery
	ch    *amqp.Channel
}

// Done marks the Job for queue removal.
func (j *Job) Done() error {
	defer j.topic.queue.putChannel(j.ch)
	err := j.ch.Ack(j.del.DeliveryTag, false)
	if err != nil {
		return errors.Wrap(err, "error acknowledging delivery")
	}
	return nil
}

// Fail marks the Job as failed. If requeue is true, the Job will be added
// back to the queue; otherwise, it will be dropped.
func (j *Job) Fail(requeue bool) error {
	defer j.topic.queue.putChannel(j.ch)
	err := j.ch.Nack(j.del.DeliveryTag, false, requeue)
	if err != nil {
		return errors.Wrap(err, "error rejecting delivery")
	}
	return nil
}

// Body returns the Job's body
func (j *Job) Body() []byte {
	return j.del.Body
}

// Delivery returns the Job's underlying delivery. Use this if you need more
// control over the AMQP message.
func (j *Job) Delivery() *amqp.Delivery {
	return j.del
}
