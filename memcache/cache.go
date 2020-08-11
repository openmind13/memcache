package memcache

import (
	"errors"
	"sync"
	"time"
)

var (
	errKeyNotFound  = errors.New("Key not found")
	errItemNotFound = errors.New("Item not found")
	errTimeExpired  = errors.New("Item lifetime has expired")
)

// Cache struct
type Cache struct {
	mutex sync.RWMutex

	defaultExpiration time.Duration
	cleanupInterval   time.Duration

	count uint
	data  map[string]item
}

// item struct
type item struct {
	value      interface{}
	createdAt  time.Time
	expiration int64
}

// New - create new cache
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	data := make(map[string]item)
	cache := Cache{
		data:              data,
		defaultExpiration: 10 * time.Second,
		cleanupInterval:   20 * time.Second,
		count:             0,
	}

	if cache.cleanupInterval > 0 {
		cache.GC()
	}

	return &cache
}

// Add new value into cache
func (cache *Cache) Add(key string, value interface{}, duration time.Duration) {
	var expiration int64
	if duration == 0 {
		duration = cache.defaultExpiration
	}

	if duration > 0 {
		expiration = time.Now().Add(duration).UnixNano()
	}

	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	cache.data[key] = item{
		value:      value,
		expiration: expiration,
		createdAt:  time.Now(),
	}
	cache.count++
}

// Get value from cache by key
func (cache *Cache) Get(key string) (interface{}, error) {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	item, found := cache.data[key]
	if !found {
		return nil, errItemNotFound
	}

	if item.expiration > 0 {
		if time.Now().UnixNano() > item.expiration {
			return nil, errTimeExpired
		}
	}

	return item.value, nil
}

// Delete value by key
func (cache *Cache) Delete(key string) error {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	if _, found := cache.data[key]; !found {
		return errKeyNotFound
	}

	delete(cache.data, key)
	cache.count--

	return nil
}

// Count return count of items in cache
func (cache *Cache) Count() uint {
	return cache.count
}

// Exist check exist item or not
func (cache *Cache) Exist(key string) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	_, found := cache.data[key]
	return found
}

// GC start cleaning cache
func (cache *Cache) GC() {
	go cache.gc()
}

func (cache *Cache) gc() {
	for {
		<-time.After(cache.cleanupInterval)

		if cache.data == nil {
			return
		}

		if keys := cache.expiredKeys(); len(keys) != 0 {
			cache.clearExpiredItems(keys)
		}
	}
}

// expiredKeys return slice of strings with expired lifetime
func (cache *Cache) expiredKeys() []string {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()

	var keys []string

	for k, i := range cache.data {
		if time.Now().UnixNano() > i.expiration && i.expiration > 0 {
			keys = append(keys, k)
		}
	}

	return keys
}

func (cache *Cache) clearExpiredItems(keys []string) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	for _, k := range keys {
		delete(cache.data, k)
		cache.count--
	}
}
