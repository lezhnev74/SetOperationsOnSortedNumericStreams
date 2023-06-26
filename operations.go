package sorted_numeric_streams

import "golang.org/x/exp/constraints"

// SortedNumbersStream allows to iterate over sorted data
// Algorithms imply the data behind this interface is sorted
type SortedNumbersStream[T constraints.Ordered] interface {
	// Next return the next available item from the sorted stream
	// ok shows if the stream is drained and no further read will give anything (like a closed channel)
	Next() (item T, ok bool)
}

// operation represent the set operation (union, diff etc)
// since positions of set operands matter, so do operands of this func
// when both are present - means they are equal and found in every set
// otherwise left or right is present reflecting left or right set of the operation (A op B)
type operation[T constraints.Ordered] func(a, b *T)

// An operation can know that no further results will be found
// at which case it should stop reading from streams
// shouldStop checks both streams if any is closed and may decide to stop all processing as pointless
// example: (A AND B) should stop once either stream is drained
// returns TRUE if no further processing is allowed
// Note that stopped processing does not drain streams and separate cleanup required
type shouldStop func(aClosed, bClosed bool) bool

// ChannelStream is used as a result of operation on other streams
type ChannelStream[T constraints.Ordered] struct {
	pipe chan T
}

func (s *ChannelStream[T]) Next() (item T, ok bool) {
	item, ok = <-s.pipe
	return
}

func (s *ChannelStream[T]) Push(item T) { s.pipe <- item }

func (s *ChannelStream[T]) Close() { close(s.pipe) }

func NewChannelStream[T constraints.Ordered]() *ChannelStream[T] {
	return &ChannelStream[T]{
		pipe: make(chan T),
	}
}

// SliceStream implements SortedNumbersStream for static slices (convenient in tests)
type SliceStream[T constraints.Ordered] struct {
	slice []T
	pos   int
}

func (s *SliceStream[T]) Reset() { s.pos = 0 }
func (s *SliceStream[T]) Next() (item T, ok bool) {
	if s.pos < len(s.slice) {
		item = s.slice[s.pos]
		s.pos++
		return item, true
	}
	var empty T // zero initialized
	return empty, false
}
func NewSliceStream[T constraints.Ordered](slice []T) *SliceStream[T] {
	return &SliceStream[T]{
		slice: slice,
		pos:   0,
	}
}

// Union returns the stream consisting of elements that are either in stream1 or stream2
func Union[T constraints.Ordered](stream1, stream2 SortedNumbersStream[T], asc bool) SortedNumbersStream[T] {
	result := NewChannelStream[T]()
	unionOperation := func(a, b *T) {
		// equal: both present
		if a != nil && b != nil {
			result.Push(*a)
			return
		}
		// only left present
		if a != nil {
			result.Push(*a)
			return
		}
		// only right present
		if b != nil {
			result.Push(*b)
			return
		}
	}
	shouldStopDecision := func(aClosed, bClosed bool) bool { return false }

	go func() {
		iterate(stream1, stream2, unionOperation, shouldStopDecision, asc)
		result.Close()
	}()

	return result
}

// Intersect returns the stream consisting of elements that are in both stream1 and stream2
func Intersect[T constraints.Ordered](stream1, stream2 SortedNumbersStream[T], asc bool) SortedNumbersStream[T] {
	result := NewChannelStream[T]()
	unionOperation := func(a, b *T) {
		// equal: both present
		if a != nil && b != nil {
			result.Push(*a)
		}
	}
	shouldStopDecision := func(aClosed, bClosed bool) bool { return aClosed || bClosed }

	go func() {
		iterate(stream1, stream2, unionOperation, shouldStopDecision, asc)
		result.Close()
	}()

	return result
}

// Diff returns the stream consisting of elements that are in stream1 but not in stream2
func Diff[T constraints.Ordered](stream1, stream2 SortedNumbersStream[T], asc bool) SortedNumbersStream[T] {
	result := NewChannelStream[T]()
	unionOperation := func(a, b *T) {
		if a != nil && b == nil {
			result.Push(*a)
		}
	}
	shouldStopDecision := func(aClosed, bClosed bool) bool { return aClosed }

	go func() {
		iterate(stream1, stream2, unionOperation, shouldStopDecision, asc)
		result.Close()
	}()

	return result
}

func iterate[T constraints.Ordered](stream1, stream2 SortedNumbersStream[T], op operation[T], stop shouldStop, asc bool) {
	var (
		i1, i2         T
		empty1, empty2 bool
		readOk         bool
	)
	empty1, empty2 = true, true

	for {
		if empty1 {
			i1, readOk = stream1.Next()
			if !readOk {
				if stop(true, false) {
					return
				}
				for { // no more in stream1 -> return all from stream2
					if !empty2 {
						op(nil, &i2)
					}
					i2, readOk = stream2.Next()
					if !readOk {
						return
					}
					op(nil, &i2)
				}
			}
			empty1 = false
		}

		if empty2 {
			i2, readOk = stream2.Next()
			if !readOk {
				if stop(false, true) {
					return
				}
				for { // no more from stream2 -> return all from stream1
					if !empty1 {
						op(&i1, nil)
						empty1 = true
					}
					i1, readOk = stream1.Next()
					if !readOk {
						return
					}
					op(&i1, nil)
				}
			}
			empty2 = false
		}

		// Both streams have values
		if i1 == i2 {
			op(&i1, &i2)
			empty1, empty2 = true, true
		} else if asc && i1 < i2 {
			op(&i1, nil)
			empty1 = true
		} else if asc && i1 > i2 {
			op(nil, &i2)
			empty2 = true
		} else if !asc && i1 < i2 {
			op(nil, &i2)
			empty2 = true
		} else if !asc && i1 > i2 {
			op(&i1, nil)
			empty1 = true
		}
	}
}

func ToSlice[T constraints.Ordered](stream SortedNumbersStream[T]) []T {
	ret := make([]T, 0)
	for {
		i, ok := stream.Next()
		if !ok {
			break
		}
		ret = append(ret, i)
	}
	return ret
}
