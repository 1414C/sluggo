package main

import (
	"flag"
	"fmt"
	"sync"

	"github.com/1414C/sluggo/wssrv"
)

func main() {

	a := flag.String("a", "127.0.0.1:7070", "address:port that sluggo will accept ws connections on")
	flag.Parse()

	wg := sync.WaitGroup{}
	sv := wssrv.CacheServ{}

	// the waitgroup count is not decremented in sv.Serve, due to the
	// manner in which this codebase is expected to be deployed.
	// Typically the cache server will be run in parallel with a http/
	// rpc set of services therefore sv.Serve(...) starts a go routine
	// to handle incoming cache traffic over websockets then exits.
	// the http/rpc handler would then start its own ListenAndServe on
	// the main go routine, thereby allowing sv.Serve(...) to run.
	wg.Add(1)
	fmt.Printf("starting cache server on %s ...\n", *a)
	sv.Serve(*a)
	wg.Wait()
}
