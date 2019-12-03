package consumer

import (
	"fmt"
	"github.com/streadway/amqp"
	"kumparan/internal/cetak"
	"kumparan/internal/config"
	"kumparan/internal/contract"
	"kumparan/internal/db"
	"kumparan/internal/es"
	"kumparan/internal/queue"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	cfg config.Config
)

func init() {
	cfg = config.GetConfig()
	err := queue.InitializeConnection(cfg.Broker)
	if err != nil {
		log.Fatalf("could not connect to broker: %s", err)
	}

	err = db.InitializeDBConnection(cfg.DB)
	if err != nil {
		log.Fatalf("could not connect to database: %s", err)
	}

	_, err = es.InitializeElasticClient(cfg.Elastic)
	if err != nil {
		log.Fatalf("could not connect to elastic: %s", err)
	}

	cetak.Printf("Running with max queue: %d", cfg.Srv.MaxQueue)
	jobQueue = make(chan Job, cfg.Srv.MaxQueue)


	disp := newDispatcher(cfg.Srv.MaxWorker)
	disp.run()
}

func RunConsumer() {
	c, err := consumer()
	if err != nil {
		log.Fatalf(err.Error())
	}

	// listen for os signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
	go func() {
		for range sig {
			log.Println("got termination signal, commencing shutdown")
			if err := c.Shutdown(); err != nil {
				log.Fatalf("error during shutdown: %s", err)
			}

			os.Exit(0)
		}
	}()

	// run forever
	select {}
}

func consumer() (*queue.Consumer, error) {
	var err error
	log.Println("Creating new AMQP Consumer...")
	cons := queue.NewConsumer("consumer01")
	go func() {
		log.Printf("closing: %s\n", <-cons.Conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Println("Declaring exchange...")
	cons.Channel, err = queue.DeclareExchange(cons.Conn)
	if err != nil {
		return nil, fmt.Errorf("error declaring exchange: %s", err)
	}

	log.Printf("Declaring queue: %s", queue.QName)
	q, err := queue.DeclareQueue(cons.Channel)
	if err != nil {
		return nil, fmt.Errorf("error declaring Queue: %s -> %s", queue.QName, err)
	}

	log.Println("Binding queue...")
	if err = cons.Channel.QueueBind(q.Name, queue.RoutingKeys, queue.SubjectNews, false, nil); err != nil {
		return nil, fmt.Errorf("could not bind queue: %s", err)
	}

	log.Println("Consuming message(s)...")
	pipe, err := cons.Channel.Consume(q.Name, cons.Tag,
		false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("could not initialize consume channel: %s", err)
	}

	// this one should be done concurently.
	go consume(pipe, cons.Done)
	return cons, nil
}

func consume(msgs <-chan amqp.Delivery, done chan error) {
	var err error
	for m := range msgs {
		log.Printf(
			"got %dB delivery: [%v] %q",
			len(m.Body),
			m.DeliveryTag,
			m.Body,
		)

		var n = new(contract.NewsData)
		err = n.UnMarshal(m.Body)
		if err != nil {
			log.Printf("Error unmarshaling message: %s", err)
		}

		jobQueue <- Job{
			Payload:  n,
			delivery: m,
		}

		//err = saveNews(n)
		//if err != nil {
		//	log.Printf("Error saving message to db: %s", err)
		//}
		//
		//err = saveToES(n)
		//if err != nil {
		//	log.Printf("Error saving message to elastic: %s", err)
		//}

		//if err == nil {
		//	// ack when everything is fine.
		//	err = m.Ack(false)
		//	if err != nil {
		//		log.Printf("Error ack'ing message: %s", err)
		//	}
		//}
	}
	log.Printf("handle: deliveries channel closed")
	done <- nil
}
