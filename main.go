package main

import (
	_ "fmt"
	"github.com/firnsan/incubator"
)

var (
	gApp = NewApplication()
)

func main() {
	// Incubate this program, including the command-line options, the OS signals
	incubator.Incubate(gApp)

	gApp.Run()
}
