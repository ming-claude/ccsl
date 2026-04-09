package input

import (
	"bytes"
	"testing"
)

func FuzzReadFrom(f *testing.F) {
	f.Add([]byte(`{"cwd":"/tmp"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`invalid json`))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _, _ = ReadFrom(bytes.NewReader(data))
	})
}
