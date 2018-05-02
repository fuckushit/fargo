/*
* Author: chenzhifeng01@baidu.com
* Date:   2018-04-29 14:26
* Description: http测试模块
* ./install httpclient && bin/httpclient -h http://127.0.0.1:6060 -m appdemo -a test1 param1=11 param2=22
* ./install httpclient && bin/httpclient -h http://127.0.0.1:6060 -m appdemo -a test2 param1=11 param2=22
 */

package main

import (
	"flag"
	"fmt"
	"module"
	"strings"
	"time"
)

var (
	host   = flag.String("h", "", "host")
	model  = flag.String("m", "", "model")
	action = flag.String("a", "", "action")
	debug  = flag.Bool("d", true, "debug")
)

func main() {

	flag.Parse()
	if *model == "" || *host == "" {
		fmt.Println("host and model not is null")
		return
	}
	if !strings.HasPrefix(*host, "http://") {
		*host = "http://" + *host
	}

	actions, appid, appkey, service, err := module.Module(*model, *action)
	fmt.Println(actions, appid, appkey, service, err)
	if err != nil {
		fmt.Println(err)
		return
	}

	module.Debug = *debug
	params := map[string]string{}

	userArgs := flag.Args()
	for _, kv := range userArgs {
		kvlist := strings.Split(kv, "=")
		if len(kvlist) != 2 {
			fmt.Printf("args format is k=v %s\n", kv)
			return
		}
		params[kvlist[0]] = kvlist[1]
	}
	if err = module.Prepare(actions, params); err != nil {
		fmt.Println(err)
		return
	}

	be := time.Now()
	mcli := new(module.Mhttp)
	mcli.Host = *host
	mcli.AppID = appid
	mcli.AppKey = appkey
	mcli.Service = service
	mcli.Debug = *debug

	resList := []ResultInfo{}
	for _, action := range actions {
		ab := time.Now()
		code, desc := module.RunAction(mcli, action)
		resList = append(resList, ResultInfo{
			code: code,
			desc: desc,
			cost: time.Now().Sub(ab).String(),
		})
	}

	fmt.Printf("client run finish cost:%s\n", time.Now().Sub(be).String())
	fmt.Println("-------------cost detial---------------")

	fmt.Println(strings.Repeat("-", 50))
	for id, action := range actions {
		fmt.Printf("| action:%s\t |code:%s, desc:%s, cost:%s|\n",
			action.RPath, resList[id].code, resList[id].desc, resList[id].cost)
		fmt.Println(strings.Repeat("-", 50))
	}
}

// ResultInfo _
type ResultInfo struct {
	code string
	desc string
	cost string
}
