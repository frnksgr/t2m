package t2m

import "testing"

func BenchmarkCPULoop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// this should run for about one millisecond
		cpuloop(1000)
	}
}
