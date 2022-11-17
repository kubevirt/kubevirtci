package cmd

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetCPUsRange(t *testing.T) {
	testCases := []struct {
		description    string
		cpu            uint
		numaCount      uint
		expectedRanges []string
	}{
		{
			description:    "two numa nodes, 6 cpus",
			cpu:            6,
			numaCount:      2,
			expectedRanges: []string{"0-2", "3-5"},
		},
		{
			description:    "three numa nodes, 6 cpus",
			cpu:            6,
			numaCount:      3,
			expectedRanges: []string{"0-1", "2-3", "4-5"},
		},
		{
			description:    "two numa nodes, 10 cpus",
			cpu:            10,
			numaCount:      2,
			expectedRanges: []string{"0-4", "5-9"},
		},
		{
			description:    "1 numa nodes, 10 cpus",
			cpu:            10,
			numaCount:      1,
			expectedRanges: []string{"0-9"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(*testing.T) {
			ranges := getCpusRanges(tc.cpu, tc.numaCount)
			if !reflect.DeepEqual(tc.expectedRanges, ranges) {
				t.Errorf("Expected ranges to equal: %v not: %v", tc.expectedRanges, ranges)
			}
		})
	}
}

func TestGetMemorySizeAndUnit(t *testing.T) {
	testCases := []struct {
		description      string
		memory           string
		expectedSize     uint64
		expectedUnit     string
		expectedErrorMsg string
	}{
		{
			description:      "simple case",
			memory:           "8096M",
			expectedSize:     8096,
			expectedUnit:     "M",
			expectedErrorMsg: "",
		},
		{
			description:      "wrong unit format",
			memory:           "8096Mi",
			expectedSize:     0,
			expectedUnit:     "",
			expectedErrorMsg: "expected memory unit to be",
		},
		{
			description:      "wrong unit",
			memory:           "8096T",
			expectedSize:     0,
			expectedUnit:     "",
			expectedErrorMsg: "expected memory unit to be",
		},
		{
			description:      "cannot convert size",
			memory:           "80a96T",
			expectedSize:     0,
			expectedUnit:     "",
			expectedErrorMsg: "cannot convert memory size",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(*testing.T) {
			size, unit, err := getMemorySizeAndUnit(tc.memory)
			if size != tc.expectedSize {
				t.Errorf("Expected size to equal: %d not: %d", tc.expectedSize, size)
			}
			if unit != tc.expectedUnit {
				t.Errorf("Expected unit to equal: %s not: %s", tc.expectedUnit, unit)
			}

			if tc.expectedErrorMsg == "" && err != nil {
				t.Errorf("Expected err to be nil not: %v", err)
			}

			if err != nil && !strings.Contains(err.Error(), tc.expectedErrorMsg) {
				t.Errorf("Expected err message: %s to contains %s", err.Error(), tc.expectedErrorMsg)
			}
		})
	}
}
