package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/msmclass/samvaad/pkg/proto/samvaad"
)

func TestIceConfigCache(t *testing.T) {
	cache := NewIceConfigCache[string](10 * time.Second)
	t.Cleanup(cache.Stop)

	cache.Put("test", &samvaad.ICEConfig{})
	require.NotNil(t, cache)
}


