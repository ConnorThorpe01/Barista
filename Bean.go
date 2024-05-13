package Beans

type Bean struct {
	pool *Barista      // Reference to the pool
	id   int           // Identifier for the bean (worker)
	quit chan struct{} // Channel to signal the worker to stop
}

func (b *Bean) brew() {
	for {
		select {
		case <-b.quit:
			return // Quit the worker goroutine
		case order, ok := <-b.pool.orders:
			if !ok {
				return
			}
			order() // Execute the task
		}
	}
}

func (b *Bean) stop() {
	// Signal the worker to stop by closing the quit channel
	close(b.quit)
}
