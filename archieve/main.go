package archieve

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v2"
)

const (
	defaultValue = "default"
)

const (
	timelessKey  = "[timeless] key"
	fetchableKey = "[fetchable] key"
	regularKey   = "[regular] key"
)

type config struct {
	globalTTL      time.Duration
	cacheSizeLimit int
}

func randomGenerator() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

type cache struct {
	cacheEngine *ttlcache.Cache
}

func NewCache(
	config *config,
) *cache {
	c := ttlcache.NewCache()

	// callback fn for new item to be added
	newItemCallback := func(key string, value interface{}) {
		fmt.Printf("New key (%s) added\n\t with value: (%v)\n", key, value)
	}
	// // callback fn for expiration checking before deletion
	checkExpirationCallback := func(key string, value interface{}) bool {
		return !strings.HasPrefix(key, "[timeless]")
		// return false
	}
	// callback fn after expiration
	expirationReasonCallback := func(key string, reason ttlcache.EvictionReason, value interface{}) {
		fmt.Printf("This key(%s) has expired because of %s\n", key, reason)

		// c.Set(key, value)
		// go func() {
		// 	value, err := getFromNetwork(key)
		// 	if err != nil {
		// 		log.Printf("Error fetching from network: %v\n", err)
		// 		return
		// 	}
		// 	log.Printf("Set (%v) with value (%v) at: %v\n", key, value, time.Now().UTC())
		// 	c.Set(key, value)
		// }()

		c.Set(key, value)
		time.Sleep(time.Second * 10)
		// value, err := getFromNetwork(key)
		// if err != nil {
		// 	log.Printf("Error fetching from network: %v\n", err)
		// 	return
		// }
		// log.Printf("Set (%v) with value (%v) at: %v\n", key, value, time.Now().UTC())
		fmt.Println("Fetching thru network")
		c.Set(key, randomGenerator())
	}
	// callback fn for loading data on cache miss
	loaderFunction := func(key string) (data interface{}, ttl time.Duration, err error) {
		if !strings.HasPrefix(key, "[fetchable]") {
			return nil, 0, fmt.Errorf("key not fetchable")
		}

		fmt.Println("Fetching thru network")
		ttl = time.Second * 30
		data, err = getFromNetwork(key)

		return data, ttl, err
	}

	c.SetTTL(config.globalTTL)
	c.SetExpirationReasonCallback(expirationReasonCallback)
	c.SetLoaderFunction(loaderFunction)
	c.SetNewItemCallback(newItemCallback)
	c.SetCheckExpirationCallback(checkExpirationCallback)
	c.SetCacheSizeLimit(config.cacheSizeLimit)
	c.SkipTTLExtensionOnHit(true)

	return &cache{
		cacheEngine: c,
	}
}

func (c *cache) Set(key string, value interface{}) error {
	return c.cacheEngine.Set(key, value)
}

func (c *cache) SetWithTTL(key string, value interface{}, ttl time.Duration) error {
	return c.cacheEngine.SetWithTTL(key, value, ttl)
}

func (c *cache) Get(key string) (interface{}, error) {
	return c.cacheEngine.Get(key)
}

func depmain() {
	config := config{
		globalTTL:      time.Duration(30 * time.Second),
		cacheSizeLimit: 10,
	}
	cache := NewCache(&config)

	// ticker := time.NewTicker(20 * time.Second)
	// go func() {
	// 	for t := range ticker.C {
	// 		fmt.Println("Tick at: ", t.UTC())
	// 		key := regularKey
	// 		value, err := getFromNetwork(key)
	// 		if err != nil {
	// 			return
	// 		}

	// 		log.Printf("Set (%v) with value (%v) at: %v\n", key, value, time.Now().UTC())
	// 		cache.Set(key, value)
	// 	}
	// }()

	key := regularKey
	value := randomGenerator()
	log.Printf("Set (%v) with value (%v) at: %v\n", key, value, time.Now().UTC())
	cache.Set(key, value)

	cache.Set(timelessKey, "timeless")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := html.EscapeString(r.URL.Path)
		cleanKey := strings.TrimPrefix(key, "/")

		if value, err := cache.Get(cleanKey); err == nil {
			fmt.Fprintf(w, "%v", value)
			return
		}

		fmt.Fprintf(w, "%v", defaultValue)
	})

	fmt.Println("Listening on port 9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}

func getFromNetwork(key string) (string, error) {
	time.Sleep(time.Second * 3)
	return randomGenerator(), nil
}
