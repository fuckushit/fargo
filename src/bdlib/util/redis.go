package util

import (
	"bdlib/config"
	"bdlib/redis"
	"time"
)

// NewRedisManager _
func NewRedisManager(section config.Section) (redisStore *redis.RedisManager, err error) {
	host, err := section.GetValue("host")
	if err != nil {
		return
	}
	auth, err := section.GetValue("auth")
	if err != nil {
		return
	}
	timeoutMilli, err := section.GetIntValue("timeout")
	if err != nil {
		return
	}
	timeout := time.Duration(timeoutMilli) * time.Millisecond
	poolsize, err := section.GetIntValue("poolsize")
	if err != nil {
		return
	}

	if redisStore, err = redis.NewRedisManager(host, auth, poolsize, timeout); err != nil {
		return
	}

	return
}

// NewRedisManagerTimeout _
func NewRedisManagerTimeout(section config.Section) (redisStore *redis.RedisManager, err error) {

	host, err := section.GetValue("host")
	if err != nil {
		return
	}
	auth, err := section.GetValue("auth")
	if err != nil {
		return
	}
	conTimeoutMilli, err := section.GetIntValue("contimeout")
	if err != nil {
		return
	}
	conTimeout := time.Duration(conTimeoutMilli) * time.Millisecond
	rwTimeoutMilli, err := section.GetIntValue("rwtimeout")
	if err != nil {
		return
	}
	rwTimeout := time.Duration(rwTimeoutMilli) * time.Millisecond
	poolsize, err := section.GetIntValue("poolsize")
	if err != nil {
		return
	}

	if redisStore, err = redis.NewRedisManagerTimeout(host, auth, poolsize, conTimeout, rwTimeout); err != nil {
		return
	}

	return
}
