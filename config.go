package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/config"
)

type Config struct {
	Addr     string 
	LogDir   string 
	LogLevel string
	cnfPath  string
	innerCnf config.Configer
}

func NewConfig() (*Config) {
	cnf := &Config{}
	return cnf
}

func (o *Config) Load(cnfPath string) error {
	var err error
	o.cnfPath = cnfPath

	err = o.Reload()
	if err != nil {
		return err
	}

	return nil
}

func (o *Config) Reload() error {
	var err error
	if o.cnfPath != "" {
		o.innerCnf, err = config.NewConfig("ini", o.cnfPath)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		o.innerCnf = config.NewFakeConfig()
	}

	o.LogDir = o.innerCnf.DefaultString("log::dir", "../log")
	o.LogLevel = o.innerCnf.DefaultString("log::level", "info")

	return nil
}
