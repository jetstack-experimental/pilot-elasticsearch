package util

import "testing"

func TestNodeIndex(t *testing.T) {
	tests := []struct {
		Input  string
		Output int
		Err    error
	}{
		{
			Input:  "test-1",
			Output: 1,
			Err:    nil,
		},
		{
			Input:  "1",
			Output: 1,
			Err:    nil,
		},
		{
			Input:  "5",
			Output: 5,
			Err:    nil,
		},
		{
			Input:  "55a-5",
			Output: 5,
			Err:    nil,
		},
		{
			Input:  "55a-5",
			Output: 5,
			Err:    nil,
		},
		{
			Input:  "55a-g5",
			Output: 5,
			Err:    nil,
		},
		{
			Input:  "test-elasticsearch-data-1",
			Output: 1,
			Err:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Input, func(t *testing.T) {
			n, err := NodeIndex(test.Input)

			if err == nil {
				if test.Output != n {
					t.Errorf("expected %d but got %d", test.Output, n)
				} else {
					return
				}
			}

			if err != nil && test.Err == nil {
				t.Errorf("expected no error but got: %s", err.Error())
			} else {
				return
			}
		})
	}
}
