package utils

import (
	"context"
	"sync"
)

// ProcessBatchFunc defines the function signature for processing a batch
// T is input item type, R is output result type (map key: string)
type ProcessBatchFunc[T any, R any] func(context.Context, []T) (map[string]R, error)

// RunBatchParallel splits the input items into batches and runs the processing function in parallel.
// It returns a merged map of all batch results or the first error encountered.
func RunBatchParallel[T any, R any](
	ctx context.Context,
	items []T,
	batchSize int,
	maxParallel int,
	process ProcessBatchFunc[T, R],
) (map[string]R, error) {
	batches := chunkSlice(items, batchSize)
	type result struct {
		data map[string]R
		err  error
	}

	resultCh := make(chan result, len(batches))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}
		go func(b []T) {
			defer func() {
				wg.Done()
				<-sem
			}()

			out, err := process(ctx, b)
			resultCh <- result{data: out, err: err}
			if err != nil {
				cancel()
			}
		}(batch)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	final := make(map[string]R)
	var firstErr error
	for res := range resultCh {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
			continue
		}
		for k, v := range res.data {
			final[k] = v
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return final, nil
}

func chunkSlice[T any](items []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
