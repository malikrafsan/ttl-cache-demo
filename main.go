package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/jellydator/ttlcache/v2"
)

var countCallThruService = 0

type config struct {
	globalTTL      time.Duration
	cacheSizeLimit int
	enableCache    bool
}

type CampaignType int

const (
	CampaignRandomizedCashback CampaignType = iota
	CampaignLegacyCampaign
	CampaignRcPartners
	CampaignCoupon
	CampaignPromoCode
	CampaignSkuPromo
	CampaignFeeCampaign
	CampaignAll
)

var flowType map[string][]CampaignType = map[string][]CampaignType{
	"REDEEM_ESTIMATE": {
		CampaignLegacyCampaign, CampaignRandomizedCashback, CampaignRcPartners,
	},
	"FEE_EVALUATE": {
		CampaignFeeCampaign,
	},
	"FEE_APPLY": {
		CampaignFeeCampaign,
	},
}

type campaign struct {
	ID   string
	Name string
	Desc string
}

type campaignCaller struct{}

func NewCampaignCaller() *campaignCaller {
	return &campaignCaller{}
}

func (c *campaignCaller) LiveByTypes(ctx context.Context, campaignTypes ...CampaignType) ([]campaign, error) {
	fmt.Println("calling campaign service...")
	countCallThruService++
	time.Sleep(4 * time.Second)

	id := time.Now().UnixNano()

	campaigns := []campaign{}
	for _, campaignType := range campaignTypes {
		campaigns = append(campaigns, campaign{
			ID:   fmt.Sprintf("id:%d", id),
			Name: fmt.Sprintf("name:%d", campaignType),
			Desc: fmt.Sprintf("desc:%d", campaignType),
		})
	}

	return campaigns, nil
}

type cache struct {
	config      *config
	cacheEngine *ttlcache.Cache
	caller      *campaignCaller
	mutex       sync.Mutex
	loading     map[string]bool
}

func NewCache(config *config) *cache {
	c := ttlcache.NewCache()
	c.SkipTTLExtensionOnHit(true)
	c.SetTTL(config.globalTTL)

	campaignCaller := NewCampaignCaller()

	return &cache{
		cacheEngine: c,
		config:      config,
		caller:      campaignCaller,
		mutex:       sync.Mutex{},
		loading:     map[string]bool{},
	}
}

func (c *cache) LiveByTypes(ctx context.Context, key string) ([]campaign, error) {
	if !c.config.enableCache {
		validTypes, ok := flowType[key]
		if !ok {
			return nil, fmt.Errorf("invalid key")
		}

		return c.caller.LiveByTypes(ctx, validTypes...)
	}

	loader := func(key string) (interface{}, error) {
		validTypes, ok := flowType[key]
		if !ok {
			return nil, fmt.Errorf("invalid key")
		}

		campaigns, err := c.caller.LiveByTypes(ctx, validTypes...)
		if err != nil {
			return nil, err
		}

		return campaigns, nil
	}

	campaigns, err := c.cacheEngine.Get(key)
	if err != nil {
		c.mutex.Lock()
		shouldCall := false
		if val, ok := c.loading[key]; !ok || !val {
			shouldCall = true
			c.loading[key] = true
		}
		c.mutex.Unlock()

		fmt.Printf("shouldCall: %v; %+v\n", shouldCall, c.loading)

		if shouldCall {
			// call campaign service
			// renew cache value
			go func() {
				value, err := loader(key)
				if err != nil {
					log.Printf("Error fetching from network: %v\n", err)
					return
				}

				c.cacheEngine.Set(key, value)
				c.cacheEngine.SetWithTTL(fmt.Sprintf("persist:%s", key), value, time.Hour*24*30) // TODO: make it configurable

				c.mutex.Lock()
				c.loading[key] = false
				c.mutex.Unlock()
			}()
		}

		val, err := c.cacheEngine.Get(fmt.Sprintf("persist:%s", key))
		if err != nil {
			// if not found in cache, call campaign service (synchonously)
			// cons: it will block the request and call multiple times if there are multiple requests
			// alternative: return nil, error

			val, err = loader(key)
			if err != nil {
				return nil, err
			}
			c.cacheEngine.Set(key, val)
			return val.([]campaign), nil
		}
		return val.([]campaign), nil
	}

	return campaigns.([]campaign), nil
}

const (
	cacheMiss = "cache miss"
)

func main() {
	config := &config{
		globalTTL:      10 * time.Second,
		cacheSizeLimit: 10,
		enableCache:    true,
	}
	cache := NewCache(config)
	ctx := context.Background()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// metrics
		fmt.Println("================================")
		// print cache metric
		metrics := cache.cacheEngine.GetMetrics()
		fmt.Printf("metrics: %+v\n", metrics)
		keys := cache.cacheEngine.GetKeys()
		fmt.Printf("keys: %+v\n", keys)
		fmt.Printf("countCallThruService: %d\n", countCallThruService)
		fmt.Println("================================")

		key := html.EscapeString(r.URL.Path)
		cleanKey := strings.TrimPrefix(key, "/")

		if value, err := cache.LiveByTypes(ctx, cleanKey); err == nil {
			fmt.Fprintf(w, "%v", value)
			return
		}

		fmt.Fprintf(w, "%v", cacheMiss)
	})

	fmt.Println("Listening on port 9999")
	log.Fatal(http.ListenAndServe(":9999", nil))
}
