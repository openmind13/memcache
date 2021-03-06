package memcache

import (
	"errors"
	"sync"
	"time"
)

var (
	errKeyNotFound  = errors.New("key not found")
	errItemNotFound = errors.New("item not found")
	errTimeExpired  = errors.New("item lifetime has expired")
)

type Cache struct {
	mutex             sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	count             uint
	data              map[string]item
	quitChan          chan bool
}
type item struct {
	value      interface{}
	createdAt  time.Time
	expireAt   time.Time
	expiration int64
}

// New - create new cache
func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	data := make(map[string]item)
	cache := Cache{
		data:              data,
		defaultExpiration: 10 * time.Second,
		cleanupInterval:   10 * time.Second,
		count:             0,
	}
	if cache.cleanupInterval > 0 {
		go cache.gcCollect()
	}
	return &cache
}

func (cache *Cache) Destroy() error {
	cache.data = nil
	cache.count = 0
	cache.quitChan <- true
	return nil
}

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

func (cache *Cache) Get(key string) (interface{}, error) {
	cache.mutex.RLock()
	item, found := cache.data[key]
	cache.mutex.RUnlock()
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

func (cache *Cache) Count() uint {
	return cache.count
}

func (cache *Cache) IsExist(key string) bool {
	cache.mutex.RLock()
	defer cache.mutex.RUnlock()
	_, found := cache.data[key]
	return found
}

func (cache *Cache) gcCollect() {
	for {
		// <-time.After(cache.cleanupInterval)
		time.Sleep(1 * time.Second)
		select {
		case <-cache.quitChan:
			return
		default:
			if keys := cache.expiredKeys(); len(keys) != 0 {
				cache.clearExpiredItems(keys)
			}
		}
	}
}

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
	for _, k := range keys {
		delete(cache.data, k)
		cache.count--
	}
	cache.mutex.Unlock()
}
