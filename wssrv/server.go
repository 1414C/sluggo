package wssrv

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/1414C/sluggo/wscom"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// CacheServInt outlines the cache server interface
type CacheServInt interface {
	Init() error
	initAndSet(as []wscom.Article) error
	set(a wscom.Article) error // ok?
	get(a *wscom.Article) error
	invalidate(key string) error
	flush() error

	processCmdChannel()
	SetHandler(a wscom.Article) error
	Serve(port uint)
	IsAlive() bool
}

// CacheServ is the server access struct
type CacheServ struct {
	cacheMap   map[string]wscom.Article
	chin       chan wscom.Article
	chout      chan wscom.Article
	count      int
	HTTPServer *http.Server
	CacheServInt
}

// init an empty cache server
func (cs *CacheServ) init() error {
	if cs.cacheMap == nil {
		cs.cacheMap = make(map[string]wscom.Article)

		// initialize the command/serialization channel
		cs.chin = make(chan wscom.Article)
		cs.chout = make(chan wscom.Article)

		// start blocking on the command/serialization channel
		go cs.processCmdChannel()
		return nil
	}
	return fmt.Errorf("cache already initialized - use Flush() instead")
}

// processCmdChannel is used to process commands from the command/serialization channel
func (cs *CacheServ) processCmdChannel() {

	// block on the channel
	for v := range cs.chin {
		// fmt.Println(v)
		cs.count++
		fmt.Println("cs.count:", cs.count)

		switch v.Op {
		case "AU":
			cs.cacheMap[v.Key] = v
			cs.chout <- v
		case "G":
			// if the get finds nothing, clear the key to prevent match-checking error
			v, ok := cs.cacheMap[v.Key]
			if !ok {
				v.Key = ""
			}
			cs.chout <- v
		case "D":
			delete(cs.cacheMap, v.Key)
			_, ok := cs.cacheMap[v.Key]
			if !ok {
				cs.chout <- v
			} else {
				log.Printf("warning: delete of user-cache record %v failed\n", v.Key)
			}
		case "FLUSH":
			log.Println("executing FLUSH...")
			cs.cacheMap = make(map[string]wscom.Article)
			log.Printf("cs.cacheMap contains %d entries\n", len(cs.cacheMap))
			cs.chout <- v
		default:
			// do nothing

		}
	}
}

// Flush the local cache
func (cs *CacheServ) flush() error {
	cs.cacheMap = nil
	cs.cacheMap = make(map[string]wscom.Article)
	return nil
}

// InitAndSet inits and preloads a cache server from a slice of Articles
func (cs *CacheServ) initAndSet(as []wscom.Article) error {
	return nil
}

// Set or overwrite a value in the cache
func (cs *CacheServ) set(a wscom.Article) error {
	if cs.cacheMap != nil {
		cs.cacheMap[a.Key] = a
		return nil
	}
	return fmt.Errorf("cache not initialized - call Init() first")
}

// Get the Article referenced in a.Key from the cache
func (cs *CacheServ) get(a *wscom.Article) bool {
	// var ok bool
	if cs.cacheMap != nil {
		article, ok := cs.cacheMap[a.Key]
		if !ok {
			return false
		}
		if article.Valid == false {
			return false
		}
		*a = article
		return true
	}
	return false
}

// Invalidate the Article referenced by the key
func (cs *CacheServ) invalidate(key string) error {
	if cs.cacheMap != nil {
		a, ok := cs.cacheMap[key]
		if !ok {
			return nil
		}
		a.Valid = false
		cs.cacheMap[a.Key] = a
		return nil
	}
	return fmt.Errorf("cache not initialized - call Init() first")
}

// Serve starts the cache server
func (cs *CacheServ) Serve(port string) {
	err := cs.init()
	if err != nil {
		panic("Serve()" + err.Error())
	}

	mux := http.NewServeMux()
	mux.Handle("/set", websocket.Handler(cs.SetHandler))
	mux.Handle("/get", websocket.Handler(cs.GetHandler))
	mux.Handle("/delete", websocket.Handler(cs.DeleteHandler))
	mux.Handle("/flush", websocket.Handler(cs.FlushHandler))
	mux.Handle("/isalive", websocket.Handler(cs.IsAliveHandler))

	// http.Handle("/set", websocket.Handler(cs.SetHandler))
	// http.Handle("/get", websocket.Handler(cs.GetHandler))
	fmt.Printf("listening for ws traffic on %s\n", port)

	// create the ws server - this can be stopped by calling cs.httpServer.Shutdown(...)
	cs.HTTPServer = &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		err := cs.HTTPServer.ListenAndServe()
		if err != nil {
			// this will write something in a Shutdown scenario too
			log.Printf("cs.httpServer.ListenAndServe() error - got: %s", err)
		}
	}()
}

