package conf

//server config will used in file variables.go

type configModel struct {
	Server *serverModel `yaml:"server"`
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
