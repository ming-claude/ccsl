package transcript

import "testing"

func BenchmarkParse(b *testing.B) {
	path := testdataPath("basic.jsonl")
	b.ResetTimer()
	for b.Loop() {
		_, _ = Parse(path)
	}
}
