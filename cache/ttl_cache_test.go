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

const (
	key = "testKey"
)

var _ = Describe("TTLCache", func() {
	It("Cache item expired", func() {
		cache := NewTTLCache("test-cache", time.Millisecond*500)

		cache.Put(key, &testCloseable{})

		cachedValue, found := cache.Get(key)
		Expect(cachedValue).NotTo(BeNil(), "cache item present")
		Expect(found).To(BeTrue())

		<-time.After(time.Millisecond * 600)
		cachedValue, found = cache.Get(key)
		Expect(cachedValue).To(BeNil(), "cache item not found after expiration")
		Expect(found).To(BeFalse())
	})
})
