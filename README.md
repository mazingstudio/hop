# Hop

[![GoDoc](https://godoc.org/github.com/mazingstudio/hop?status.svg)](https://godoc.org/github.com/mazingstudio/hop)
[![Go Report Card](https://goreportcard.com/badge/github.com/mazingstudio/hop)](https://goreportcard.com/report/github.com/mazingstudio/hop)
[![license](https://img.shields.io/github/license/mashape/apistatus.svg)]()

_An AMQP client wrapper that provides easy work queue semantics_

## Introduction

Hop consists of a simple set of abstractions over AMQP concepts that allow you to think in terms of jobs and topics, rather than queues, exchanges and routing.

### Jobs

The most basic abstraction in Hop are Jobs, which are work units that may be marked as done or failed by your worker processes. When a Job is completed succesfully, it gets removed from the queue. If a Job fails, on the other hand, the worker has the option of requeueing it or dropping it altogether. If a worker dies in the middle of processing a Job, the Job is placed back into the queue. All of this maps to AMQP messages and `ack` and `reject` mechanics.

### Topics

Similarly to the beanstalkd concept of tubes, Hop introduces Topics, which are distinct queues into which producers can put work units, and from which workers can pull them for processing. Underneath, a Topic is a mapping to an AMQP TCP channel, and a queue with the same name as the Topic. Note that Topics are _not_ the same as AMQP exchange topics.

### The Work Queue

The third Hop abstraction is the Work Queue, which is nothing more than the grouping of Topics a worker or producer can interact with. The Work Queue abtracts over the TCP connection to the AMQP broker and performs such tasks as dialing, graceful shutdown, and declaration of new channels. Underneath, it maps to an AMQP direct exchange.

## Usage

```go
package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mazingstudio/hop"
)

func main() {
	// Initialize a WorkQueue with the default configuration
	wq, err := hop.DefaultQueue("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("error creating queue: %s", err)
	}
	// Make sure to gracefully shut down
	defer wq.Close()

	// Get the "tasks" Topic handle
	tasks, err := wq.GetTopic("tasks")
	if err != nil {
		log.Fatalf("error getting topic: %s", err)
	}

	// Put a Job into the "tasks" Topic
	err = tasks.Put([]byte("Hello"))
	if err != nil {
		log.Fatalf("error putting: %s", err)
	}

	// Pull the Job from the Topic
	hello, err := tasks.Pull()
	if err != nil {
		log.Fatalf("error pulling: %s", err)
	}
	// This should print "hello"
	log.Infof("job body: %s", hello.Body())

	// Mark the Job as failed and requeue
	hello.Fail(true)

	// Pull the Job again
	hello2, err := tasks.Pull()
	if err != nil {
		log.Fatalf("error pulling: %s", err)
	}
	log.Infof("job body: %s", hello.Body())

	// Mark the Job as done
	hello2.Done()
}
```

## License & Third Party Code

Hop uses [`streadway/amqp`](https://github.com/streadway/amqp) internally.

Hop is distributed under the MIT License. Please refer to `LICENSE` for more details.