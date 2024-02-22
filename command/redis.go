package command

import (
	"github.com/gomodule/redigo/redis"
	"time"
)

func NewPool(server string, db int, username, password string) *redis.Pool {
	return &redis.Pool{

		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,

		Dial: func() (redis.Conn, error) {
			var opts []redis.DialOption
			opts = append(opts, redis.DialDatabase(db), redis.DialConnectTimeout(5*time.Minute))
			if password != "" {
				opts = append(opts, redis.DialUsername(username), redis.DialPassword(password))
			}
			c, err := redis.Dial("tcp", server, opts...)
			if err != nil {
				return nil, err
			}
			return c, err
		},

		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
