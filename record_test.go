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

func TestRecordFree(t *testing.T) {
	size := uint32(10)
	want := []byte{0, 1, 2}

	record := NewRecord(size)
	record.Set(want)
	record.Free()
	if !reflect.DeepEqual(record.Get(), []byte{}) {
		t.Errorf("[%d] %v != %v", 10, record.Get(), want)
	}
}

func BenchmarkRecordNew512(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewRecord(512)
	}
}

func BenchmarkRecordNew1024(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewRecord(1024)
	}
}

func BenchmarkRecordNew2048(b *testing.B) {
	for n := 0; n < b.N; n++ {
		NewRecord(1024)
	}
}

func BenchmarkRecordSet512(b *testing.B) {
	var data []byte
	size := uint32(512)
	record := NewRecord(size)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Set(data)
	}
}

func BenchmarkRecordSet1024(b *testing.B) {
	var data []byte
	size := uint32(1024)
	record := NewRecord(size)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Set(data)
	}
}

func BenchmarkRecordSet2048(b *testing.B) {
	var data []byte
	size := uint32(2048)
	record := NewRecord(size)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Set(data)
	}
}

func BenchmarkRecordGet512(b *testing.B) {
	var data []byte
	size := uint32(512)
	record := NewRecord(size)
	record.Set(data)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Get()
	}
}

func BenchmarkRecordGet1024(b *testing.B) {
	var data []byte
	size := uint32(1024)
	record := NewRecord(size)
	record.Set(data)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Get()
	}
}

func BenchmarkRecordGet2048(b *testing.B) {
	var data []byte
	size := uint32(2048)
	record := NewRecord(size)
	record.Set(data)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	for n := 0; n < b.N; n++ {
		record.Get()
	}
}
