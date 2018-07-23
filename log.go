package main

import (
	log "github.com/Sirupsen/logrus"
	rotater "github.com/firnsan/file_rotater"
	"io"
	stdlog "log"
	"os"
	"os/exec"
)

func InitLog() error {
	var err error
	level, err := log.ParseLevel(gApp.Cnf.LogLevel)
	if err != nil {
		log.Errorf("parse log level failed: %s", err)
		return err
	}

	fw, err := rotater.NewFileRotater(gApp.Cnf.LogDir + "/app.log")
	if err != nil {
		log.Errorf("set log failed: %s", err)
		return err
	}

	log.SetOutput(fw)
	log.SetLevel(level)
	log.AddHook(NewAlarmHook())

	// Also need to set the stdlog's output to this writer
	w := log.StandardLogger().Writer()
	// defer w.Close()
	stdlog.SetOutput(w)

	return nil
}

func UninitLog() {
	// Reset the hooks
	log.StandardLogger().Hooks = make(log.LevelHooks)

	// Close Writer
	w := log.StandardLogger().Out
	log.SetOutput(os.Stderr)
	if wc, ok := w.(io.Closer); ok {
		wc.Close()
	}

	log.Printf("uninit log success")
}

type AlarmHook struct {
}

func NewAlarmHook() *AlarmHook {
	return &AlarmHook{}
}

func (o *AlarmHook) Fire(entry *log.Entry) error {
	msg, err := entry.String()
	if err != nil {

	}

	switch entry.Level {
	case log.PanicLevel:
		fallthrough
	case log.FatalLevel:
		return o.alarm(msg, "")
	case log.ErrorLevel:
		return o.alarm(msg, "")
	default:
		return nil
	}
}

func (o *AlarmHook) Levels() []log.Level {
	return []log.Level{log.ErrorLevel, log.FatalLevel, log.PanicLevel}
}

func (o *AlarmHook) alarm(msg string, alarmId string) error {
	var err error
	if msg == "" || alarmId == "" {
		return nil
	}
	log.Printf("execute alarm cmd, msg: %s, id: %s", msg, alarmId)
	cmd := exec.Command("./alarm.sh", msg, alarmId)
	if err = cmd.Run(); err != nil {
		log.Printf("execute alarm cmd error: %s", err)
		return err
	}
	return nil
}
