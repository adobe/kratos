/*

Copyright 2020 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.

*/

package controller

import (
	"time"

	"k8s.io/client-go/util/workqueue"
)

type FixedItemIntervalRateLimiter struct {
	interval time.Duration
}

// When returns the interval of the rate limiter
func (r *FixedItemIntervalRateLimiter) When(item interface{}) time.Duration {
	return r.interval
}

// NumRequeues returns back how many failures the item has had
func (r *FixedItemIntervalRateLimiter) NumRequeues(item interface{}) int {
	return 1
}

// Forget indicates that an item is finished being retried.
func (r *FixedItemIntervalRateLimiter) Forget(item interface{}) {
}

// NewFixedItemIntervalRateLimiter creates a new instance of a RateLimiter using a fixed interval
func NewFixedItemIntervalRateLimiter(interval time.Duration) workqueue.RateLimiter {
	return &FixedItemIntervalRateLimiter{
		interval: interval,
	}
}
