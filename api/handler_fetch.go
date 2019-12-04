package api

import (
	"context"
	"encoding/json"
	"fmt"
	"kumparan/internal/cache"
	"kumparan/internal/cetak"
	"kumparan/internal/contract"
	"kumparan/internal/db"
	"kumparan/internal/es"
	"net/http"
	"strconv"
	"sync"
)

const (
	CacheNewsResult = "kNewsCache"
)

type (
	newsFetchResponse struct {
		Data []contract.NewsData `json:"data"`
	}

	newsFetchContainer struct {
		Stack []contract.NewsData `json:"Stack"`
		mu    *sync.Mutex
	}
)

func getCacheKey(page int) string {
	return fmt.Sprintf("%s:%d", CacheNewsResult, page)
}

func (n *newsFetchContainer) append(data contract.NewsData, wg *sync.WaitGroup) {
	defer wg.Done()
	cetak.Printf("appending data: %s", data.Author)

	n.mu.Lock()
	local := append(n.Stack, data)
	n.Stack = local
	n.mu.Unlock()

	cetak.Printf("[!] Stack length: %d", len(n.Stack))
}

func paginate(start int, wg *sync.WaitGroup) (newsFetchContainer, int) {
	var stack newsFetchContainer
	stack.mu = new(sync.Mutex)

	ctx := context.Background()
	esc := es.GetElasticClient()

	cetak.Printf("ES From: %d | size: %d", start, cfg.Srv.PaginationLimit)
	res, err := esc.Search().
		Index(cfg.Elastic.IndexName).
		Sort("created", false).
		From(start).
		Size(cfg.Srv.PaginationLimit).
		Pretty(true).
		Do(ctx)
	if err != nil {
		cetak.Printf("error fecthing data from ES: %s", err)
		return stack, http.StatusInternalServerError
	}

	if res.Hits.TotalHits.Value > 0 {
		cetak.Printf("got %d data from es", len(res.Hits.Hits))
		cetak.Printf("got %d total data from es", res.Hits.TotalHits.Value)
		var cFetch = make(chan contract.NewsOnElastic, 1)
		var cResult = make(chan contract.NewsData, 1)
		var tmp contract.NewsOnElastic

		go func() {
			for data := range cFetch {
				fetcher(data.ID, cResult, wg)
			}
		}()

		go func() {
			for result := range cResult {
				stack.append(result, wg)
			}
		}()

		for _, data := range res.Hits.Hits {
			err := json.Unmarshal(data.Source, &tmp)
			if err != nil {
				cetak.Printf("[!] Got Error: %s", err)
				continue
			}
			cFetch <- tmp
			wg.Add(2)
			println("============================")
		}

	}

	wg.Wait()

	httpcode := http.StatusOK
	if len(stack.Stack) < 1 {
		httpcode = http.StatusNoContent
	}

	return stack, httpcode
}

func fetcher(cID int64, result chan<- contract.NewsData, wg *sync.WaitGroup) {
	defer wg.Done()
	var data contract.NewsData
	query := `SELECT id, author, body, created FROM news WHERE id = ? LIMIT 1;`
	dbcon := db.GetDB()
	st, err := dbcon.Prepare(query)
	if err != nil {
		cetak.Printf("[!] error preparing statement")
		return
	}

	defer st.Close()
	r, err := st.Query(cID)
	if err != nil {
		cetak.Printf("error querying data: %s", err)
		return
	}

	defer r.Close()
	for r.Next() {
		err = r.Scan(&data.ID, &data.Author, &data.Body, &data.Created)
		if err != nil {
			cetak.Printf("error scanning data to struct: %s", err)
			return
		}
	}

	result <- data
	cetak.Printf("done fetching %d", cID)
}

func handleNewsGet(w http.ResponseWriter, r *http.Request) {
	var wg = &sync.WaitGroup{}
	var setCache bool
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil {
		page = 1
	} else if page < 1 {
		page = 1
	}
	start := (page - 1) * cfg.Srv.PaginationLimit

	// check for cache
	cacheKey := getCacheKey(page)
	fromCache, err := fetchFromCache(cache.GetRedisConn(), cacheKey)
	if err != nil {
		setCache = true
		cetak.Printf(err.Error())
	}

	// got cache
	if len(fromCache) > 0 {
		cetak.Printf("responding query with data from cache...")
		responseAsJSON(w, newsFetchResponse{
			Data: fromCache,
		}, http.StatusOK)
		return
	} else {
		setCache = true
	}

	// fetch when cache empty
	result, httpCode := paginate(start, wg)
	if len(result.Stack) > 0 {
		// set to cache
		cetak.Printf("-------------------- setCache: %b", setCache)
		if setCache {
			toCache, err := json.Marshal(result.Stack)
			if err != nil {
				cetak.Printf("[!] error marshaling cache data: %s", err)
			} else {
				println(toCache)
				if err = setToCache(cache.GetRedisConn(), cacheKey, toCache, cfg.Srv.CacheTTL); err != nil {
					cetak.Printf("[!] error saving cache data: %s", err)
				}
			}
		}

		responseAsJSON(w, newsFetchResponse{
			Data: result.Stack,
		}, httpCode)
		return
	}

	cetak.Printf("got %d of data", len(result.Stack))
	http.Error(w, http.StatusText(httpCode), httpCode)
	return
}
