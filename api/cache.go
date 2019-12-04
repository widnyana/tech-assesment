package api

import (
	"encoding/json"
	"fmt"
	"kumparan/internal/cache"
	"kumparan/internal/cetak"
	"kumparan/internal/contract"
)

func fetchFromCache(con cache.RConn, key string) ([]contract.NewsData, error) {
	result, err := con.Get(key)
	if err != nil {
		err = fmt.Errorf("could not fetch result from cache %s : %s", key, err)
		return nil, err
	}

	var sets []contract.NewsData
	err = json.Unmarshal(result, &sets)
	if err != nil {
		err = fmt.Errorf("could not unmarshal result from cache: %s", err)
		return nil, err
	}

	return sets, nil

}

func setToCache(con cache.RConn, key string, data []byte, ttl int) error {
	con.Ping()
	status, err := con.Set(key, data, ttl)
	if err != nil {
		return fmt.Errorf("error storing data to cache: %s", err)
	}

	cetak.Printf("storing data to cache. key: %s | status: %s", key, status)
	return nil
}
