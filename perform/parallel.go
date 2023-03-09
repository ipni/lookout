package perform

import (
	"context"
	"sync"
)

func InParallel[K, V any](
	ctx context.Context,
	parallelism int,
	items []K,
	forEach func(context.Context, K) V) chan V {
	results := make(chan V, 1)
	go func() {
		defer close(results)
		targets := make(chan K, 1)
		var wg sync.WaitGroup
		for i := 0; i < parallelism; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case target, ok := <-targets:
						if !ok {
							return
						}
						select {
						case <-ctx.Done():
							return
						case results <- forEach(ctx, target):
						}
					}
				}
			}()
		}
		for _, item := range items {
			select {
			case <-ctx.Done():
				return
			case targets <- item:
			}
		}
		close(targets)
		wg.Wait()
	}()
	return results
}
