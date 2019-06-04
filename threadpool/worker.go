package threadpool

import (
	"sync"
)

type WorkerId int

type worker struct {
	id WorkerId
	tasksCh chan func ()
	wg *sync.WaitGroup
	stopAcceptingTasksCh chan struct{}
}

func (w *worker) run() {

	w.wg.Add(1)

workerLoop:
	for {
		select {
			case task, isOpen := <-w.tasksCh:
				if !isOpen {
					break workerLoop
				}

				task()

			case <- w.stopAcceptingTasksCh:
				break workerLoop
		}
	}

	w.wg.Done()
}