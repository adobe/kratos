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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testCloseable struct {
}

func (t *testCloseable) Close() error {
	return nil
}

var _ = Describe("CacheItem", func() {
	It("Cache item expired", func() {
		cacheItem := newCacheItem(&testCloseable{}, time.Millisecond*100)
		<-time.After(time.Millisecond * 400)
		Expect(cacheItem.expired()).To(BeTrue(), "cache item should be expired after TTL")
	})

	It("Empty TTL considered NOT expired", func() {
		cacheItem := cacheItem{value: &testCloseable{}}
		Expect(cacheItem.expired()).To(BeFalse(), "cache item with Nil TTL is not expired")
	})

	It("Cache item NOT expired after update", func() {
		cacheItem := newCacheItem(&testCloseable{}, time.Millisecond*100)
		<-time.After(time.Millisecond * 400)
		Expect(cacheItem.expired()).To(BeTrue(), "cache item should be expired after TTL")

		cacheItem.updateTTL(time.Hour)
		<-time.After(time.Second)
		Expect(cacheItem.expired()).To(BeFalse(), "cache item should NOT be expired after TTL update")
	})
})
