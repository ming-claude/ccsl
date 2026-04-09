package render

import "testing"

func BenchmarkDisplayWidth_ASCII(b *testing.B) {
	s := "the quick brown fox jumps over"
	b.ResetTimer()
	for b.Loop() {
		DisplayWidth(s)
	}
}

func BenchmarkDisplayWidth_CJK(b *testing.B) {
	s := "你好世界测试用例中文字符串宽"
	b.ResetTimer()
	for b.Loop() {
		DisplayWidth(s)
	}
}

func BenchmarkDisplayWidth_ANSI(b *testing.B) {
	s := "\x1b[31mhello\x1b[0m \x1b[32mworld\x1b[0m \x1b[1m\x1b[34mbold blue\x1b[0m"
	b.ResetTimer()
	for b.Loop() {
		DisplayWidth(s)
	}
}
