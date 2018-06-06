# sluggo

## What is this?
Sluggo is a simple cache server with a very small API, no automatic eviction and no high-watermark for memory usage.  Sluggo uses string keys for clarity, but this should probably be updated internally to use something like xxhash to create unique uint64 keys for speed(?) and key structure consistency.

I wrote this quickly to support some testing I was working on.  Be mindful if you choose to use this for anything important.

## Use
go run main.go -a 192.168.1.40:7070

## API
The caller API is contained in the wscl package and consists of three discrete functions as shown in the following code snippet:
```golang

    func AddUpdCacheEntry(key string, i interface{}, address string) error {}
    func GetCacheEntry(key string, i interface{}, address string) error {}
    func RemoveCacheEntry(key string, address string) error {}

```

Interface{} is used as a passing/receiving reference-type parameter in order to allow any data to be placed into the cache.  It is the reponsibility of the caller to determine the best way to use the API. i.e. call with a static-type, or call with interface-type (read-case) and then perform a type-assertion.

- consider multicast like memcached?

https://gist.github.com/scottjbarr/255828

https://github.com/memcached/memcached/wiki


