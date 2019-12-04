package api

import (
	"kumparan/internal/cache"
	"kumparan/internal/config"
	"kumparan/internal/db"
	"kumparan/internal/es"
	"kumparan/internal/queue"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var cfg config.Config

func init() {
	cfg = config.GetConfig()

	err := queue.InitializeConnection(cfg.Broker)
	if err != nil {
		log.Fatalf("error opening connection to broker: %s", err)
	}

	_, err = es.InitializeElasticClient(cfg.Elastic)
	if err != nil {
		log.Fatalf("error opening connection to elasticsearch: %s", err)
	}

	err = db.InitializeDBConnection(cfg.DB)
	if err != nil {
		log.Fatalf("error opening connection to database: %s", err)
	}

	err = cache.InitializeRedisCache(cfg.Redis)
	if err != nil {
		log.Fatalf("error opening connection to cache: %s", err)
	}

	jobQueue = make(chan createNewsJob, cfg.Srv.MaxQueue)
	disp := newDispatcher(cfg.Srv.MaxWorker)
	disp.run()

}

func RunAPI() {
	// listen for os signal
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGTERM)
	go func() {
		for range sig {
			log.Println("got termination signal, commencing shutdown")

			_ = queue.GetBrokerCon().Close()
			_ = db.GetDB().Close()
			_ = es.GetElasticClient().CloseIndex(cfg.Elastic.IndexName)
			_ = cache.GetRedisPool().Close()
			os.Exit(0)
		}
	}()

	http.Handle("/news", http.HandlerFunc(handleNewsEndpoint))
	log.Printf("Running news service on %s\n", cfg.Srv.Bind)
	if err := http.ListenAndServe(cfg.Srv.Bind, nil); err != nil {
		log.Printf("Error bang :( %s", err.Error())
	}

}
