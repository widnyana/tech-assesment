package api

import (
	"kumparan/internal/config"
	"kumparan/internal/queue"
	"log"
	"net/http"
)

var cfg config.Config

func init() {
	cfg = config.GetConfig()
	err := queue.InitializeConnection(cfg.Broker)
	if err != nil {
		log.Fatalf("error opening connection to broker: %s", err)
	}

	jobQueue = make(chan createNewsJob, cfg.Srv.MaxQueue)
	disp := newDispatcher(cfg.Srv.MaxWorker)
	disp.run()

}

func RunAPI() {
	http.Handle("/news", http.HandlerFunc(handleCreation))
	log.Printf("Running news service on %s\n", cfg.Srv.Bind)
	if err := http.ListenAndServe(cfg.Srv.Bind, nil); err != nil {
		log.Printf("Error bang :( %s", err.Error())
	}
}
