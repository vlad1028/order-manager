package cli

import (
	"sync"
)

type Task func()

type WorkerPool struct {
	tasks      chan Task
	cancel     chan struct{}
	wg         sync.WaitGroup
	mu         sync.Mutex
	numWorkers uint
}

func NewWorkerPool(numWorkers uint, bufSize uint) *WorkerPool {
	pool := &WorkerPool{
		tasks:      make(chan Task, bufSize),
		numWorkers: numWorkers,
	}

	pool.addWorkers(numWorkers)

	return pool
}

func (p *WorkerPool) addWorkers(n uint) {
	for i := uint(0); i < n; i++ {
		go p.worker()
	}
}

func (p *WorkerPool) removeWorkers(n uint) {
	for i := uint(0); i < n; i++ {
		p.cancel <- struct{}{}
	}
}

//gocognit:ignore
func (p *WorkerPool) worker() {
	for {
		select {
		case <-p.cancel:
			return
		case task, ok := <-p.tasks:
			if !ok {
				return
			}
			task()
			p.wg.Done()
		}
	}
}

func (p *WorkerPool) AddTask(task Task) {
	p.wg.Add(1)
	p.tasks <- task
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool) Close() {
	close(p.tasks)
	p.Wait()
}

func (p *WorkerPool) SetNumWorkers(numWorkers uint) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if numWorkers > p.numWorkers {
		p.addWorkers(numWorkers - p.numWorkers)
	} else if numWorkers < p.numWorkers {
		p.removeWorkers(p.numWorkers - numWorkers)
	}

	p.numWorkers = numWorkers
}
