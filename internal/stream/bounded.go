package stream

import (
	"context"
	"sync/atomic"
)

// Bounded is a drop-oldest channel used to cap memory for live streams.
type Bounded[T any] struct {
	ch      chan T
	dropped atomic.Int64
}

func NewBounded[T any](size int) *Bounded[T] {
	if size <= 0 {
		size = 100
	}
	return &Bounded[T]{ch: make(chan T, size)}
}

func (b *Bounded[T]) Send(ctx context.Context, value T) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	select {
	case b.ch <- value:
		return true
	default:
		select {
		case <-b.ch:
		default:
		}
		b.dropped.Add(1)
		select {
		case b.ch <- value:
			return true
		case <-ctx.Done():
			return false
		}
	}
}

func (b *Bounded[T]) C() <-chan T {
	return b.ch
}

func (b *Bounded[T]) Dropped() int {
	return int(b.dropped.Load())
}
