package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

var _ = fmt.Println

const (
	// MISHOPCLIENTID 商城client的id180100031022
	MISHOPCLIENTID = "180100031052"
	// MISHOPAUTHFIXKEY 商城auth的fixkey
	MISHOPAUTHFIXKEY = "7a675ae5a2e256e87340dd3dac4f5b36"
)

// HTTPClient ...
type HTTPClient struct {
	client       *http.Client
	serviceToken string
	cookie       *http.Cookie
	//loginStatu   bool
}

// NewHTTPClient 新建一个client, timeout单位ms
func NewHTTPClient(proxyAddr string, timeout int) *HTTPClient {
	jar, _ := cookiejar.New(nil)
	lazy := time.Duration(timeout) * time.Millisecond
	cli := &HTTPClient{
		client: &http.Client{
			Jar:     jar,
			Timeout: lazy,
		},
	}
	if len(proxyAddr) != 0 {
		proxy, err := url.Parse(proxyAddr)
		if err != nil {
			fmt.Println(err)
			return cli
		}
		cli.client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
	}
	return cli
}

// SetToken 设置serviceToken
func (c *HTTPClient) SetToken(token string) {
	c.cookie = &http.Cookie{
		Name:  "serviceToken",
		Value: token,
	}
	//c.loginStatu = true
}

// SetDappToken 设置serviceToken
func (c *HTTPClient) SetDappToken() {
	c.cookie = &http.Cookie{
		Name:  "PHPSESSID",
		Value: "quh90et6mvgj9bc2bg3dm3e5v3",
		// Value: "60u2un7dogo0s6k9sjugtufan0",
	}
}

// GetCookie ...
func (c *HTTPClient) GetCookie() *http.Cookie {
	return c.cookie
}

// PostForm POST格式化数据
func (c *HTTPClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	return c.Post(url, strings.NewReader(data.Encode()))
}

// Post post数据包到服务端
func (c *HTTPClient) Post(url string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Mishop-Client-Id", MISHOPCLIENTID)

	return c.doRequest(req)
}

// PostWithAuth post数据包到服务端
func (c *HTTPClient) PostWithAuth(url string, postData url.Values) (resp *http.Response, err error) {
	body := strings.NewReader(postData.Encode())
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Screen-Width-Px", "1080")
	req.Header.Set("Screen-DensityDpi", "480")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Mishop-Client-Id", MISHOPCLIENTID)
	req.Header.Set("Mishop-Auth", GenHmac(postData, MISHOPAUTHFIXKEY))
	return c.doRequest(req)
}

// PostDapp post数据包到dapp服务端
func (c *HTTPClient) PostDapp(url string, data map[string]interface{}) (resp *http.Response, err error) {
	if data == nil {
		data = make(map[string]interface{})
	}
	data["imei"] = "C0FDF778CF09265D11ABD063368268B2"
	// data["imei"] = "DE66EDC092CA8FE5F2A0A52FB8F31782"

	var dataJSON []byte
	if dataJSON, err = json.Marshal(data); err != nil {
		return
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(dataJSON))
	if err != nil {
		return nil, err
	}
	c.SetDappToken()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return c.doRequest(req)
}

// Get 从服务端拉取数据
func (c *HTTPClient) Get(url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Mishop-Client-Id", MISHOPCLIENTID)
	return c.doRequest(req)
}

// DoRequest 给服务端发出请求
func (c *HTTPClient) DoRequest(req *http.Request) (resp *http.Response, err error) {
	return c.doRequest(req)
}

// doRequest 给服务端发出请求
func (c *HTTPClient) doRequest(req *http.Request) (resp *http.Response, err error) {
	if c.cookie != nil {
		req.AddCookie(c.cookie)
	}
	return c.client.Do(req)
}
