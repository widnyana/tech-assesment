package consumer

import (
	"fmt"
	"github.com/streadway/amqp"
	"kumparan/internal/cetak"
	"kumparan/internal/contract"
)

type (
	Job struct {
		Payload  *contract.NewsData
		delivery amqp.Delivery
	}

	Worker struct {
		id          int
		WorkerPool  chan chan Job
		JobPipeline chan Job
		quit        chan bool
	}

	dispatcher struct {
		wpool      chan chan Job
		maxWorkers int
	}
)

var (
	jobQueue chan Job
)

func (j Job) save(workerID int) error {
	err := saveNews(j.Payload)
	if err != nil {
		cetak.Printf("w:%d | error saving %s to DB\n", workerID, j.Payload.Author)
		return err
	} else {
		cetak.Printf("w:%d | success saving %s to DB: %d\n", workerID, j.Payload.Author, j.Payload.ID)
		err = saveToES(j.Payload)
		if err != nil {
			cetak.Printf("w:%d | Error saving message to elastic: %s", workerID, err)
			return err
		}
	}

	err = j.delivery.Ack(false)
	if err != nil {
		cetak.Printf("Error ack'ing message: %s", err)
		return err
	}

	return nil
}

func newWorker(wpool chan chan Job, workerNum int) Worker {
	return Worker{
		id:          workerNum,
		WorkerPool:  wpool,
		JobPipeline: make(chan Job),
		quit:        make(chan bool),
	}
}

// Start ...
func (w Worker) Start() {
	cetak.Printf("worker id: %d started", w.id)
	go func() {
		for {
			w.WorkerPool <- w.JobPipeline

			select {
			case job := <-w.JobPipeline:
				cetak.Printf("(%s) [w%d] Got job!", job.Payload.ID, w.id)
				if err := job.save(w.id); err != nil {
					cetak.Printf("(%s) [w%d]  error: %s", job.Payload.ID, w.id, err.Error())
				} else {
					cetak.Printf("(%s) [w%d]  Success!", job.Payload.ID, w.id)
				}
				fmt.Println("============================================")
			case <-w.quit:
				cetak.Printf("[w%d] told to quit", w.id)
				return
			}
		}
	}()
}

func (w Worker) stop() {
	go func() {
		w.quit <- true
	}()
}

func newDispatcher(maxWorker int) *dispatcher {
	pool := make(chan chan Job, maxWorker)
	return &dispatcher{wpool: pool, maxWorkers: maxWorker}
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-jobQueue:
			cetak.Printf("wololo!")
			go func(job Job) {
				jobChannel := <-d.wpool
				jobChannel <- job
			}(job)
		}
	}
}

func (d *dispatcher) run() {
	cetak.Printf("Running dispatcher for %d worker", d.maxWorkers)
	for i := 0; i < d.maxWorkers; i++ {
		worker := newWorker(d.wpool, i)
		cetak.Printf("running worker #%d", i)
		worker.Start()
	}

	go d.dispatch()
}



