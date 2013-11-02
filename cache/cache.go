package cache

import (
	"net/url"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/simonz05/util/log"
)

type Cache struct {
	cfg  *config
	Pool *redis.Pool
}

type config struct {
	password string
	db       uint8
	addr     string
}

func parseDSN(dsn string) (*config, error) {
	cfg := new(config)

	if dsn == "" {
		dsn = "redis://:@localhost:6379/0"
	}

	u, err := url.Parse(dsn)

	if err != nil {
		return nil, err
	}

	if pass, ok := u.User.Password(); ok {
		cfg.password = pass
	}

	db := u.Path

	if len(db) > 1 && db[0] == '/' {
		db = db[1:len(db)]
	}

	idb, err := strconv.ParseUint(db, 10, 8)

	if err != nil {
		idb = 0
	}

	cfg.db = uint8(idb)
	cfg.addr = u.Host
	return cfg, nil
}

func Open(dataSourceName string) (*Cache, error) {
	var err error

	cache := new(Cache)
	cache.cfg, err = parseDSN(dataSourceName)

	if err != nil {
		return nil, err
	}

	cache.Pool = &redis.Pool{
		MaxIdle:     128,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return cache.dial()
		},
		TestOnBorrow: nil,
	}
	return cache, nil
}

func (cache *Cache) Get() redis.Conn {
	return cache.Pool.Get()
}

func (cache *Cache) dial() (redis.Conn, error) {
	conn, err := redis.Dial("tcp", cache.cfg.addr)

	if err != nil {
		return nil, err
	}

	if cache.cfg.password != "" {
		if _, err := conn.Do("AUTH", cache.cfg.password); err != nil {
			log.Errorf("Redis AUTH err: %v", err)
			conn.Close()
			return nil, err
		}
	}

	if cache.cfg.db != 0 {
		if _, err := conn.Do("SELECT", cache.cfg.db); err != nil {
			log.Errorf("Redis SELECT err: %v", err)
			conn.Close()
			return nil, err
		}
	}

	return conn, nil
}
