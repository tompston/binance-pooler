// package tasklock holds a util function for executing functions and
// ensuring that only one instance of the function is running at a
// time. This is useful if we setup a way in which a caller can
// execute functions remotely, and ensure that if the same
// function is called multiple times, only one instance
// of it runs at a time.
package tasklock

import (
	"fmt"
	"sync"
)

type Tasklock struct {
	runningCommands sync.Map
}

func New() *Tasklock { return &Tasklock{} }

// Run returns false if the called function is already running.
// An optional identifier can be provided to distinguish
// between different invocations of the same function.
func (c *Tasklock) Run(command string, fn func(), identifier ...string) bool {
	// Construct a unique key for the function invocation
	key := command
	if len(identifier) > 0 {
		key = fmt.Sprintf("%s-%s", command, identifier[0])
	}

	if _, funcIsAlreadyRunning := c.runningCommands.LoadOrStore(key, true); funcIsAlreadyRunning {
		return false
	}

	// Execute the function in a goroutine, so it doesn't block
	go func() {
		fn()
		c.runningCommands.Delete(key) // Remove the command from the running list after completion
	}()

	return true
}
