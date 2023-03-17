package main

import (
	"fmt"
	"testing"
)

func BenchmarkFmtSprinf(b *testing.B) {
	var str string
	for n := 0; n < b.N; n++ {
		str = fmt.Sprintf("%s%s", str, "a")
	}
}
