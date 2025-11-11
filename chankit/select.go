package chankit

import "context"

func Take[T any](ctx context.Context, in <-chan T, count int, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		taken := 0

		for {
			if taken >= count {
				return
			}

			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				select {
				case <-ctx.Done():
					return
				case outChan <- val:
					taken++
				}
			}
		}
	}()

	return outChan
}

func Skip[T any](ctx context.Context, in <-chan T, count int, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		skipped := 0

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				if skipped < count {
					skipped++
					continue
				}

				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}
	}()

	return outChan
}

func TakeWhile[T any](ctx context.Context, in <-chan T, predicate func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				if !predicate(val) {
					return
				}

				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}

	}()

	return outChan
}

func SkipWhile[T any](ctx context.Context, in <-chan T, predicate func(T) bool, opts ...ChanOption[T]) <-chan T {
	outChan := applyChanOptions(opts...)

	go func() {
		defer close(outChan)
		skipping := true

		for {
			select {
			case <-ctx.Done():
				return

			case val, ok := <-in:
				if !ok {
					return
				}

				if skipping {
					if predicate(val) {
						continue
					}
					skipping = false
				}

				select {
				case <-ctx.Done():
					return
				case outChan <- val:
				}
			}
		}
	}()

	return outChan
}
