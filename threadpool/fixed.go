package threadpool

import (
	"errors"
	"sync"
)

type FixedPool struct {
	capacity int
	tasksCh chan func()
	workers []worker
	wg *sync.WaitGroup
}

func NewFixed(capacity int) (Pool, error) {

	if capacity < 1 {
		return nil, errors.New("capacity should be a positive number different than zero")
	}

	tasksCh := make(chan func())

	var wg sync.WaitGroup

	var workers []worker
	for i:=0; i<capacity; i++ {
		workers = append(workers, worker{
			WorkerId(i),
			tasksCh,
			&wg,
			make(chan struct{}),
		})

		go workers[i].run()
	}

	return FixedPool{
		capacity,
		tasksCh,
		workers,
		&wg,
	}, nil
}

// Runs the 'task' function inside the thread pool.
// Blocks until a worker thread is available.
func (p FixedPool) Run(taskId string, task func()) {
	p.tasksCh <- task
}

func (p FixedPool) Stop() {

runPendingTasks:
	for {
		select {
			case task := <- p.tasksCh:
				task()
			default:
				break runPendingTasks
		}
	}

	close(p.tasksCh)

	p.wg.Wait()
}