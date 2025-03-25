package utils

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
)

func TestRunBatchParallel(t *testing.T) {
	type testCase struct {
		name        string
		items       []string
		batchSize   int
		maxParallel int
		process     ProcessBatchFunc[string, string]
		wantErr     bool
		expectKeys  []string
	}

	tests := []testCase{
		{
			name:        "all batches succeed",
			items:       []string{"192.168.0.1", "192.168.0.2"},
			batchSize:   2,
			maxParallel: 2,
			process:     makeProcessWithErrorIP(""),
			wantErr:     false,
			expectKeys:  []string{"192.168.0.1", "192.168.0.2"},
		},
		{
			name:        "first batch fails",
			items:       []string{"192.168.0.1", "192.168.0.2"},
			batchSize:   2,
			maxParallel: 2,
			process:     makeProcessWithErrorIP("192.168.0.1"),
			wantErr:     true,
			expectKeys:  nil,
		},
		{
			name:        "context canceled before second batch",
			items:       []string{"192.168.0.1", "192.168.0.2"},
			batchSize:   2,
			maxParallel: 1,
			process:     makeProcessWithErrorIP("192.168.0.1"),
			wantErr:     true,
			expectKeys:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := RunBatchParallel(ctx, tt.items, tt.batchSize, tt.maxParallel, tt.process)

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected error: %v, got: %v", tt.wantErr, err)
			}

			if !tt.wantErr {
				var gotKeys []string
				for k := range got {
					gotKeys = append(gotKeys, k)
				}
				sort.Strings(gotKeys)
				sort.Strings(tt.expectKeys)
				if !reflect.DeepEqual(gotKeys, tt.expectKeys) {
					t.Errorf("expected keys %v, but got %v (case: %s)", tt.expectKeys, gotKeys, tt.name)
				}
			}
		})
	}
}

func makeProcessWithErrorIP(errorIp string) ProcessBatchFunc[string, string] {
	return func(ctx context.Context, batch []string) (map[string]string, error) {
		for _, ip := range batch {
			if ip == errorIp {
				return nil, errors.New("error from first batch")
			}
		}
		res := make(map[string]string)
		for _, v := range batch {
			res[v] = "ok"
		}
		return res, nil
	}
}
