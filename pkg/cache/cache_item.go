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
)

type cacheItem struct {
	mutex   sync.RWMutex
	value   Closeable
	expires *time.Time
}

func newCacheItem(value Closeable, ttl time.Duration) *cacheItem {
	cacheItem := &cacheItem{value: value}
	cacheItem.updateTTL(ttl)
	return cacheItem
}

func (item *cacheItem) updateTTL(ttl time.Duration) {
	item.mutex.Lock()
	expiration := time.Now().Add(ttl)
	item.expires = &expiration
	item.mutex.Unlock()
}

func (item *cacheItem) expired() bool {
	item.mutex.RLock()
	result := item.expires != nil && item.expires.Before(time.Now())
	item.mutex.RUnlock()
	return result
}
