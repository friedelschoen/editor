package syncutil

import (
	"testing"
)

func BenchmarkSyncedQ(b *testing.B) {
	for i := 0; i < b.N; i++ {
		bSyncedQ()
	}
}
func bSyncedQ() {
	sq := NewSyncedQ()
	for i := 0; i < 1000; i++ {
		sq.PushBack(i)
	}
	for i := 0; i < 1000; i++ {
		sq.PopFront()
	}
}
