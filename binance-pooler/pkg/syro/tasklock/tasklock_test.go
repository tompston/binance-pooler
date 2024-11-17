package tasklock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestExecute(t *testing.T) {
	tasklock := New()

	var executed int32 // Using int32 with atomic operations for thread-safety
	fn := func() {
		time.Sleep(50 * time.Millisecond) // Simulate some work
		atomic.AddInt32(&executed, 1)
	}

	// Run function in 5 goroutines simultaneously
	for i := 0; i < 5; i++ {
		go tasklock.Run("testCommand", fn)
	}

	// Sleep to allow all goroutines to potentially execute
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&executed) != 1 {
		t.Fatalf("Expected function to execute once, but executed %d times", executed)
	}
}

func TestMultipleCommands(t *testing.T) {
	commando := New()

	var cmd1Executed, cmd2Executed int32

	cmd1Fn := func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&cmd1Executed, 1)
	}

	cmd2Fn := func() {
		time.Sleep(50 * time.Millisecond)
		atomic.AddInt32(&cmd2Executed, 1)
	}

	// Run both functions in 5 goroutines simultaneously
	for i := 0; i < 5; i++ {
		go commando.Run("cmd1", cmd1Fn)
		go commando.Run("cmd2", cmd2Fn)
	}

	// Sleep to allow all goroutines to potentially execute
	time.Sleep(300 * time.Millisecond)

	if atomic.LoadInt32(&cmd1Executed) != 1 {
		t.Fatalf("Expected cmd1 function to execute once, but executed %d times", cmd1Executed)
	}

	if atomic.LoadInt32(&cmd2Executed) != 1 {
		t.Fatalf("Expected cmd2 function to execute once, but executed %d times", cmd2Executed)
	}
}

func TestRun_WithIdentifier(t *testing.T) {
	tl := New()

	var wg sync.WaitGroup
	var fn = func() {
		time.Sleep(100 * time.Millisecond) // Simulate some work
	}

	wg.Add(2)

	// Test running the same function with different identifiers
	go func() {
		defer wg.Done()
		if !tl.Run("testCommand", fn, "identifier1") {
			t.Errorf("Function with identifier1 was not allowed to run")
		}
	}()

	go func() {
		defer wg.Done()
		if !tl.Run("testCommand", fn, "identifier2") {
			t.Errorf("Function with identifier2 was not allowed to run")
		}
	}()

	wg.Wait()
}
