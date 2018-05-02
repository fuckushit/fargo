package session

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// hmac key for encrypt and decrypt.
var (
	HMACKEY = "&^34x$Mf"
)

// SessionStore session 存储接口.
type SessionStore interface {
	Set(key, value interface{}) (err error)
	Get(key interface{}) (value interface{})
	Delete(key interface{}) (err error)
	SessionID() (sid string)
	Flush() (err error)
}

// Provider session 句柄接口.
type Provider interface {
	SessionInit(maxlifetime int64, options map[interface{}]interface{}) (err error)
	SessionRead(sid string) (store SessionStore, err error)
	SessionExists(sid string) (exist bool)
	SessionRegenerate(oldsid, sid string) (store SessionStore, err error)
	SessionDestroy(sid string) (err error)
	// SessionAll() (all int)
	// SessionGC()
}

// providers session 句柄集合.
var providers = make(map[string]Provider)

// Register 注册一个 session provider.
// Parameters:
// - name:     注册的 provider 名称.
// - provider: 注册的 provider 接口.
func Register(name string, provider Provider) {
	if provider == nil {
		panic("session: Register provider is nil")
	}
	if _, ok := providers[name]; ok {
		panic(fmt.Sprintf("session: Provider %s is registered", name))
	}
	providers[name] = provider
}

// Manager session 操作对象.
type Manager struct {
	cookieName  string
	provider    Provider
	maxlifeTime int64
}

// 初始化时将 redis 存储对象注册到集合中.
func init() {

	Register("redis", redisProvider)

	// TODO mysql ...
}

// NewManager 新建 session manager.
// Parameters:
// - providerName: 要使用的 session 存储类型, 如 redis, memcache, mysql 等.
// - cookieName:   写入 cookie 的名字.
// - cookieName:   session 失效时间.
// - cookieName:   session 存储参数.
// Return:
//  - manager:     session 操作对象.
//  - err
func NewManager(providerName, cookieName string, maxlifeTime int64, options map[interface{}]interface{}) (manager *Manager, err error) {

	provider, ok := providers[providerName]
	if !ok {
		err = fmt.Errorf("session: provider %s not found", providerName)
		return
	}
	if err = provider.SessionInit(maxlifeTime, options); err != nil {
		return
	}
	manager = &Manager{
		cookieName:  cookieName,
		provider:    provider,
		maxlifeTime: maxlifeTime,
	}

	return
}

// SessionStart 当 cookie 存在时读取出 session 对象, 不存在时新建一个新的 session 对象, 并设置 session 失效时间.
// Parameters:
// - w:        http response.
// - r:        http request.
// Return:
//  - session: session 操作对象.
func (m *Manager) SessionStart(w http.ResponseWriter, r *http.Request) (session SessionStore) {
	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		sid := m.sessionID(r)
		session, _ = m.provider.SessionRead(sid)
		cookie = &http.Cookie{
			Name:     m.cookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)
	} else {
		sid, _ := url.QueryUnescape(cookie.Value)
		if m.provider.SessionExists(sid) {
			session, _ = m.provider.SessionRead(sid)
		} else {
			sid = m.sessionID(r)
			session, _ = m.provider.SessionRead(sid)
			cookie = &http.Cookie{
				Name:     m.cookieName,
				Value:    url.QueryEscape(sid),
				Path:     "/",
				HttpOnly: true,
			}
			http.SetCookie(w, cookie)
			r.AddCookie(cookie)
		}
	}

	return
}

// SessionDestroy 销毁 session.
// Parameters:
// - w:        http response.
// - r:        http request.
func (m *Manager) SessionDestroy(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie(m.cookieName)
	if err != nil || cookie.Value == "" {
		return
	}

	m.provider.SessionDestroy(cookie.Value)
	expiration := time.Now()
	cookie = &http.Cookie{
		Name:     m.cookieName,
		Path:     "/",
		HttpOnly: true,
		Expires:  expiration,
		MaxAge:   -1,
	}
	http.SetCookie(w, cookie)

	return
}

// SessionRegenerate 重新生成 session store.
// Parameters:
// - w:        http response.
// - r:        http request.
// Return:
//  - session: session 操作对象.
func (m *Manager) SessionRegenerate(w http.ResponseWriter, r *http.Request) (session SessionStore) {
	sid := m.sessionID(r)
	cookie, err := r.Cookie(m.cookieName)
	if err != nil && cookie.Value == "" {
		session, _ = m.provider.SessionRead(sid)
		cookie = &http.Cookie{
			Name:     m.cookieName,
			Value:    url.QueryEscape(sid),
			Path:     "/",
			HttpOnly: true,
		}
	} else {
		oldsid, _ := url.QueryUnescape(cookie.Value)
		session, _ = m.provider.SessionRegenerate(oldsid, sid)
		cookie.Value = url.QueryEscape(sid)
		cookie.HttpOnly = true
		cookie.Path = "/"
	}
	http.SetCookie(w, cookie)
	r.AddCookie(cookie)

	return
}

// sessionID 通过随机 string, unix nano time, remote addr 和 hash 函数生成 session id.
// Parameters:
// - r:    http request.
// Return:
//  - sid: 生成的 sid 字符串.
func (m *Manager) sessionID(r *http.Request) (sid string) {
	bs := make([]byte, 24)
	if _, err := io.ReadFull(rand.Reader, bs); err != nil {
		return
	}
	sig := fmt.Sprintf("%s%d%s", r.RemoteAddr, time.Now().UnixNano(), bs)
	h := hmac.New(sha1.New, []byte(HMACKEY))
	fmt.Fprintf(h, "%s", sig)
	sid = hex.EncodeToString(h.Sum(nil))

	return
}
