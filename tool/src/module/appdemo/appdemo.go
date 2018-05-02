package appdemo

import (
	"encoding/json"
	"io/ioutil"
	"module/base"
)

// MioneModel _
var (
	Model = base.BaseInfo{}
)

func init() {
	data, err := ioutil.ReadFile("./src/module/appdemo/api_data.json")
	if err != nil {
		panic(err)
	}

	if err := json.Unmarshal(data, &Model); err != nil {
		panic(err)
	}
}
