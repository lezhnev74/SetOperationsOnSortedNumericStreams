# Operations On Streaming Sorted Arrays

MIT Licensed.
This package contains function to apply set operations on streams of sorted numbers.

Main motivation of this package is to be used in search engines where big inverted indexes must be used to find matching
documents. Imagine a query `find docs where contains "letter" and not contains "message"`. This operation maps to making
an intersection of the two inverted indexes for "letter" and "message". Now if the indexes are sorted... here comes this
package.

Operations:

- union (returns the stream consisting of elements that are either in stream1 or stream2)
- intersection (returns the stream consisting of elements that are in both stream1 and stream2)
- difference (returns the stream consisting of elements that are in stream1 but not in stream2)

Features:

- generics to support any ordered number type
- asc/desc orders supported
- streaming support is added to reduce memory usage for potentially big data sources

## Sample

With this tool it is possible to arrange complex pipelines.

```go
// see tests
// expression: (a and b) and not c
func TestComposition(t *testing.T) {
a := NewSliceStream([]int{1, 2, 3})
b := NewSliceStream([]int{2, 3})
c := NewSliceStream([]int{3})
result := Diff[int](Intersect[int](a, b, true), c, true)
require.EqualValues(t, []int{2}, ToSlice(result))
}
```

## Usage

- model your data with `SortedNumbersStream` interface (see `SliceStream` for reference)
- decide if your sorted data goes asc or desc
- apply operations on your data `Union[<your data type>](stream1, stream2, <isAsc>): SortedNumbersStream`
- use `ToSlice` to dump your iterable to a slice (for testing/debugging)