package atomiccache

// import (
// 	"sync"
// 	"time"
// )

// // LookupTable ...
// type LookupTable struct {
// 	sync.RWMutex
// 	table []LookupRecord
// }

// // LookupRecord ...
// type LookupRecord struct {
// 	Key         uint64
// 	OrderKey    int64
// 	ShardIndex  uint32
// 	RecordIndex uint32
// }

// // NewLookupTable ...
// func NewLookupTable() *LookupTable {
// 	return &LookupTable{}
// }

// // Set ...
// func (l *LookupTable) Set(key uint64, si, ri uint32, expire time.Duration) {
// 	ok := l.getOrderKey(expire)

// 	l.Delete(key)

// 	l.Lock() // Lock for writing and reading
// 	for index, record := range l.table {
// 		if ok < record.OrderKey {
// 			l.table = append(l.table, LookupRecord{})
// 			copy(l.table[index+1:], l.table[index:])
// 			l.table[index] = LookupRecord{Key: key, OrderKey: ok, ShardIndex: si, RecordIndex: ri}
// 			l.Unlock() // Unlock for writing and reading
// 			return
// 		}
// 	}
// 	l.table = append(l.table, LookupRecord{Key: key, OrderKey: ok, ShardIndex: si, RecordIndex: ri})
// 	l.Unlock() // Unlock for writing and reading
// }

// // Get ...
// func (l *LookupTable) Get(key uint64) (si, ri uint32, ok bool) {
// 	l.RLock() // Lock for reading
// 	for _, record := range l.table {
// 		if key == record.Key {
// 			si = record.ShardIndex
// 			ri = record.RecordIndex
// 			l.RUnlock() // Unlock for reading
// 			return si, ri, true
// 		}
// 	}
// 	l.RUnlock() // Unlock for reading

// 	return 0, 0, false
// }

// // Pop ...
// func (l *LookupTable) Pop() (si, ri uint32, ok bool) {
// 	var record LookupRecord

// 	l.Lock() // Lock for writing and reading

// 	if len(l.table) == 0 {
// 		l.Unlock() // Unlock for writing and reading
// 		return 0, 0, false
// 	} else if len(l.table) == 1 {
// 		record, l.table = l.table[0], nil
// 	} else {
// 		record, l.table = l.table[0], l.table[1:]
// 	}

// 	l.Unlock() // Unlock for writing and reading
// 	return record.ShardIndex, record.RecordIndex, true
// }

// // Delete ...
// func (l *LookupTable) Delete(key uint64) {
// 	l.Lock() // Lock for writing and reading
// 	for index, record := range l.table {
// 		if key == record.Key {
// 			copy(l.table[index:], l.table[index+1:])
// 			l.table[len(l.table)-1] = LookupRecord{}
// 			l.table = l.table[:len(l.table)-1]
// 			l.Unlock() // Unlock for writing and reading
// 			return
// 		}
// 	}
// 	l.Unlock() // Unlock for writing and reading
// }

// // getOrderKey ...
// func (l *LookupTable) getOrderKey(expire time.Duration) int64 {
// 	return time.Now().Add(expire).UnixNano()
// }
