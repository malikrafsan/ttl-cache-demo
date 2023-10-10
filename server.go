package main

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"

	"github.com/jellydator/ttlcache/v2"
)

var (
	notFound = ttlcache.ErrNotFound
	isClosed = ttlcache.ErrClosed
)

func server() {

	newItemCallback := func(key string, value interface{}) {
		fmt.Printf("New key(%s) added\n", key)
	}
	checkExpirationCallback := func(key string, value interface{}) bool {
		return key != "key1"
	}

	expirationCallback := func(key string, reason ttlcache.EvictionReason, value interface{}) {
		fmt.Printf("This key(%s) has expired because of %s\n", key, reason)
	}

	loaderFunction := func(key string) (data interface{}, ttl time.Duration, err error) {
		ttl = time.Second * 300
		data, err = getFromNetwork(key)

		return data, ttl, err
	}

	cache := ttlcache.NewCache()
	cache.SetTTL(time.Duration(10 * time.Second))
	cache.SetExpirationReasonCallback(expirationCallback)
	cache.SetLoaderFunction(loaderFunction)
	cache.SetNewItemCallback(newItemCallback)
	cache.SetCheckExpirationCallback(checkExpirationCallback)
	cache.SetCacheSizeLimit(2)

	cache.Set("key", "value")
	cache.SetWithTTL("keyWithTTL", "value", 10*time.Second)

	if value, exists := cache.Get("key"); exists == nil {
		fmt.Printf("Got value: %v\n", value)
	}
	count := cache.Count()
	if result := cache.Remove("keyNNN"); result == notFound {
		fmt.Printf("Not found, %d items left\n", count)
	}

	cache.Set("key6", "value")
	cache.Set("key7", "value")
	metrics := cache.GetMetrics()
	fmt.Printf("Total inserted: %d\n", metrics.Inserted)

	cache.Close()

	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for t := range ticker.C {
			log.Printf("Tick at: %v\n", t.UTC())
			// do something
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
	})

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
