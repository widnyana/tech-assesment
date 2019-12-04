package cache

import (
	"fmt"
	"github.com/gomodule/redigo/redis"
	"kumparan/internal/cetak"
	"kumparan/internal/config"
	"time"
)

type RConn struct {
	con redis.Conn
}

var (
	rCon *redis.Pool
)

func (r RConn) Ping() {
	s, err := redis.String(r.Do("PING"))
	if err != nil {
		cetak.Printf("error pinging redis: %s", err)
		return
	}

	cetak.Printf("result of pinging redis: %s", s)
}

func (r RConn) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return r.con.Do(commandName, args)
}

func (r RConn) Set(name string, value []byte, ttl int) (string, error) {
	return redis.String(r.con.Do("SET", name, value, "EX", ttl))
}

func (r RConn) Get(name string) ([]byte, error) {
	return redis.Bytes(r.con.Do("GET", name))
}

func (r RConn) Del(name string) (bool, error) {
	return redis.Bool(r.con.Do("DEL", name))
}

func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		// Dial or DialContext must be set. When both are set, DialContext takes precedence over Dial.
		Dial: func() (redis.Conn, error) { return redis.DialURL(addr) },
	}
}

func InitializeRedisCache(c config.RedisConf) error {
	if c.Host == "" || c.Port == 0 {
		return fmt.Errorf("recheck redis config pls")
	}

	pool := newPool(c.DSN())
	rCon = pool

	cetak.Printf("redis pool initialized!: %s", c.DSN())
	return nil
}

func GetRedisConn() RConn {

	c := rCon.Get()
	return RConn{c}
}


func GetRedisPool() *redis.Pool {
	return rCon
}
