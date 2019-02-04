package atomiccache

import (
	"reflect"
	"testing"
)

func TestRecordSimple(t *testing.T) {
	for _, c := range []struct {
		size uint32
		in   []byte
		want []byte
	}{
		{10, []byte{0}, []byte{0}},
		{10, []byte{0, 1, 2, 3, 4, 5}, []byte{0, 1, 2, 3, 4, 5}},
		{1, []byte{0, 1, 2}, []byte{0}},
	} {
		record := NewRecord(c.size)
		record.Set(c.in)
		if !reflect.DeepEqual(record.Get(), c.want) {
			t.Errorf("[%d] %v != %v", c.size, record.Get(), c.want)
		}
	}
}
