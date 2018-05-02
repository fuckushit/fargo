package session

import (
	"bdlib/crypt"
	"bdlib/redis"
	"bdlib/util"
	"sync"
	"time"
)

// 默认 redis 连接参数
var (
	gDefaultRedisHost            = "127.0.0.1:6379"
	gDefaultPoolSize       int64 = 20
	gDefaultDB             int64 = 0
	gDefaultExpireChanSize       = 1024
	gDefaultTimeout              = 100
)

// RedisStore redis 存储对象, 实现了接口 SessionStore.
type RedisStore struct {
	mgr         *redis.RedisManager
	sid         string
	lock        *sync.RWMutex
	values      map[interface{}]interface{}
	maxlifeTime int64
}

// RedisProvider redis 句柄对象, 实现了接口 Provider.
type RedisProvider struct {
	mgr         *redis.RedisManager
	host        string
	db          int64
	poolsize    int64
	auth        string
	expireChan  chan string
	maxlifeTime int64
}

// redisProvider 默认 redis 句柄对象, session_init 时会重新进行设置.
var redisProvider = &RedisProvider{
	mgr:         new(redis.RedisManager),
	maxlifeTime: 0,
	host:        gDefaultRedisHost,
	db:          gDefaultDB,
	poolsize:    gDefaultPoolSize,
	auth:        "",
	expireChan:  make(chan string, gDefaultExpireChanSize),
}

// Set 设置 session key - value,
// Parameters:
//  - key:   session key
//  - value: session value
// Return
//  - err
func (r *RedisStore) Set(key, value interface{}) (err error) {
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	r.values[key] = value
	val, err := encodeGob(r.values)
	if err != nil {
		return
	}

	_, err = r.mgr.Set(r.sid, val)

	return
}

// Get 获取 session value,
// Parameters:
// - key:   session key
// Return:
// - value: session value
func (r *RedisStore) Get(key interface{}) (value interface{}) {
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	v, err := r.mgr.Get(r.sid)
	if err != nil {
		return
	}
	if v == nil {
		return
	}
	val, err := decodeGob(v.([]byte))
	if err != nil {
		return
	}
	value = val[key]

	return
}

// Delete 删除 session,
// Parameters:
// - key:   session key
// Return:
//  - err
func (r *RedisStore) Delete(key interface{}) (err error) {
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	delete(r.values, key)
	val, err := encodeGob(r.values)
	if err != nil {
		return
	}

	_, err = r.mgr.Set(r.sid, val)

	return
}

// Flush 清除所有 session.
// Return:
//  - err
func (r *RedisStore) Flush() (err error) {
	r.lock.Lock()
	defer func() {
		r.lock.Unlock()
	}()

	r.values = make(map[interface{}]interface{})
	val, err := encodeGob(r.values)
	if err != nil {
		return
	}

	_, err = r.mgr.Set(r.sid, val)

	return
}

// SessionID 获取 session id.
// Return:
//  - sid: session id.
func (r *RedisStore) SessionID() (sid string) {
	return r.sid
}

// SessionInit 初始化 session 句柄对象.
// Parameters:
// - maxlifetime: session 失效时间.
// - options:     session 存储选项.
// Return:
//  - err
func (r *RedisProvider) SessionInit(maxlifetime int64, options map[interface{}]interface{}) (err error) {
	r.maxlifeTime = maxlifetime
	// host
	if host, ok := options["host"]; ok {
		r.host = host.(string)
	} else {
		r.host = gDefaultRedisHost
	}
	// db
	if db, ok := options["db"]; ok {
		r.db = util.Int64(db.(string))
	} else {
		r.db = gDefaultDB
	}
	// poolsize
	if poolsize, ok := options["poolsize"]; ok {
		r.poolsize = util.Int64(poolsize.(string))
	} else {
		r.poolsize = gDefaultPoolSize
	}
	if r.poolsize == 0 {
		r.poolsize = gDefaultPoolSize
	}
	// auth
	if auth, ok := options["auth"]; ok && len(auth.(string)) != 0 {
		if options["auth_key"] == nil {
			return
		}
		var authStr string
		authStr, err = crypt.Decrypt(auth.(string), options["auth_key"].(string))
		if err != nil {
			return
		}
		r.auth = authStr
	}
	timeoutMilli := options["timeout"]
	var timeoutMilliInt int
	if timeoutMilli == nil {
		timeoutMilliInt = gDefaultTimeout
	} else {
		timeoutMilliInt = timeoutMilli.(int)
	}
	timeout := time.Duration(timeoutMilliInt) * time.Millisecond

	r.mgr, err = redis.NewRedisManager(r.host, r.auth, int(r.poolsize), timeout)
	if err != nil {
		return
	}

	// session 失效时间
	go r.setExpire()

	return
}

// SessionRead 通过 sid 获取 session 操作句柄对象.
// Parameters:
// - sid: session 的 sid.
// Return:
//  - ss: session 句柄对象.
//  - err:
func (r *RedisProvider) SessionRead(sid string) (ss SessionStore, err error) {
	kvs, err := r.mgr.Get(sid)
	if err != nil {
		return
	}
	var kv map[interface{}]interface{}
	if kvs == nil {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = decodeGob(kvs.([]byte))
		if err != nil {
			return
		}
	}
	lock := new(sync.RWMutex)

	ss = &RedisStore{mgr: r.mgr, sid: sid, lock: lock, values: kv, maxlifeTime: r.maxlifeTime}
	r.expireChan <- sid

	return
}

// SessionExists 判断 sid 对应的 session 是否存在.
// Parameters:
// - sid:    session 的 sid.
// Return:
//  - exist: session 是否存在, 返回 true or false.
func (r *RedisProvider) SessionExists(sid string) (exist bool) {
	var err error
	if r == nil || r.mgr == nil {
		return
	}
	if exist, err = r.mgr.Exists(sid); err != nil {
		return
	}

	return true
}

// SessionRegenerate 将旧的 sid 对应的信息赋值给新的 sid, session 的内容不变.
// Parameters:
// - oldsid: 旧的 session 的 sid.
// - sid:    新的 session 的 sid.
// Return:
//  - ss:    session 句柄对象.
//  - err:
func (r *RedisProvider) SessionRegenerate(oldsid, sid string) (ss SessionStore, err error) {
	var existed bool
	if existed, err = r.mgr.Exists(oldsid); err != nil || !existed {
		r.mgr.Set(sid, "")
	} else {
		r.mgr.Rename(oldsid, sid)
	}

	kvs, err := r.mgr.Get(sid)
	if err != nil {
		return
	}
	var kv map[interface{}]interface{}
	if kvs == nil {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = decodeGob(kvs.([]byte))
		if err != nil {
			return
		}
	}
	lock := new(sync.RWMutex)

	ss = &RedisStore{mgr: r.mgr, sid: sid, lock: lock, values: kv, maxlifeTime: r.maxlifeTime}
	r.expireChan <- sid

	return
}

// SessionDestroy 删除 sid 对应的 session.
// - sid:  session 的 sid.
// Return:
//  - err:
func (r *RedisProvider) SessionDestroy(sid string) (err error) {
	_, err = r.mgr.Del(sid)

	return
}

// setExpire 设置 session 失效时间.
func (r *RedisProvider) setExpire() {
	for sid := range r.expireChan {
		r.mgr.Expire(sid, r.maxlifeTime)
	}

	return
}
