package threadpool

import (
    "errors"
    "sync"
)

// Pool where the capacity (number of workers) can be updated while running.
type VariPool struct {
    capacity int
    tasksCh chan func()
    workers []worker
    wg *sync.WaitGroup
    onTaskFinished func(taskId string) (newCap int, modifyCap bool)
}

func NewVari(initialCapacity int,
        onTaskFinished func(taskId string) (newCap int, modifyCap bool)) (Pool, error) {

    if initialCapacity < 1 {
        return nil, errors.New("capacity should be a positive number different than zero")
    }

    tasksCh := make(chan func())

    var wg sync.WaitGroup

    var workers []worker
    for i:=0; i<initialCapacity; i++ {
        workers = append(workers, worker{
            WorkerId(i),
            tasksCh,
            &wg,
            make(chan struct{}),
        })

        go workers[i].run()
    }

    return VariPool{
        initialCapacity,
        tasksCh,
        workers,
        &wg,
        onTaskFinished,
    }, nil
}

// Runs the 'task' function inside the thread pool.
// Blocks until a worker thread is available.
func (p VariPool) Run(taskId string, task func()) {
    p.tasksCh <- func() {
        task()
        if newCap, modifyCap := p.onTaskFinished(taskId); modifyCap {
            p.adjustCapacity(newCap)
        }
    }
}

func (p VariPool) Stop() {

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

func (p VariPool) adjustCapacity(newCap int) {
    if newCap == p.capacity {
        return
    } else if newCap < p.capacity {
        for i := newCap; i < p.capacity; i++ {
            p.workers[i].stopAcceptingTasksCh <- struct{}{}
        }

        p.workers = p.workers[:newCap]
    } else {
        for i:=0; i < newCap - p.capacity; i++ {
            p.workers = append(p.workers, worker{
                WorkerId(i),
                p.tasksCh,
                p.wg,
                make(chan struct{}),
            })

            go p.workers[len(p.workers) - 1].run()
        }
    }
}