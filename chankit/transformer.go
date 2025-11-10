package chankit

import "context"

func Map[T, R any](ctx context.Context, in <-chan T, mapFunc func(T) R, opts ...ChanOption[R]) <-chan R {
	outChan := applyChanOptions(opts...)
	go func() {
		defer close(outChan)
		for val := range in {
			select {
			case <-ctx.Done():
				return
			case outChan <- mapFunc(val):
			}
		}
	}()
	return outChan
}

func Filter[T any](ctx context.Context, in <-chan T, filterFunc func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		for val := range in {
			select {
			case <-ctx.Done():
				return
			default:
				ok := filterFunc(val)
				if ok {
					select {
					case <-ctx.Done():
						return
					case outChan <- val:
					}
				}
			}
		}
	}()
	return outChan
}

func Reduce[T, R any](ctx context.Context, in <-chan T, reduceFunc func(R, T) R, initial R) R {
	accumulator := initial
	for {
		select {
		case <-ctx.Done():
			return accumulator
		case val, ok := <-in:
			if !ok {
				return accumulator
			}
			accumulator = reduceFunc(accumulator, val)
		}
	}
}
