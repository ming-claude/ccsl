//go:build darwin

package cmd

import (
	"testing"
)

func TestParseVMStat(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		expect map[string]int64
	}{
		{
			name: "normal vm_stat output",
			input: `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                              123456.
Pages active:                            234567.
Pages inactive:                          345678.
Pages speculative:                       45678.
Pages wired down:                        56789.
"Translation faults":                    9876543.`,
			expect: map[string]int64{
				"Pages free":             123456,
				"Pages active":           234567,
				"Pages inactive":         345678,
				"Pages speculative":      45678,
				"Pages wired down":       56789,
				"\"Translation faults\"": 9876543,
			},
		},
		{
			name:   "empty string",
			input:  "",
			expect: map[string]int64{},
		},
		{
			name:   "malformed output no colons",
			input:  "this has no colons at all\nneither does this line",
			expect: map[string]int64{},
		},
		{
			name: "lines without trailing dots",
			input: `Pages free:                              100
Pages active:                            200`,
			expect: map[string]int64{
				"Pages free":   100,
				"Pages active": 200,
			},
		},
		{
			name:   "header line skipped",
			input:  "Mach Virtual Memory Statistics: (page size of 16384 bytes)",
			expect: map[string]int64{},
		},
		{
			name: "header plus data lines",
			input: `Mach Virtual Memory Statistics: (page size of 16384 bytes)
Pages free:                              42.`,
			expect: map[string]int64{
				"Pages free": 42,
			},
		},
		{
			name:   "colon present but value not numeric",
			input:  "Some field: not_a_number.",
			expect: map[string]int64{},
		},
		{
			name: "mixed valid and invalid lines",
			input: `Header line without colon
Pages free:                              10.
Bad line
Pages active:                            abc.
Pages wired down:                        20.`,
			expect: map[string]int64{
				"Pages free":       10,
				"Pages wired down": 20,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := parseVMStat(tc.input)

			if len(got) != len(tc.expect) {
				t.Fatalf("len(got) = %d, want %d\ngot: %v", len(got), len(tc.expect), got)
			}

			for k, wantVal := range tc.expect {
				gotVal, ok := got[k]
				if !ok {
					t.Errorf("missing key %q", k)
					continue
				}
				if gotVal != wantVal {
					t.Errorf("got[%q] = %d, want %d", k, gotVal, wantVal)
				}
			}
		})
	}
}
