package chankit

import (
	"context"
	"sync"
)

func Merge[T any](ctx context.Context, chans ...<-chan T) <-chan T {
	outChan := make(chan T)

	var wg sync.WaitGroup
	for _, ch := range chans {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for val := range ch {
				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(outChan)
	}()

	return outChan
}

func Zip[T, R any](ctx context.Context, ch1 <-chan T, ch2 <-chan R) <-chan struct {
	First  T
	Second R
} {
	outChan := make(chan struct {
		First  T
		Second R
	})

	go func() {
		defer close(outChan)
		for {
			var val1 T
			var val2 R
			var ok1, ok2 bool

			select {
			case <-ctx.Done():
				return
			case val1, ok1 = <-ch1:
				if !ok1 {
					return
				}
			}

			select {
			case <-ctx.Done():
				return
			case val2, ok2 = <-ch2:
				if !ok2 {
					return // Second channel closed
				}
			}

			select {
			case <-ctx.Done():
				return
			case outChan <- struct {
				First  T
				Second R
			}{First: val1, Second: val2}:
			}
		}
	}()

	return outChan
}
