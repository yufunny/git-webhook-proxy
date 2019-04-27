package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SubscribeConfig struct {
	Url       string `yaml:"url"`
	Branch    string `yaml:"branch"`
	Action    string `yaml:"action"`
	Directory string `yaml:"directory"`
}

type SystemConfig struct {
	Server     string            `yaml:"server"`
	Subscribes []SubscribeConfig `yaml:"subscribes"`
	Mode       string            `yaml:"mode"`
}

var config *SystemConfig

// LoadConfig  加载系统配置
func LoadConfig(file string) (*SystemConfig, error) {
	b, e := ioutil.ReadFile(file)
	if nil != e {
		return nil, errors.New("Config->Read config file[" + file + "] error; " + e.Error())
	}
	config = &SystemConfig{}
	e = yaml.Unmarshal(b, config)
	if nil != e {
		return nil, errors.New("Config->Unmarshal config from config file[" + file + "] error; " + e.Error())
	}
	return config, nil
}

func GetConfig() *SystemConfig {
	return config
}
