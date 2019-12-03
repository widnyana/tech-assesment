package consumer

import (
	"context"
	"fmt"
	"kumparan/internal/contract"
	"kumparan/internal/db"
	"kumparan/internal/es"
	"log"
	"strconv"
)

func saveNews(n *contract.NewsData) error {
	var commited bool
	log.Println("begin saving news...")
	insertQuery := `INSERT INTO news (author, body, created) VALUES (?, ?, ?)`

	con := db.GetDB()
	tx, err := con.Begin()
	if err != nil {
		return fmt.Errorf("could not begin transaction: %s", err)
	}

	defer func() {
		if !commited {
			if err := tx.Rollback(); err != nil {
				log.Printf("error rolling back transaction: %s", err)
			}
		}
	}()

	prep, err := tx.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("error preparing statement: %s", err)
	}
	defer func() {
		if err := prep.Close(); err != nil {
			log.Println("error closing statement")
		}
	}()

	res, err := prep.Exec(n.Author, n.Body, n.Created)
	if err != nil {
		return fmt.Errorf("error preparing statement: %s", err)
	}

	n.ID, err = res.LastInsertId()
	if err != nil {
		return fmt.Errorf("error getting last insert ID: %s", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error commiting last insert: %s", err)
	}
	commited = true

	return nil
}

func saveToES(n *contract.NewsData) error {
	ctx := context.Background()
	ec := es.GetElasticClient()

	idxExist, err := ec.IndexExists(cfg.Elastic.IndexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("error checking for index: %s", err)
	}

	if !idxExist {
		res, err := ec.CreateIndex(cfg.Elastic.IndexName).Do(ctx)
		if err != nil {
			return fmt.Errorf("error creating index: %s", err)
		}

		if !res.Acknowledged {
			return fmt.Errorf("error, index not acknowledged: %s", err)
		}
	}

	res, err := ec.Index().
		Index(cfg.Elastic.IndexName).
		Type(contract.ElasticNewsIndexType).
		Id(strconv.FormatInt(n.ID, 10)).
		BodyJson(n.ToElasticData()).
		Do(ctx)
	if err != nil {
		return fmt.Errorf("error inserting document to index: %s", err)
	}

	ifr, err := ec.Flush(cfg.Elastic.IndexName).Do(ctx)
	if err != nil {
		return fmt.Errorf("error flushing index to disk: %s", err)
	}

	log.Printf("saved to elastic: %s | total shard: %d \n", res.Id, ifr.Shards.Total)
	return nil
}
