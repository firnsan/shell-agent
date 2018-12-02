package main

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"net"
	"net/http"
	"sync"
)

type HttpServer struct {
	ln             net.Listener
	s              http.Server
	wg             *sync.WaitGroup
	pending        int32
	stopped        bool
	started        bool
	initializers   []func() error
	uninitializers []func()
}

func NewHttpServer() *HttpServer {
	return &HttpServer{
		wg: new(sync.WaitGroup),
	}
}

var (
	gHttpServer = NewHttpServer()
)

func (o *HttpServer) Init() error {
	var err error
	for _, f := range o.initializers {
		err = f()
		if err != nil {
			return err
		}
	}
	return nil
}

func (o *HttpServer) Uninit() {
	for _, f := range o.initializers {
		f()
	}
}

func (o *HttpServer) Run() error {

	var err error
	if o.started {
		msg := fmt.Sprintf("http server start repeatly")
		log.Errorln(msg)
		return errors.New(msg)
	}

	err = o.Init()
	if err != nil {
		log.Errorf("http server init failed: %s", err)
		return err
	}
	defer o.Uninit()
	// Prepare the middleware and http handlers
	n := negroni.New()
	mux := ServeMux()
	n.UseFunc(RecoveryMiddleware)
	n.UseFunc(LoggerMiddleware)
	n.UseFunc(CutServiceMiddleware)
	n.UseHandler(mux)

	o.s.Handler = n

	o.ln, err = net.Listen("tcp", gApp.Cnf.Addr)
	if err != nil {
		log.Errorf("listen %s failed: %s", gApp.Cnf.Addr, err)
		return err
	}

	log.Printf("http server serving addr: %s", gApp.Cnf.Addr)
	o.started = true
	o.wg.Add(1)
	err = o.s.Serve(o.ln)
	log.Errorf("http server quit: %s", err)
	o.wg.Done()

	o.started = false
	if o.stopped {
		// Stop in plan, so return with no error
		o.stopped = false
		return nil
	}
	return err
}

func (o *HttpServer) Stop() {
	if !o.started {
		return
	}
	o.stopped = true
	o.ln.Close()
	o.wg.Wait()
}

func (o *HttpServer) AddToInit(f func() error) {
	o.initializers = append(o.initializers, f)
}

func (o *HttpServer) AddToUninit(f func()) {
	o.uninitializers = append(o.uninitializers, f)
}

func ServeMux() *http.ServeMux {
	apiUrlPrefix := "/api/v1"
	mux := http.NewServeMux()

	mux.HandleFunc(apiUrlPrefix+"/cmd/run", RunCmdHandler)
	mux.HandleFunc(apiUrlPrefix+"/cmd/query", QueryCmdHandler)
	mux.HandleFunc(apiUrlPrefix+"/cmd/list", ListCmdHandler)
	mux.HandleFunc(apiUrlPrefix+"/cmd/cancel", CancelCmdHandler)
	mux.HandleFunc(apiUrlPrefix+"/status/mem", StatusMemHandler)

	return mux
}
