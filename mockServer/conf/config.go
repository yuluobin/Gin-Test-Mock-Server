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
}

type RouteModel struct {
	Route string `yaml:"route"`

	Method string `yaml:"method"`

	Responses []Response `yaml:"res"`

	ErrBody map[string]interface{} `yaml:"err_body"`
}

type Response struct {
	URI string `yaml:"uri"`

	Header string `yaml:"header"`

	PostBody map[string]string `yaml:"post_body"`

	RetBody map[string]interface{} `yaml:"ret_body"`
}
