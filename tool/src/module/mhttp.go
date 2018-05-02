package module

import (
	"bdlib/bdjson"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"time"
)

var cookie = &http.Cookie{
	Name: "demoapp_test",
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 300 * time.Second,
		}).Dial,
		MaxIdleConnsPerHost: 1000,
	},
}

func init() {
	go func() {
		rand.Seed(time.Now().Unix())
		for {
			randChan <- rand.Int()
		}
	}()
}

// Mhttp _
type Mhttp struct {
	Service  string `json:"service"`
	Host     string `json:"host"`
	Platform string `json:"platform"`
	Cid      string `json:"cid"`
	Etag     string `json:"etag"`
	Debug    bool   `json:"debug"`
	AppID    string `json:"appid"`
	AppKey   string `json:"appkey"`
}

func (m *Mhttp) newRequest(method string, requrl string, body url.Values) (req *http.Request, err error) {
	req, err = http.NewRequest(method, requrl, bytes.NewBufferString(body.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return
}

func (m *Mhttp) getResp(req *http.Request) (content []byte, err error) {

	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	if resp.StatusCode > 400 {
		fmt.Printf("request server except:%s\n", resp.Status)
		return
	}

	defer resp.Body.Close()
	content, _ = ioutil.ReadAll(resp.Body)
	if m.Debug {
		for k, v := range resp.Header {
			fmt.Println(k, v)
		}
	}
	obj, err := bdjson.NewJSON(content)
	if err != nil {
		return content, nil
	}
	ncontent, err := json.MarshalIndent(obj.Data(), " ", "   ")
	if err != nil {
		return content, err
	}
	return ncontent, nil
}

var randChan = make(chan int, 1024)
