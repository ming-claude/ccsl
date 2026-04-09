package transcript

import (
	"os"
	"path/filepath"
	"testing"
)

func FuzzParse(f *testing.F) {
	f.Add([]byte(`{"type":"assistant","message":{"content":[],"usage":{"input_tokens":10,"output_tokens":5}}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`invalid jsonl line`))
	f.Add([]byte("{\"type\":\"assistant\"}\n{\"type\":\"progress\"}"))
	f.Fuzz(func(t *testing.T, data []byte) {
		tmp := filepath.Join(t.TempDir(), "fuzz.jsonl")
		if err := os.WriteFile(tmp, data, 0o644); err != nil {
			t.Fatal(err)
		}
		_, _ = Parse(tmp)
	})
}
