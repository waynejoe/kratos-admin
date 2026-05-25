package conf

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Bootstrap struct {
	Server     Server     `yaml:"server"`
	Data       Data       `yaml:"data"`
	Security   Security   `yaml:"security"`
	Permission Permission `yaml:"permission"`
}

type Server struct {
	HTTP HTTP `yaml:"http"`
	GRPC GRPC `yaml:"grpc"`
}

type HTTP struct {
	Addr string `yaml:"addr"`
}

type GRPC struct {
	Addr string `yaml:"addr"`
}

type Data struct {
	Global GlobalData `yaml:"global"`
}

type GlobalData struct {
	AdminDatabase Database `yaml:"admin_database"`
	Redis         Redis    `yaml:"redis"`
}

type Database struct {
	Driver string `yaml:"driver"`
	Source string `yaml:"source"`
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int32  `yaml:"db"`
}

type Security struct {
	JwtSecret    string   `yaml:"jwt_secret"`
	JwtTTL       int64    `yaml:"jwt_ttl"`
	JwtSkipPaths []string `yaml:"jwt_skip_paths"`
	ThirdPaths   []string `yaml:"third_paths"`
}

type Permission struct {
	ClosePermission     bool     `yaml:"close_permission"`
	SkipPaths           []string `yaml:"skip_paths"`
	DataPermissionPaths []string `yaml:"data_permission_paths"`
}

func Load(path string) (*Bootstrap, error) {
	bs, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out Bootstrap
	if err := yaml.Unmarshal(bs, &out); err != nil {
		return nil, err
	}
	if out.Server.HTTP.Addr == "" {
		out.Server.HTTP.Addr = ":18080"
	}
	if out.Security.JwtTTL == 0 {
		out.Security.JwtTTL = 1209600
	}
	return &out, nil
}
