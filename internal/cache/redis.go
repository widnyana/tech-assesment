package cache

import (
	"fmt"
	"github.com/go-redis/redis"
	"kumparan/internal/cetak"
	"kumparan/internal/config"
	"time"
)

type RConn struct {
	client *redis.Client
}

var (
	rCon *RConn
)

func NewHandler(client *redis.Client) *RConn {
	return &RConn{
		client: client,
	}
}

func (r RConn) Set(name string, value []byte, ttl int) error {
	status := r.client.Set(name, value, time.Second*time.Duration(ttl))
	return status.Err()
}

func (r RConn) Get(name string) ([]byte, error) {
	status := r.client.Get(name)

	if status.Err() != nil {
		return nil, status.Err()
	}

	return status.Bytes()

}

//
func (r RConn) Del(name string) (bool, error) {
	status := r.client.Del(name)
	if status.Err() != nil {
		return false, status.Err()
	}

	return true, nil
}

func (r RConn) Ping() error {
	status := r.client.Ping()

	if status.Err() != nil {
		cetak.Printf("error pinging redis: %s", status.Err())
		return status.Err()
	}

	cetak.Printf("result of pinging redis: %s", status.Val())
	return nil
}

func (r RConn) Close() error {
	return r.client.Close()
}

func InitializeRedisCache(c config.RedisConf) error {
	if c.Host == "" || c.Port == 0 {
		return fmt.Errorf("recheck redis config pls")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", c.Host, c.Port),
		Password: c.Password,
		DB:       c.DB,
		PoolSize: 10,
	})

	pool := NewHandler(client)
	rCon = pool

	_ = rCon.Ping()
	cetak.Printf("redis pool initialized!")
	return nil
}

func GetRedisConn() *RConn {
	return rCon
}
