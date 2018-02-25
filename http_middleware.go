package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/urfave/negroni"
	"net/http"
	"runtime"
	"time"
)

func CutServiceMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	if gHttpServer.stopped {
		log.Print("http server is quiting, ignore this request")
		rw.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	gHttpServer.wg.Add(1)
	next(rw, r)
	gHttpServer.wg.Done()
}

func RecoveryMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	defer func() {
		if err := recover(); err != nil {
			if rw.Header().Get("Content-Type") == "" {
				rw.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
			rw.WriteHeader(http.StatusInternalServerError)
			stack := make([]byte, 1024*8)
			stack = stack[:runtime.Stack(stack, false)]

			f := "PANIC: %s\n%s"
			log.Errorf(f, err, stack)
		}
	}()

	next(rw, r)
}

func LoggerMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	//r.ParseForm()
	//log.Printf("request recived %s %s", r.Method, r.URL.Path, r.Form)
	log.Printf("request recived %s %s", r.Method, r.URL.Path)

	next(rw, r)

	res := rw.(negroni.ResponseWriter)
	log.Printf("request completed %v in %v", res.Status(), time.Since(start))
}
