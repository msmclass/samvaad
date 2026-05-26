package utils

import (
	"time"

	"github.com/jellydator/ttlcache/v3"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

const (
	iceConfigTTLMin = 5 * time.Minute
)

type IceConfigCache[T comparable] struct {
	c *ttlcache.Cache[T, *samvaad.ICEConfig]
}

func NewIceConfigCache[T comparable](ttl time.Duration) *IceConfigCache[T] {
	cache := ttlcache.New(
		ttlcache.WithTTL[T, *samvaad.ICEConfig](max(ttl, iceConfigTTLMin)),
		ttlcache.WithDisableTouchOnHit[T, *samvaad.ICEConfig](),
	)
	go cache.Start()

	return &IceConfigCache[T]{cache}
}

func (icc *IceConfigCache[T]) Stop() {
	icc.c.Stop()
}

func (icc *IceConfigCache[T]) Put(key T, iceConfig *samvaad.ICEConfig) {
	icc.c.Set(key, iceConfig, ttlcache.DefaultTTL)
}

func (icc *IceConfigCache[T]) Get(key T) *samvaad.ICEConfig {
	if it := icc.c.Get(key); it != nil {
		return it.Value()
	}
	return &samvaad.ICEConfig{}
}


