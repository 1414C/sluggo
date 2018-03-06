package main

import (
	"fmt"
	"github.com/1414C/sluggo/wssrv"
	"sync"
)

func main() {

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
	fmt.Println("starting cache server...")
	sv.Serve(":7070")
	wg.Wait()
}
