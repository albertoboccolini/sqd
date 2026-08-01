package files

import (
	"sync"

	"github.com/overthinkinglabs/sqd/models"
	"github.com/overthinkinglabs/sqd/services"
)

type Parallelizer struct {
	utils *services.Utils
}

func NewParallelizer(utils *services.Utils) *Parallelizer {
	return &Parallelizer{
		utils: utils,
	}
}

func (parallelizer *Parallelizer) ProcessFilesInParallel(
	files []string,
	processor func(string) (int, error),
	stats *models.ExecutionStats,
) (int, *models.ErrorCollection) {
	var (
		totalCount      int
		mutex           sync.Mutex
		waitingGroup    sync.WaitGroup
		errorCollection = models.NewErrorCollection()
		sem             = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for _, file := range files {
		waitingGroup.Add(1)
		sem <- struct{}{}

		go func(file string) {
			defer waitingGroup.Done()
			defer func() { <-sem }()

			count, err := processor(file)

			mutex.Lock()
			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
			} else {
				totalCount += count
				stats.Processed++
			}

			mutex.Unlock()
		}(file)
	}

	waitingGroup.Wait()
	return totalCount, errorCollection
}

func (parallelizer *Parallelizer) ProcessFilesInParallelNoCount(
	files []string,
	processor func(string) error,
	stats *models.ExecutionStats,
) *models.ErrorCollection {
	var (
		mutex           sync.Mutex
		waitingGroup    sync.WaitGroup
		errorCollection = models.NewErrorCollection()
		sem             = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for _, file := range files {
		waitingGroup.Add(1)
		sem <- struct{}{}

		go func(file string) {
			defer waitingGroup.Done()
			defer func() { <-sem }()

			err := processor(file)

			mutex.Lock()
			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
			} else {
				stats.Processed++
			}

			mutex.Unlock()
		}(file)
	}

	waitingGroup.Wait()
	return errorCollection
}

func (parallelizer *Parallelizer) ProcessFilesInParallelWithIndex(
	files []string,
	processor func(int, string) error,
	stats *models.ExecutionStats,
) *models.ErrorCollection {
	var (
		mutex           sync.Mutex
		waitingGroup    sync.WaitGroup
		errorCollection = models.NewErrorCollection()
		sem             = make(chan struct{}, models.MAX_CONCURRENT_GOROUTINES)
	)

	for index, file := range files {
		waitingGroup.Add(1)
		sem <- struct{}{}

		go func(index int, file string) {
			defer waitingGroup.Done()
			defer func() { <-sem }()

			err := processor(index, file)

			mutex.Lock()
			if err != nil {
				errorCollection.Add(err)
				stats.Skipped++
			} else {
				stats.Processed++
			}

			mutex.Unlock()
		}(index, file)
	}

	waitingGroup.Wait()
	return errorCollection
}
