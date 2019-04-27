package config

import (
	"errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type ListenConfig struct {
	Http   string `yaml:"http"`
	Socket string `yaml:"socket"`
}

type SystemConfig struct {
	Listen ListenConfig `yaml:"listen"`
	Mode   string       `yaml:"mode"`
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
