package Beans

import (
	"errors"
	"sync"
	"time"
)

type Barista struct {
	Beans     []*Bean
	MinBeans  int32
	MaxBeans  int32
	orders    chan func()
	monitor   chan func()
	IsRunning bool
}

type BaristaSettings struct {
	MinBeans      int32
	MaxBeans      int32
	QueueCapacity uint32
}

type BaristaOptions func(*BaristaSettings)

var ErrPoolShutdown = errors.New("pool has been shut down")
var ErrQueueFull = errors.New("task queue is full")

// WithMinBeans sets the minimum number of beans (workers) in the pool.
func WithMinBeans(minBeans int32) BaristaOptions {
	return func(options *BaristaSettings) {
		options.MinBeans = minBeans
	}
}

// WithMaxBeans sets the maximum number of beans (workers) in the pool.
func WithMaxBeans(maxBeans int32) BaristaOptions {
	return func(options *BaristaSettings) {
		options.MaxBeans = maxBeans
	}
}

// WithQueueCapacity sets the capacity of the task queue.
func WithQueueCapacity(capacity uint32) BaristaOptions {
	return func(options *BaristaSettings) {
		options.QueueCapacity = capacity
	}
}

func NewBeans(options ...BaristaOptions) (*Barista, error) {
	// Default options
	defaultOptions := BaristaSettings{
		MinBeans:      2,
		MaxBeans:      100,
		QueueCapacity: 500,
	}

	// Apply custom options
	for _, option := range options {
		option(&defaultOptions)
	}

	// Validate MinBeans and MaxBeans
	if defaultOptions.MaxBeans < defaultOptions.MinBeans {
		return nil, errors.New("maxBeans cannot be less than minBeans")
	}

	pool := &Barista{
		orders:    make(chan func(), defaultOptions.QueueCapacity),
		monitor:   make(chan func(), defaultOptions.QueueCapacity),
		MinBeans:  defaultOptions.MinBeans,
		MaxBeans:  defaultOptions.MaxBeans,
		IsRunning: true,
	}

	for i := 0; i < int(defaultOptions.MinBeans); i++ {
		b := &Bean{
			pool: pool,
			id:   i,
			quit: make(chan struct{}),
		}
		pool.Beans = append(pool.Beans, b)
		go b.brew()
	}

	go pool.monitorTasks()

	return pool, nil
}

func (p *Barista) monitorTasks() {
	ticker := time.NewTicker(10 * time.Millisecond) // Adjust interval as needed
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Periodically check workload to kill or spawn beans
			if p.shouldKillBean() {
				p.killBean()
			}
		case _, ok := <-p.monitor:
			if !ok {
				// p.monitor channel is closed, terminate monitoring
				return
			}
			// Handle new task received on p.monitor if needed
			// This could trigger a check to spawn a new bean if workload conditions require
			if p.shouldSpawnNewBean() {
				p.spawnNewBean()
			}
		}
	}
}

func (p *Barista) shouldSpawnNewBean() bool {
	// Adjust these thresholds based on workload and desired behavior
	return len(p.orders) > len(p.Beans) && len(p.Beans) < int(p.MaxBeans)
}

func (p *Barista) shouldKillBean() bool {
	return len(p.orders) < len(p.Beans) && len(p.Beans) > int(p.MinBeans)
}

func (p *Barista) spawnNewBean() {
	if len(p.Beans) < int(p.MaxBeans) {
		b := &Bean{
			pool: p,
			id:   len(p.Beans),
			quit: make(chan struct{}),
		}
		p.Beans = append(p.Beans, b)
		go b.brew()
	}
}

func (p *Barista) killBean() {
	if len(p.Beans) > int(p.MinBeans) {
		// Signal the last Bean to stop
		b := p.Beans[len(p.Beans)-1]
		b.stop()
		p.Beans = p.Beans[:len(p.Beans)-1]
	}
}

func (p *Barista) Submit(task func()) error {
	if !p.IsRunning {
		return ErrPoolShutdown
	}

	select {
	case p.orders <- task:
		return nil
	default:
		return ErrQueueFull
	}
}

func (p *Barista) Shutdown() error {
	if p.IsRunning {
		var wg sync.WaitGroup

		for _, b := range p.Beans {
			wg.Add(1)
			go func(b *Bean) {
				defer wg.Done()
				b.stop()
			}(b)
		}

		wg.Wait()

		close(p.orders)
		p.IsRunning = false
		return nil
	}
	return ErrPoolShutdown
}
