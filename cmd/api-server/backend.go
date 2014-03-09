package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/simonz05/util/kvstore"
	"github.com/simonz05/util/session"
)

type RedisBackend struct {
	region        string
	apiIndex      string
	sessionPrefix string
	db            *kvstore.KVStore
}

func NewRedisBackend(db *kvstore.KVStore, region string) *RedisBackend {
	region = strings.ToLower(region)
	return &RedisBackend{
		region:        region,
		apiIndex:      fmt.Sprintf("%s:api-token", region),
		sessionPrefix: fmt.Sprintf("%s:session", region),
		db:            db,
	}
}

func (w *RedisBackend) Count() (int, error) {
	conn := w.db.Get()
	defer conn.Close()
	return redis.Int(conn.Do("ZCARD", w.apiIndex))
}

func (w *RedisBackend) Get() ([]string, error) {
	conn := w.db.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("ZRANGE", w.apiIndex, 0, -1))
}

func (w *RedisBackend) sessionKey(token string) string {
	return fmt.Sprintf("%s:%s", w.sessionPrefix, token)
}

func (w *RedisBackend) Set(token string, ses *session.Session) error {
	data, err := json.Marshal(ses)

	if err != nil {
		return err
	}

	conn := w.db.Get()
	defer conn.Close()
	conn.Send("MULTI")
	conn.Do("ZADD", w.apiIndex, time.Now().UTC().Unix(), token)
	conn.Do("SET", w.sessionKey(token), data)
	_, err = conn.Do("EXEC")
	return err
}

func (w *RedisBackend) Delete(token string) error {
	conn := w.db.Get()
	defer conn.Close()
	conn.Send("MULTI")
	conn.Send("ZREM", w.apiIndex, token)
	conn.Send("DEL", w.sessionKey(token))
	_, err := conn.Do("EXEC")
	return err
}
