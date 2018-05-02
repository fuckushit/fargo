package module

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"module/appdemo"
	"module/base"
	"net/http"
	"net/url"
	"os"

	"strings"

	"bdlib/util"
)

// Var _
var (
	Token = "cookie test"
	Debug = false
)

// Module _
func Module(module, action string) (actions []base.Modules, appid, appkey, service string, err error) {

	switch module {
	case "appdemo":
		if action == "" {
			for _, ac := range appdemo.Model.Actions {
				actions = append(actions, ac)
			}
		} else {
			ac, found := appdemo.Model.Actions[action]
			if !found {
				fmt.Printf("module:[%s] not found action:[%s]\n", module, action)
				return
			}
			actions = append(actions, ac)
		}
		service = appdemo.Model.Service
	default:
		err = fmt.Errorf("module:[%s] not defined", module)
	}

	return
}

// Prepare _
func Prepare(actions []base.Modules, params map[string]string) (err error) {
	var errs []error
	for _, action := range actions {
		for k, v := range params {
			action.Params[k] = v
		}
		for _, mustkey := range action.MustArgs {
			if _, ok := action.Params[mustkey]; !ok {
				errs = append(errs, fmt.Errorf("too few arguments must-args: %s", mustkey))
			}
		}
	}

	if len(errs) != 0 {
		err = fmt.Errorf("%v", errs)
		return
	}

	if len(actions) == 1 && len(params) != 0 {
		actions[0].Params = params
	}
	return
}

// RunAction _
func RunAction(mcli *Mhttp, action base.Modules) (code, desc string) {
	requrl := fmt.Sprintf("%s/%s", mcli.Host, action.RPath)
	fmt.Printf("Host: %s\n", requrl)

	var (
		req *http.Request
		err error
	)

	for k, v := range action.Params {
		if strings.Index(k, "file2_") == 0 && v != "" {
			f, err := os.Open(v)
			if err != nil {
				fmt.Println(err)
				return
			}
			defer f.Close()
			content, err := ioutil.ReadAll(f)
			if err != nil {
				fmt.Println(err)
				return
			}
			action.Params[k[6:]] = util.Base64Encode(content)
			action.Params[k] = ""
		}
	}

	body := url.Values{}
	for k, v := range action.Params {
		body.Set(k, v)
	}
	req, err = mcli.newRequest("POST", requrl, body)
	if err != nil {
		fmt.Println(err)
		return
	}

	if Token != "" {
		cookie.Value = Token
		req.AddCookie(cookie)
	}

	type Header struct {
		Code interface{} `json:"code"`
		Desc string      `json:"desc"`
		Info string      `json:"Info"`
	}

	content, err := mcli.getResp(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	var head Header
	if err := json.Unmarshal(content, &head); err != nil {
		fmt.Printf("err:%v url:%s, head:%v\n", err, requrl, head)
		return
	}

	fmt.Println(strings.Repeat("=", 50))
	if util.String(head.Code) != "0" {
		fmt.Printf("resp:%#v xxxxxxxx\n", head)
	} else if !Debug {
		fmt.Printf("resp:%#v length:%d\n", head, len(content))
	} else {
		fmt.Printf("result: %s\n", content)
	}
	fmt.Println(strings.Repeat("=", 50))

	code = fmt.Sprintf("%v", head.Code)
	desc = head.Desc

	return
}
