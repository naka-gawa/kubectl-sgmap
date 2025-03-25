package utils

import (
	"context"
	"maps"
	"sync"
)

// ProcessBatchFunc defines the function signature for processing a batch
// T is input item type, R is output result type (map key: string)
type ProcessBatchFunc[T any, R any] func(context.Context, []T) (map[string]R, error)

// batchResult represents the result of a single batch
type batchResult[R any] struct {
	data map[string]R
	err  error
}

// RunBatchParallel splits items into batches and processes them concurrently.
//
// Arguments:
//
//	ctx         - Context for cancellation
//	items       - Input items
//	batchSize   - Size of each batch
//	maxParallel - Maximum number of parallel workers
//	process     - Processing function
//
// Returns:
//
//	A map with string keys representing merged batch results.
//	Returns an error if any batch fails.
func RunBatchParallel[T any, R any](
	ctx context.Context,
	items []T,
	batchSize int,
	maxParallel int,
	process ProcessBatchFunc[T, R],
) (map[string]R, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	batches := chunkSlice(items, batchSize)
	resultCh := make(chan batchResult[R], len(batches))

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxParallel)

	for _, batch := range batches {
		wg.Add(1)
		sem <- struct{}{}

		go func(b []T) {
			defer func() {
				wg.Done()
				<-sem
			}()
			runBatch(ctx, b, process, resultCh, cancel)
		}(batch)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	return collectResults(resultCh)
}

// runBatch processes a single batch and sends the result to the channel
//
// Arguments:
//
//	ctx: Context for cancellation
//	batch: Batch of items
//	process: Processing function
//	resultCh: Channel for results
//	cancel: Cancellation function
func runBatch[T any, R any](
	ctx context.Context,
	batch []T,
	process ProcessBatchFunc[T, R],
	resultCh chan<- batchResult[R],
	cancel context.CancelFunc,
) {
	res, err := process(ctx, batch)
	if err != nil {
		cancel()
	}
	resultCh <- batchResult[R]{data: res, err: err}
}

// collectResults collects results from the channel and merges them into a single map
func collectResults[R any](results <-chan batchResult[R]) (map[string]R, error) {
	final := make(map[string]R)
	var firstErr error

	for res := range results {
		if res.err != nil && firstErr == nil {
			firstErr = res.err
			continue
		}
		maps.Copy(final, res.data)
	}

	if firstErr != nil {
		return nil, firstErr
	}
	return final, nil
}

// chunkSlice splits a slice into chunks of the given size
func chunkSlice[T any](items []T, size int) [][]T {
	var chunks [][]T
	for i := 0; i < len(items); i += size {
		end := min(i+size, len(items))
		chunks = append(chunks, items[i:end])
	}
	return chunks
}
