package threadpool

import (
    "github.com/stretchr/testify/assert"
    "testing"
)

func TestShouldFailToCreatePoolWithNegOrZeroCapacity(t *testing.T) {
    assert := assert.New(t)

    pool, error := NewFixed(0)
    assert.NotNil(error, "Expected error to not be nil")
    assert.Nil(pool, "Expected pool to be nil")

    pool, error = NewFixed(-1)
    assert.NotNil(error, "Expected error to not be nil")
    assert.Nil(pool, "Expected pool to be nil")
}

func TestShouldStartAndStopPool(t *testing.T) {
    assert := assert.New(t)

    pool, error := NewFixed(1)
    assert.NotNil(pool, "Expected pool to not be nil")
    assert.Nil(error, "Expected error to be nil")

    pool.Stop()
}

func TestShouldRunAScheduledTask(t *testing.T) {
    assert := assert.New(t)

    pool, _ := NewFixed(1)

    taskHasRan := false

    pool.Run("taskId", func() {
        taskHasRan = true
    })

    pool.Stop()

    assert.True(taskHasRan, "Expected task to ran")
}

func TestShouldRunMoreTasksThanItsCapacity(t *testing.T) {
    assert := assert.New(t)

    const capacity = 5
    const tasksToRun = 10

    pool, _ := NewFixed(capacity)

    tasksRan := 0

    taskRanCh := make(chan bool)

    go func() {
        for {
            _, chanIsOpen := <- taskRanCh
            if !chanIsOpen {
                break
            }

            tasksRan++
        }
    } ()

    for i:=0; i<tasksToRun; i++ {
        pool.Run("taskId", func() {
            taskRanCh <- true
        })
    }

    pool.Stop()
    close(taskRanCh)

    assert.Equal(tasksToRun, tasksRan, "Expected to ran %d tasks")
}