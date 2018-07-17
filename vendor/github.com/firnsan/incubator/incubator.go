package incubator

import (
	"github.com/docopt/docopt-go"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type incubator struct {
	app Application
}

func newIncubator(app Application) *incubator {
	return &incubator{app: app}
}

func (o *incubator) handleSignal(c chan os.Signal) {
	var err error
	for {
		s := <-c
		log.Printf("Signal received: %d\n", s)
		switch s {
		case syscall.SIGUSR1:
			err = o.app.OnReload()
			if err != nil {
				log.Fatalf("App OnReload failed: %s", err)
			}
		case syscall.SIGUSR2:
			fallthrough
		case syscall.SIGINT:
			fallthrough
		case syscall.SIGTERM:
			o.app.OnStop()
			os.Exit(1)
			// If the whole process exits from here, then something is wrong 
		}
	}
}

func (o *incubator) incubate() {

	opts, err := docopt.Parse(o.app.GetUsage(), nil, true, o.app.GetVersion(), false)
	if err != nil {
		log.Fatalf("Parse cmd options failed: %s", err)
	}

	o.app.OnOptParsed(opts)

	// one slot per signal
	sc := make(chan os.Signal, 4)
	signal.Notify(sc, syscall.SIGUSR1, syscall.SIGUSR2, syscall.SIGINT, syscall.SIGTERM)

	go o.handleSignal(sc)

}
