package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"
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
}

func NewCache(config *config) *cache {
	c := ttlcache.NewCache()
	// c.SkipTTLExtensionOnHit(true)

	campaignCaller := NewCampaignCaller()

	// c.SetExpirationReasonCallback(func(key string, reason ttlcache.EvictionReason, value interface{}) {
	// 	c.Set(key, value)
	// })

	return &cache{
		cacheEngine: c,
		config:      config,
		caller:      campaignCaller,
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

	loader := func(key string) (interface{}, time.Duration, error) {
		validTypes, ok := flowType[key]
		if !ok {
			return nil, 0, fmt.Errorf("invalid key")
		}

		campaigns, err := c.caller.LiveByTypes(ctx, validTypes...)
		if err != nil {
			return nil, 0, err
		}

		return campaigns, c.config.globalTTL, nil
	}

	campaigns, err := c.cacheEngine.GetByLoader(key, loader)
	if err != nil {
		return nil, err
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
