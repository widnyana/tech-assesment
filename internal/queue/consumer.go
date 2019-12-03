package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

type (
	Consumer struct {
		Conn    *amqp.Connection
		Channel *amqp.Channel
		Tag     string
		Done    chan error
	}
)

func NewConsumer(ctag string) *Consumer {
	c := &Consumer{
		Conn:    GetBrokerCon(),
		Channel: nil,
		Tag:     ctag,
		Done:    make(chan error),
	}

	return c
}

func (c *Consumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.Channel.Cancel(c.Tag, true); err != nil {
		return fmt.Errorf("Consumer cancel failed: %s", err)
	}

	if err := c.Conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %s", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.Done

}
