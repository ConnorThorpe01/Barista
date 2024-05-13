package Beans

import (
	"Beans"
	"sync"
	"testing"
	"time"
)

func TestNewBeans(t *testing.T) {
	// Test NewBeans with default options
	pool, err := Beans.NewBeans()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if pool.MinBeans != 2 {
		t.Errorf("Expected MinBeans to be 2, got %d", pool.MinBeans)
	}

	if pool.MaxBeans != 100 {
		t.Errorf("Expected MaxBeans to be 100, got %d", pool.MaxBeans)
	}

	err = pool.Shutdown()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestSubmitAndShutdown(t *testing.T) {
	// Create a new Barista with custom options
	pool, err := Beans.NewBeans(
		Beans.WithMinBeans(2),
		Beans.WithMaxBeans(4),
		Beans.WithQueueCapacity(10),
	)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	numTasks := 10

	// Submit tasks to the pool
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		taskID := i
		err := pool.Submit(func() {
			defer wg.Done()
			t.Logf("Task %d is executed", taskID)
			time.Sleep(100 * time.Millisecond)  // Simulate task execution
			t.Logf("Task %d completed", taskID) // Log completion
		})

		if err != nil {
			return
		}
	}

	// Wait for all tasks to be executed
	wg.Wait()

	// Verify that all tasks have been processed by checking pool size
	if len(pool.Beans) > 4 {
		t.Errorf("Expected number of beans to be <= 4, got %d", len(pool.Beans))
	}

	// Shutdown the pool and ensure it stops all workers
	err = pool.Shutdown()
	if err != nil {
		t.Errorf("Expected no error got %s", err)
	}
}

func TestSubmitAndShutdown1000Tasks(t *testing.T) {
	// Create a new Barista with custom options
	pool, err := Beans.NewBeans(
		Beans.WithMinBeans(10),
		Beans.WithMaxBeans(500),
		Beans.WithQueueCapacity(100),
	)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Use a wait group to synchronize goroutines
	var wg sync.WaitGroup
	numTasks := 1000

	// Submit tasks to the pool
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		taskID := i
		err := pool.Submit(func() {
			defer wg.Done()
			t.Logf("Task %d is executed", taskID)
			time.Sleep(10 * time.Millisecond)   // Simulate task execution
			t.Logf("Task %d completed", taskID) // Log completion
		})

		if err != nil {
			return
		}
	}

	// Wait for all tasks to be executed
	wg.Wait()

	// Shutdown the pool and ensure it stops all workers
	err = pool.Shutdown()
	if err != nil {
		t.Errorf("Expected no error got %s", err)
	}
}
