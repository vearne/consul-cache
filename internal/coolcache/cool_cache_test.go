package coolcache

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestSetGet(t *testing.T) {
	cache := NewCoolCache(10, 5*time.Minute, 5*time.Minute)
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		value := key
		cache.Set(key, value, 0)
	}
	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)
		expected := key
		actual, ok := cache.Get(key)
		assert.Equal(t, ok, true)
		assert.Equal(t, expected, actual)
	}
	_, ok := cache.Get("1000")
	assert.Equal(t, ok, false)
}
