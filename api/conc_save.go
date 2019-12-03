package api

import (
	"errors"
	"fmt"
	"io"
	"kumparan/internal/cetak"
	"kumparan/internal/contract"
	"kumparan/internal/queue"
	"time"
)

type (
	createNewsJob struct {
		body io.ReadCloser
	}

	newsWorker struct {
		id         int
		WorkerPool chan chan createNewsJob
		JobChannel chan createNewsJob
		quit       chan bool
	}

	dispatcher struct {
		wpool      chan chan createNewsJob
		maxWorkers int
	}
)

var (
	jobQueue chan createNewsJob
)

func (job createNewsJob) save(workerID int) error {
	var incoming contract.NewsData

	jderr := jsonDecoder(job.body, &incoming)
	if jderr.isError {
		cetak.Printf("error decoding json body: %s\n", jderr.message)
		return errors.New(jderr.message)
	}
	incoming.Created = time.Now()

	msg, err := incoming.Marshal()
	if err != nil {
		cetak.Printf("cannot marshal NewsData to json")
		return errors.New("cannot marshal NewsData to json")
	}

	// send to message queue
	if err := queue.Publish(msg); err != nil {
		return err
	}

	return nil
}

func newWorker(wpool chan chan createNewsJob, workerNum int) newsWorker {
	return newsWorker{
		id:         workerNum,
		WorkerPool: wpool,
		JobChannel: make(chan createNewsJob),
		quit:       make(chan bool),
	}
}

// Start ...
func (w newsWorker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel

			select {
			case job := <-w.JobChannel:
				cetak.Printf("[w%d] Got job!", w.id)
				if err := job.save(w.id); err != nil {
					cetak.Printf("[w%d]  error: %s", w.id, err.Error())
				} else {
					cetak.Printf("[w%d]  Success!", w.id)
				}
				fmt.Println("============================================")
			case <-w.quit:
				cetak.Printf("[w%d] told to quit", w.id)
				return
			}
		}
	}()
}

func (w newsWorker) stop() {
	go func() {
		w.quit <- true
	}()
}

func newDispatcher(maxWorker int) *dispatcher {
	pool := make(chan chan createNewsJob, maxWorker)
	return &dispatcher{wpool: pool, maxWorkers: maxWorker}
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-jobQueue:
			go func(job createNewsJob) {
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