// SetHandler pushes the AU Article into the serialization channel
func (cs *CacheServ) SetHandler(ws *websocket.Conn) {

	// gob decoding
	var a wscom.Article
	var msg = make([]byte, 1024)
	l, err := ws.Read(msg)
	if err != nil {
		// error thing
		log.Println("ws.Read() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	m := msg[0:l]
	decBuf := bytes.NewBuffer(m)
	err = gob.NewDecoder(decBuf).Decode(&a)
	if err != nil {
		// error thing
		log.Println("gob.Decode() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	// fmt.Println("Article:", a)

	// put the Article command into the channel
	// cs.addCommand(a)
	cs.chin <- a

	a2 := <-cs.chout
	if a2.Key == a.Key {
		// fmt.Println("KEY MATCH")
		// fmt.Println("a2.Key:", a2.Key)
		// fmt.Println("a.Key: ", a.Key)
		log.Printf("Add/Update of Article %v succeeded\n", a)
	} else {
		log.Println()
		log.Println()
		log.Println("KEY FAILURE TO MATCH")
		panic("KEY FAILURE TO MATCH")
	}

	// fmt.Println(cs.cacheMap)
	_, err = ws.Write([]byte("true"))
	if err != nil {
		log.Println("error:", err)
	}
}

// GetHandler pushes the G Article into the serialization channel
func (cs *CacheServ) GetHandler(ws *websocket.Conn) {

	// gob decoding
	var a wscom.Article
	var msg = make([]byte, 1024)
	l, err := ws.Read(msg)
	if err != nil {
		// error thing
		log.Println("CacheServ.GetHandler() ws.Read() error - got:", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	m := msg[0:l]
	decBuf := bytes.NewBuffer(m)
	err = gob.NewDecoder(decBuf).Decode(&a)
	if err != nil {
		// error thing
		log.Println("CacheServ.GetHandler() gob.Decode() error - got:", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	// fmt.Println("Article:", a)

	// put the Article command into the channel
	cs.chin <- a

	// get the Article command response from the channel
	a2 := <-cs.chout

	// if a match was found, or no match was found send the appropriate
	// repsonse back to the caller.
	if a2.Key == a.Key || a2.Key == "" {
		encBuf := new(bytes.Buffer)
		err = gob.NewEncoder(encBuf).Encode(a2)
		if err != nil {
			log.Println("CacheServ.GetHandler(): gob failed to encode the Article")
		}

		value := encBuf.Bytes()
		_, err = ws.Write(value)
		if err != nil {
			fmt.Println("CacheServ.GetHandler() error writing response - got:", err)
		}
		return
	}
	log.Println()
	log.Println()
	log.Println("KEY FAILURE TO MATCH")
	log.Println(a2.Key)
	log.Println(a.Key)
	panic("KEY FAILURE TO MATCH")
}

// DeleteHandler pushes the D Article into the serialization channel
func (cs *CacheServ) DeleteHandler(ws *websocket.Conn) {

	// gob decoding
	var a wscom.Article
	var msg = make([]byte, 1024)
	l, err := ws.Read(msg)
	if err != nil {
		// error thing
		log.Println("ws.Read() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	m := msg[0:l]
	decBuf := bytes.NewBuffer(m)
	err = gob.NewDecoder(decBuf).Decode(&a)
	if err != nil {
		// error thing
		log.Println("gob.Decode() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	// fmt.Println("Article for Deletion:", a)

	// put the Article command into the channel
	cs.chin <- a

	a2 := <-cs.chout
	if a2.Key == a.Key {
		// fmt.Println("KEY MATCH")
		// fmt.Println("a2.Key:", a2.Key)
		// fmt.Println("a.Key: ", a.Key)
		log.Printf("Delete of Article %v succeeded\n", a)
	} else {
		log.Println()
		log.Println()
		log.Println("KEY FAILURE TO MATCH")
		panic("KEY FAILURE TO MATCH")
	}

	// fmt.Println(cs.cacheMap)
	_, err = ws.Write([]byte("true"))
	if err != nil {
		log.Println("error:", err)
	}
}

// FlushHandler flushes the cache
func (cs *CacheServ) FlushHandler(ws *websocket.Conn) {

	// gob decoding
	var a wscom.Article
	var msg = make([]byte, 1024)
	l, err := ws.Read(msg)
	if err != nil {
		// error thing
		log.Println("ws.Read() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	m := msg[0:l]
	decBuf := bytes.NewBuffer(m)
	err = gob.NewDecoder(decBuf).Decode(&a)
	if err != nil {
		// error thing
		log.Println("gob.Decode() error", err)
		_, err = ws.Write([]byte("false"))
		return
	}
	fmt.Println("Article for Deletion:", a)

	// put the Article command into the channel
	cs.chin <- a

	a2 := <-cs.chout
	if a2.Key == a.Key {
		// fmt.Println("KEY MATCH")
		// fmt.Println("a2.Key:", a2.Key)
		// fmt.Println("a.Key: ", a.Key)
		fmt.Printf("Delete of Article %v succeeded\n", a)
	} else {
		fmt.Println()
		fmt.Println()
		fmt.Println("KEY FAILURE TO MATCH")
		panic("KEY FAILURE TO MATCH")
	}

	// fmt.Println(cs.cacheMap)
	_, err = ws.Write([]byte("true"))
	if err != nil {
		fmt.Println("error:", err)
	}
}

// IsAliveHandler can be used to determine whether the cache server is
// running, without performing a read/write operation.
func (cs *CacheServ) IsAliveHandler(ws *websocket.Conn) {

	var err error

	if cs.cacheMap != nil {
		_, err = ws.Write([]byte("true"))
	} else {
		_, err = ws.Write([]byte("false"))
	}
	if err != nil {
		log.Println("error:", err)
	}
}
