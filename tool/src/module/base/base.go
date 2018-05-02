package base

// Modules 模块结构
type Modules struct {
	RepeatNum int64             `json:"repeatNum"`
	MustArgs  []string          `json:"must_args"`
	Params    map[string]string `json:"params"`
	RPath     string            `json:"path"`
}

// BaseInfo _
type BaseInfo struct {
	AppID   string             `json:"appid"`
	AppKey  string             `json:"appkey"`
	Service string             `json:"service"`
	Actions map[string]Modules `json:"actions"`
}
