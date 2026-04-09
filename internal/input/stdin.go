package input

import (
	"fmt"
	"io"
	"os"

	gojson "github.com/goccy/go-json"
)

const maxStdinSize = 256 * 1024

// ReadFrom reads and parses stdin JSON from the given reader.
// Returns the parsed data and the raw bytes for debug logging.
func ReadFrom(r io.Reader) (*StdinData, []byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxStdinSize))
	if err != nil {
		return nil, nil, fmt.Errorf("reading stdin: %w", err)
	}
	if len(data) == 0 {
		return nil, nil, fmt.Errorf("empty stdin")
	}
	var stdin StdinData
	if err := gojson.Unmarshal(data, &stdin); err != nil {
		return nil, nil, fmt.Errorf("parsing stdin JSON: %w", err)
	}
	return &stdin, data, nil
}

// ReadStdin reads from os.Stdin when piped. Returns nil, nil, nil for TTY.
func ReadStdin() (*StdinData, []byte, error) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, nil, err
	}
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, nil, nil
	}
	return ReadFrom(os.Stdin)
}
