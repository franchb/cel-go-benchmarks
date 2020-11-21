package iterator

import (
	"sync/atomic"

	benchmarkv1 "github.com/franchb/cel-go-benchmarks/proto/benchmark/v1"
)

type Iterator struct {
	buf    []*benchmarkv1.Message
	length uint64
	index  uint64
}

func New(buf []*benchmarkv1.Message) *Iterator {
	return &Iterator{
		buf:    buf,
		length: uint64(len(buf)),
	}
}

func (it *Iterator) hasNext() bool {
	return atomic.LoadUint64(&it.index) <= it.length
}

func (it *Iterator) Next() *benchmarkv1.Message {
	if ix := it.index; ix >= it.length {
		atomic.CompareAndSwapUint64(&it.index, it.index, 0)
		return it.buf[0]
	} else {
		atomic.AddUint64(&it.index, 1)
		return it.buf[ix]
	}
}
