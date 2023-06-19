package sorted_numeric_streams

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSliceStream(t *testing.T) {
	s1 := NewSliceStream([]int{1, 2, 3})
	s2 := ToSlice[int](s1)
	require.EqualValues(t, s2, []int{1, 2, 3})
}

func TestChannelStream(t *testing.T) {
	s1 := NewChannelStream[int]()
	go func() {
		s1.Push(1)
		s1.Push(2)
		s1.Push(3)
		s1.Close()
	}()

	fetchedData := ToSlice[int](s1)
	require.EqualValues(t, []int{1, 2, 3}, fetchedData)
}

func TestUnion(t *testing.T) {
	type test struct {
		a, b, result []int
		asc          bool
	}
	tests := []test{
		// asc
		{[]int{}, []int{}, []int{}, true},
		{[]int{}, []int{1}, []int{1}, true},
		{[]int{1}, []int{1}, []int{1}, true},
		{[]int{1}, []int{2}, []int{1, 2}, true},
		{[]int{1}, []int{0, 2}, []int{0, 1, 2}, true},
		{[]int{1, 2, 3}, []int{0}, []int{0, 1, 2, 3}, true},
		{[]int{1}, []int{0, 1, 2, 3}, []int{0, 1, 2, 3}, true},
		// desc
		{[]int{}, []int{}, []int{}, false},
		{[]int{}, []int{1}, []int{1}, false},
		{[]int{1}, []int{1}, []int{1}, false},
		{[]int{1}, []int{2}, []int{2, 1}, false},
		{[]int{1}, []int{2, 0}, []int{2, 1, 0}, false},
		{[]int{3, 2, 1}, []int{0}, []int{3, 2, 1, 0}, false},
		{[]int{1}, []int{3, 2, 1, 0}, []int{3, 2, 1, 0}, false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			a := NewSliceStream(tt.a)
			b := NewSliceStream(tt.b)
			c := Union[int](a, b, tt.asc)
			require.EqualValues(t, tt.result, ToSlice(c))
		})
	}
}

func TestIntersection(t *testing.T) {
	type test struct {
		a, b, result []int
		asc          bool
	}
	tests := []test{
		// asc
		{[]int{}, []int{}, []int{}, true},
		{[]int{}, []int{1}, []int{}, true},
		{[]int{1}, []int{1}, []int{1}, true},
		{[]int{1}, []int{2}, []int{}, true},
		{[]int{1}, []int{0, 1}, []int{1}, true},
		{[]int{0, 1, 2}, []int{1, 2, 3}, []int{1, 2}, true},
		// desc
		{[]int{}, []int{}, []int{}, false},
		{[]int{}, []int{1}, []int{}, false},
		{[]int{1}, []int{1}, []int{1}, false},
		{[]int{1}, []int{2}, []int{}, false},
		{[]int{1}, []int{1, 0}, []int{1}, false},
		{[]int{2, 1, 0}, []int{3, 2, 1}, []int{2, 1}, false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			a := NewSliceStream(tt.a)
			b := NewSliceStream(tt.b)
			c := Intersect[int](a, b, tt.asc)
			require.EqualValues(t, tt.result, ToSlice(c))
		})
	}
}

func TestDiff(t *testing.T) {
	type test struct {
		a, b, result []int
		asc          bool
	}
	tests := []test{
		// asc
		{[]int{}, []int{}, []int{}, true},
		{[]int{}, []int{1}, []int{}, true},
		{[]int{1}, []int{1}, []int{}, true},
		{[]int{1}, []int{2}, []int{1}, true},
		{[]int{1}, []int{0, 1}, []int{}, true},
		{[]int{0, 1}, []int{0, 1, 2}, []int{}, true},
		{[]int{1, 2, 3}, []int{0, 1, 2}, []int{3}, true},
		{[]int{0, 1, 2}, []int{1}, []int{0, 2}, true},
		// desc
		{[]int{}, []int{}, []int{}, false},
		{[]int{}, []int{1}, []int{}, false},
		{[]int{1}, []int{1}, []int{}, false},
		{[]int{1}, []int{2}, []int{1}, false},
		{[]int{1}, []int{0, 1}, []int{}, false},
		{[]int{1, 0}, []int{2, 1, 0}, []int{}, false},
		{[]int{3, 2, 1}, []int{2, 1, 0}, []int{3}, false},
		{[]int{2, 1, 0}, []int{1}, []int{2, 0}, false},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			a := NewSliceStream(tt.a)
			b := NewSliceStream(tt.b)
			c := Diff[int](a, b, true)
			require.EqualValues(t, tt.result, ToSlice(c))
		})
	}
}

func TestComposition(t *testing.T) {
	a := NewSliceStream([]int{1, 2, 3})
	b := NewSliceStream([]int{2, 3})
	c := NewSliceStream([]int{3})
	result := Diff[int](Intersect[int](a, b, true), c, true)
	require.EqualValues(t, []int{2}, ToSlice(result))
}
