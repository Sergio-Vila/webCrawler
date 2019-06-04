package threadpool

type Pool interface {
	// Run a task inside the pool. There's no guarantee that
	// the task will be ran in a different thread than the one
	// calling Run.
	Run(taskId string, task func())

	// Stops the threads inside the pool. Calling 'Run' after 'Stop'
	// results on undefined behaviour.
	Stop()
}