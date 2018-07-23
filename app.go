package main

import (
	log "github.com/Sirupsen/logrus"
	"time"
)

var (
	VERSION = "0.1.0"
)

type Application struct {
	Cnf     *Config
	cnfPath string
}

func NewApplication() *Application {
	return &Application{Cnf: NewConfig()}
}

func (o *Application) GetVersion() string {
	return VERSION
}

func (o *Application) GetUsage() string {
	return `Shell Agent.
This agnet is a program installed on remote host, help you to execute shell command on the remote host.
This agent can also help to transport file to/from remote host.

Usage:
	shell-agent [--cnf=<path>] [--addr=<addr>]
	shell-agent -h | --help
	shell-agent --version

Options:
	--cnf=<path>  config file path [default: ].
	--addr=<addr>  config file path [default: :8080].`

}

func (o *Application) OnOptParsed(m map[string]interface{}) {
	o.cnfPath = m["--cnf"].(string)
	o.Cnf.Addr = m["--addr"].(string)
}

func (o *Application) OnReload() error {
	var err error
	log.Warn("application need to reload")

	// Reload config
	if o.Cnf != nil {
		err = o.Cnf.Reload()
	}

	// Reload the logger
	UninitLog()
	err = InitLog()
	if err != nil {
		return err
	}

	log.Warn("application reloaded")

	return nil
}

func (o *Application) OnStop() {
	log.Warn("application need to stop")
	gHttpServer.Stop()
	// Here is in signal handler routine, we don't want to exit from signal handler routine.
	// So, we sleep 1 second to give some opportunity for main routine to exit normally.
	time.Sleep(time.Second)
	log.Warn("application stopped")
}

func (o *Application) Run() {
	var err error

	// Load the config
	err = o.Cnf.Load(o.cnfPath)
	if err != nil {
		log.Fatalf("init config failed: %s", err)
	}

	// Initialize the logger
	err = InitLog()
	if err != nil {
		log.Fatalf("init log failed: %s", err)
	}
	defer UninitLog()

	log.Print("")
	log.Print("application started")

	// Run the http server
	err = gHttpServer.Run()
	if err != nil {
		log.Fatalf("application quited, because http server quited abnormally: %s", err)
	}

	log.Warn("application quited")

}
