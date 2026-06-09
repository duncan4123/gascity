package main

import (
	"fmt"
	"sync"
	"testing"

	"github.com/gastownhall/gascity/internal/beads"
)

func TestOrderDispatchTrackingIndexConcurrentGatesAreRaceFree(t *testing.T) {
	idx := newOrderDispatchTrackingIndex()
	stores := []beads.Store{beads.NewMemStore()}

	const (
		goroutines = 64
		keyFanout  = 8
		iterations = 16
	)

	var wg sync.WaitGroup
	wg.Add(goroutines)
	start := make(chan struct{})
	for n := 0; n < goroutines; n++ {
		go func(n int) {
			defer wg.Done()
			<-start
			for i := 0; i < iterations; i++ {
				storeKeys := []string{fmt.Sprintf("store-%d", (n+i)%keyFanout)}
				scopedName := fmt.Sprintf("order-%d", (n+i)%keyFanout)
				if _, err := idx.hasOpenTracking(stores, storeKeys, scopedName); err != nil {
					t.Errorf("hasOpenTracking: %v", err)
				}
				lastRunFn := idx.lastRunFunc(stores, storeKeys, nil)
				if _, err := lastRunFn(scopedName); err != nil {
					t.Errorf("lastRunFunc: %v", err)
				}
			}
		}(n)
	}
	close(start)
	wg.Wait()
}
