package main

import "testing"

func TestLimitCollection(t *testing.T) {
	tcs := []struct {
		length, limit, offset, min, max int
	}{
		{length: 10, limit: 10, offset: 0, min: 0, max: 10},
		{length: 10, limit: 5, offset: 0, min: 0, max: 5},
		{length: 10, limit: 5, offset: 5, min: 5, max: 10},
		{length: 10, limit: 10, offset: 10, min: 0, max: 0},
		{length: 10, limit: 50, offset: 0, min: 0, max: 10},
		{length: 0, limit: 10, offset: 5, min: 0, max: 0},
	}

	for _, tc := range tcs {
		min, max := limitCollection(tc.length, tc.limit, tc.offset)

		if tc.min != min {
			t.Errorf("Expected min to eq %d, got %d", tc.min, min)
		}

		if tc.max != max {
			t.Errorf("Expected max to eq %d, got %d", tc.max, max)
		}
	}
}
