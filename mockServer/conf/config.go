package conf

//server config will used in file variables.go

type configModel struct {
	Server *serverModel `yaml:"server"`

	Func []*RouteModel `yaml:"func"`
}

//serverModel get server information from config.yml

type serverModel struct {
	Mode string `yaml:"mode"` // run mode

	Build string `yaml:"build"`

	Port string `yaml:"port"` // server port

	ZKPort string `yaml:"ZkAddr"`

	BlockSize uint64 `yaml:"BlockSize"`

	NumReplicas int `yaml:"NumReplicas"`

	NumTapestry int `yaml:"NumTapestry"`
}

type RouteModel struct {
	Route string `yaml:"route"`

	Method string `yaml:"method"`

	Responses []Response `yaml:"res"`
}

type Response struct {
	URI string `yaml:"uri"`

	Header string `yaml:"header"`

	PostBody map[string]interface{} `yaml:"post_body"`

	RetBody map[string]interface{} `yaml:"ret_body"`

	ErrBody map[string]interface{} `yaml:"err_body"`
}
