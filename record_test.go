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

func benchmarkRecordNew(size uint32, b *testing.B) {
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		NewRecord(size)
	}
}

func BenchmarkRecordNewSmall(b *testing.B) {
	benchmarkRecordNew(512, b)
}

func BenchmarkRecordNewMedium(b *testing.B) {
	benchmarkRecordNew(1024, b)
}

func BenchmarkRecordNewLarge(b *testing.B) {
	benchmarkRecordNew(2048, b)
}

func benchmarkRecordSet(size uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	record := NewRecord(size)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		record.Set(data)
	}
}

func BenchmarkRecordSetSmall(b *testing.B) {
	benchmarkRecordSet(512, b)
}

func BenchmarkRecordSetMedium(b *testing.B) {
	benchmarkRecordSet(1024, b)
}

func BenchmarkRecordSetLarge(b *testing.B) {
	benchmarkRecordSet(2048, b)
}

func benchmarkRecordGet(size uint32, b *testing.B) {
	b.ReportAllocs()

	var data []byte
	record := NewRecord(size)
	record.Set(data)

	for i := uint32(0); i < size; i++ {
		data = append(data, 1)
	}

	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		record.Get()
	}
}

func BenchmarkRecordGetSmall(b *testing.B) {
	benchmarkRecordGet(512, b)
}

func BenchmarkRecordGetMedium(b *testing.B) {
	benchmarkRecordGet(1024, b)
}

func BenchmarkRecordGetLarge(b *testing.B) {
	benchmarkRecordGet(2048, b)
}
