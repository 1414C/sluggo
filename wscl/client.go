package wscl

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"

	"github.com/1414C/sluggo/wscom"
	"golang.org/x/net/websocket"
)

// AddUpdCacheEntry adds or updates a cache entry using the value of interface i
// and the string key.  It is the callers responsbility to manage the key
// structure.  For a struct type of Foo[], the function should be called as
// follows:
//
// f := Foo{F1: "value1", F2: "value2",}
// k := "myUniqueKey1234"
// err := AddUpdCacheEntry(k,f)
func AddUpdCacheEntry(key string, i interface{}) error {

	// first gob encode the data
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(i)
	if err != nil {
		log.Println("failed to gob-encode interface{} i - got:", err)
		return err
	}
	b := encBuf.Bytes()

	// create an Article to house the encoded data
	a := wscom.Article{
		Key:   key,
		Op:    "AU",
		Valid: true,
		Value: b,
		Type:  "",
	}

	// gob-encode the Article
	encBuf = new(bytes.Buffer)
	err = gob.NewEncoder(encBuf).Encode(a)
	if err != nil {
		log.Println("failed to gob-encode the Article - got:", err)
		return err
	}
	encArticle := encBuf.Bytes()

	// connect to remote server
	origin := "http://localhost/"
	url := "ws://192.168.1.82:7070/set"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println("AddUpdCacheEntry() ws connection failed - got:", err)
		return err
	}

	// push the encoded Article
	_, err = ws.Write(encArticle)
	if err != nil {
		log.Println("AddUpdCacheEntry() ws.Write error - got:", err)
		return err
	}

	var msg = make([]byte, 64)

	// single read from the ws is okay here
	n, err := ws.Read(msg)
	if err != nil {
		log.Println("GetCacheEntry() ws.Read error - got:", err)
		return err
	}

	// if update is confirmed do a little dance =)
	if string(msg[:n]) == "true" {
		// cw <- na
	} else {
		return fmt.Errorf("AddUpdCacheEntry() appeared to fail - got %v(raw),%v(string)", msg[:n], string(msg[:n]))
	}
	return nil
}

// GetCacheEntry reads the specified entry from the cache.  If the entry does
// not exist, an error will be set.  For a struct type of Foo{}, the function
// should be called as follows:
//
// f := &Foo{}
// k := "myUniqueKey1234"
// err := GetCacheEntry(k, f)
//
// followng the call, f will contain the cached value of the Foo{} type if
// the read was successful.
func GetCacheEntry(key string, i interface{}) error {

	// create an Article shell
	a := wscom.Article{
		Key:   key,
		Op:    "G",
		Valid: false,
		Value: nil,
		Type:  "",
	}

	// gob-encode the Article shell
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(a)
	if err != nil {
		log.Println("GetCacheEntry() failed to gib encode the Article Shell")
	}
	encArticle := encBuf.Bytes()

	origin := "http://localhost/"
	url := "ws://192.168.1.82:7070/get"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println("GetCacheEntry() failed to make websocket connection with cache - got:", err)
		return err
	}

	// push the encoded Article (command Op=='G' is what counts here)
	_, err = ws.Write(encArticle)
	if err != nil {
		log.Println("GetCacheEntry() ws.Write error:", err)
		return err
	}

	// read from the ws
	var msg = make([]byte, 1024)
	var n int
	readBuf := new(bytes.Buffer)
	for {
		n, err = ws.Read(msg)
		if err != nil {
			log.Println("AddUpdCacheEntry() ws.Read error - got:", err)
			return err
		}
		readBuf.Write(msg)
		if n < 1024 {
			break
		}
	}

	// decode the read Article
	decBuf := bytes.NewBuffer(msg[:n])
	err = gob.NewDecoder(decBuf).Decode(&a)
	if err != nil {
		// error thing
		log.Println("GetCacheEntry() gob.Decode() error - got:", err)
		return err
	}
	if a.Value != nil {
		decBuf := bytes.NewBuffer(a.Value)
		err := gob.NewDecoder(decBuf).Decode(i)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("GetCacheEntry() Article ID %s not in cache", key)
	}
	return nil
}

// RemoveCacheEntry removes the specified entry from the cache if it exists.
// If the specified entry does not exist, no error is returned, as the cache
// is deemd to be in the correct state.  An error will be returned only if
// the function was not able to complete the operation from a technical
// standpoint.
func RemoveCacheEntry(key string) error {

	// create an Article shell
	a := wscom.Article{
		Key:   key,
		Op:    "D",
		Valid: false,
		Value: nil,
		Type:  "",
	}

	// gob-encode the Article shell
	encBuf := new(bytes.Buffer)
	err := gob.NewEncoder(encBuf).Encode(a)
	if err != nil {
		log.Fatalf("DELETE gob failed to encode the Article Shell")
	}
	value := encBuf.Bytes()

	origin := "http://localhost/"
	url := "ws://192.168.1.82:7070/delete"

	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println("GetCacheEntry() failed to make websocket connection with cache - got:", err)
		return err
	}

	// push the encoded Article (command Op=='D' is what counts here)
	_, err = ws.Write(value)
	if err != nil {
		log.Fatal("DELETE ws.Write error:", err)
	}
	var msg = make([]byte, 1024)

	// read the bool result from the ws
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal("ws.Read error:", err)
	}

	// if delete is confirmed, update local cache
	if string(msg[:n]) != "true" {
		log.Println("Article deletion failed")
	}
	return nil
}
