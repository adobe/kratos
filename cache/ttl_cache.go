/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package cache

import (
	"sync"
	"time"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type Closeable interface {
	Close() error
}

type TTLCache struct {
	mutex sync.RWMutex
	ttl   time.Duration
	items map[string]*cacheItem
	log   logr.Logger
}

func NewTTLCache(cacheName string, ttl time.Duration) *TTLCache {
	cache := &TTLCache{
		ttl:   ttl,
		items: make(map[string]*cacheItem),
		log:   log.Log.WithName(cacheName),
	}
	cache.startEvictionThread()
	return cache
}

func (cache *TTLCache) Put(key string, value Closeable) {
	cache.mutex.Lock()
	cache.items[key] = newCacheItem(value, cache.ttl)
	cache.mutex.Unlock()
}

func (cache *TTLCache) Get(key string) (interface{}, bool) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()

	item, exists := cache.items[key]
	if !exists || item.expired() {
		return nil, false
	}

	item.updateTTL(cache.ttl)
	return item.value, true
}

func (cache *TTLCache) evictExpiredItems() {
	cache.mutex.Lock()
	for key, item := range cache.items {
		if item.expired() {
			cache.log.Info("expiring cache item", "key", key)
			delete(cache.items, key)
			_ = item.value.Close()

		}
	}
	cache.mutex.Unlock()
}

func (cache *TTLCache) startEvictionThread() {
	cache.log.Info("starting eviction thread")
	ticker := time.Tick(cache.ttl)
	go (func() {
		for {
			select {
			case <-ticker:
				cache.evictExpiredItems()
			}
		}
	})()
}
