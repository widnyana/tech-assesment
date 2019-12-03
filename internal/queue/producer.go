package queue

import (
	"github.com/streadway/amqp"
	"kumparan/internal/config"
	"log"
)

const (
	SubjectNews = "news_queue"
	QName       = "newsQueue"
	RoutingKeys = "news_route"
)

var (
	qConn      *amqp.Connection
	svcChannel *amqp.Channel
	svcQueue   amqp.Queue
)

func DeclareExchange(con *amqp.Connection) (*amqp.Channel, error) {
	schan, err := con.Channel()
	if err != nil {
		return nil, err
	}

	if err := schan.ExchangeDeclare(
		SubjectNews, "topic",
		false, false, false, false, nil,
	); err != nil {
		return nil, err
	}

	return schan, nil
}

func DeclareQueue(channel *amqp.Channel) (amqp.Queue, error) {
	return channel.QueueDeclare(QName,
		false, false, false, false, nil)
}

func InitializeConnection(bc config.BrokerConf) error {
	con, err := amqp.Dial(bc.DSN())
	if err != nil {
		return err
	}

	schan, err := DeclareExchange(con)
	if err != nil {
		return err
	}

	qConn = con
	svcChannel = schan

	return nil
}

func GetBrokerCon() *amqp.Connection {
	return qConn
}

func GetBrokerChannel() *amqp.Channel {
	return svcChannel
}

func GetBrokerQueue() amqp.Queue {
	return svcQueue
}

func Publish(in []byte) error {
	msg := amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "utf8",
		Body:            in,
	}

	con := GetBrokerCon()
	chn, err := DeclareExchange(con)
	if err != nil {
		log.Printf("got error: %s", err)
	}

	err = chn.Publish(
		SubjectNews,
		RoutingKeys,
		false, false,
		msg)
	if err != nil {
		return err
	}

	log.Printf("message sent to %s, %s", SubjectNews, RoutingKeys)
	return nil
}
