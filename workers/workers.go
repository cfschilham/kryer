package workers

import (
	"container/list"
	"errors"
	"runtime"
)

// Pool represents a pool of multiple managed workers
type Pool struct {
	size    int
	workers []*worker

	// Workers will transmit a pointer to themselves to this channel
	// when they are about to listen for incoming tasks on their task
	// channel.
	dormantPool chan *worker
	queue       *queue
	dismiss     chan bool

	// 0 = unstarted
	// 1 = running
	// 2 = dismissed
	state uint8
}

// Task represents the combination of a function and parameters to
// be executed by a worker.
type Task struct {
	fn     func(params []interface{})
	params []interface{}
}

type worker struct {
	task chan Task

	// Workers will transmit a pointer to themselves to this channel
	// when they are about to listen for incoming tasks on their task
	// channel.
	dormantPool chan *worker
	dismiss     chan bool
}

type queue struct {
	queue *list.List
	enqueue,
	dequeue chan Task
	dismiss chan bool
}

// NewPool returns a new worker pool.
func NewPool(size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("workers: pool size cannot be < 1")
	}

	return &Pool{
		size:        size,
		dormantPool: make(chan *worker, size),
		dismiss:     make(chan bool),
		queue:       newQueue(),
	}, nil
}

// NewTask returns a new task.
func NewTask(fn func(params []interface{})) *Task {
	return &Task{
		fn: fn,
	}
}

// newWorker returns a new worker.
func newWorker(dp chan *worker) *worker {
	return &worker{
		task:        make(chan Task),
		dormantPool: dp,
		dismiss:     make(chan bool, 1),
	}
}

// newQueue returns a new queue.
func newQueue() *queue {
	q := &queue{
		queue:   list.New(),
		enqueue: make(chan Task),
		dequeue: make(chan Task),
		dismiss: make(chan bool),
	}

	// A queue is started immediately on creation to be able to queue tasks
	// before the pool is started.
	q.start()
	return q
}

// QueueTask adds a task to the pools queue.
func (p *Pool) QueueTask(t Task) {
	if p.state == 2 {
		panic("workers: cannot queue tasks on a dismissed pool")
	}
	p.queue.enqueue <- t
}

// Start starts a pools goroutine as well as the goroutines of all its
// workers.
func (p *Pool) Start() {
	if p.state != 0 {
		panic("workers: cannot restart a running/dismissed pool")
	}
	p.state = 1

	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, newWorker(p.dormantPool))
	}
	for _, w := range p.workers {
		w.start()
	}
	p.start()
}

// Dismiss dismisses the pool and all of its workers.
func (p *Pool) Dismiss() {
	p.state = 2
	p.dismiss <- true
	p.queue.dismiss <- true
	for _, w := range p.workers {
		w.dismiss <- true
		// Buffer the task channel to ensure the pool's goroutine doesn't get
		// stuck on sending to a dismissed workers channel.
		w.task = make(chan Task, 1)
	}
}

// SetParams sets the parameters which will be passed to the tasks function.
func (t *Task) SetParams(params []interface{}) {
	t.params = params
}

// start starts the goroutine of the pool. It is stopped when the pool is dismissed.
func (p *Pool) start() {
	go func(p *Pool) {
		for {
			select {
			case <-p.dismiss:
				runtime.Goexit()
			case w := <-p.dormantPool: // wait for dormant worker
				select {
				case <-p.dismiss:
					runtime.Goexit()
				case t := <-p.queue.dequeue: // wait for queued task to assign
					w.task <- t // assign task to worker
				}
			}
		}
	}(p)
}

// start starts the goroutine of the worker. It is stopped when its pool is dismissed.
func (w *worker) start() {
	go func(w *worker) {
		for {
			w.dormantPool <- w
			select {
			case <-w.dismiss:
				runtime.Goexit()
			case t := <-w.task:
				t.fn(t.params)
			}
		}
	}(w)
}

// start starts the goroutine of the queue. It is stopped when its pool is dismissed.
func (q *queue) start() {
	go func(q *queue) {
		for {
			if q.queue.Front() == nil {
				select {
				case <-q.dismiss:
					runtime.Goexit()
				case t := <-q.enqueue:
					q.queue.PushBack(t)
				}
			}

			select {
			case <-q.dismiss:
				runtime.Goexit()
			case t := <-q.enqueue:
				q.queue.PushBack(t)
			case q.dequeue <- q.queue.Front().Value.(Task):
				q.queue.Remove(q.queue.Front())
			}
		}
	}(q)
}
