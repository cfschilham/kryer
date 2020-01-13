package workers

import (
	"errors"
	"sync"
)

// Pool represents a pool of multiple managed workers
type Pool struct {
	size    int
	workers []*worker

	// Workers will transmit a pointer to themselves to this channel
	// when they are about to listen for incoming tasks on their task
	// channel.
	dormantPool chan *worker
	queue       []Task

	// Finished when the queue reached 0 tasks.
	QueueWG sync.WaitGroup
	dismiss chan bool
}

// Task represents the combination of a function and parameters to
// be executed by a worker.
type Task struct {
	fn     func(params []interface{})
	params []interface{}
}

type worker struct {
	task        chan Task
	taskWG      sync.WaitGroup
	dormantPool chan *worker
	dismiss     chan bool
}

// NewPool returns a new worker pool.
func NewPool(size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("workers: pool size cannot be < 1")
	}

	return &Pool{
		size:        size,
		dormantPool: make(chan *worker, size),
		dismiss:     make(chan bool, 1),
		QueueWG:     sync.WaitGroup{},
	}, nil
}

// NewTask returns a new task.
func NewTask(fn func(params []interface{})) *Task {
	return &Task{
		fn: fn,
	}
}

// NewWorker returns a new worker.
func newWorker() *worker {
	return &worker{
		task:    make(chan Task),
		taskWG:  sync.WaitGroup{},
		dismiss: make(chan bool, 1),
	}
}

// QueueTask adds a task to the pools queue.
func (p *Pool) QueueTask(t Task) {
	p.QueueWG.Add(1)
	p.queue = append(p.queue, t)
}

// Start starts a pools goroutine as well as the goroutines of all its
// workers.
func (p *Pool) Start() error {
	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, newWorker())
	}

	for _, w := range p.workers {
		w.dormantPool = p.dormantPool
		w.Start()
	}
	go func(p *Pool) {
		for {
			select {
			case <-p.dismiss:
				break
			case w := <-p.dormantPool:
				if len(p.queue) > 0 {
					w.setTask(p.queue[0])

					if len(p.queue) > 1 {
						p.queue = p.queue[1:]
					} else {
						p.queue = nil
					}
					p.QueueWG.Done()
				}
			}
		}
	}(p)
	return nil
}

// Dismiss dismisses the pool and all of its workers.
func (p *Pool) Dismiss() {
	p.dismiss <- true
	for _, w := range p.workers {
		w.Dismiss()
	}
}

// Wait waits for the pools queue to empty and then for all workers to finish. Do
// not queue tasks at the same time as this might cause unwanted behaviour.
func (p *Pool) Wait() {
	p.QueueWG.Wait()
	for _, w := range p.workers {
		w.taskWG.Wait()
	}
}

// SetParams sets the parameters which will be passed to the tasks function.
func (t *Task) SetParams(params []interface{}) {
	t.params = params
}

// Dismiss dismisses a worker, ending its goroutine.
func (w *worker) Dismiss() {
	w.dismiss <- true

	// Safeguard to prevent the pool from sending to a dismissed worker's task channel.
	w.task = nil
}

// SetTask attempts to give a worker a new task. If the worker is busy an error
// will be returned, otherwise nil will be returned.
func (w *worker) setTask(t Task) {
	w.taskWG.Add(1)
	w.task <- t
}

// Start starts the goroutine of the worker. To stop the goroutine, call the
// workers dismiss method.
func (w *worker) Start() {
	go func(w *worker) {
		for {
			w.dormantPool <- w
			select {
			case <-w.dismiss:
				break
			case task := <-w.task:
				task.fn(task.params)
				w.taskWG.Done()
			}
		}
	}(w)
}
