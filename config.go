package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/astaxie/beego/config"
)

type Config struct {
	Addr        string // 监听地址
	LogDir      string // 日志所在目录
	LogLevel    string // 日志等级
	AdminSecret string
	DbConnect   string
	cnfPath     string
	innerCnf    config.Configer
}

func NewConfig(cnfPath string) (*Config, error) {
	var err error
	cnf := &Config{cnfPath: cnfPath}

	err = cnf.Reload()
	if err != nil {
		return nil, err
	}

	cnf.Addr = cnf.innerCnf.DefaultString("addr", ":8080")
	cnf.DbConnect = cnf.innerCnf.DefaultString("db_connect", "root:@tcp(localhost:3306)/vpn")
	cnf.AdminSecret = cnf.innerCnf.DefaultString("admin_secret", "none")
	return cnf, nil
}

func (o *Config) Reload() error {
	var err error
	o.innerCnf, err = config.NewConfig("ini", o.cnfPath)
	if err != nil {
		log.Error(err)
		return err
	}

	o.LogDir = o.innerCnf.DefaultString("log::dir", "../log")
	o.LogLevel = o.innerCnf.DefaultString("log::level", "info")

	return nil
}
