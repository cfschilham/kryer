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
	dismissChan chan bool
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
	dismissChan chan bool
}

type queue struct {
	queue *list.List
	enqueue,
	dequeue chan Task
	dismissChan chan bool
}

// NewPool returns a new worker pool.
func NewPool(size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("workers: pool size cannot be < 1")
	}

	return &Pool{
		size:        size,
		dormantPool: make(chan *worker, size),
		dismissChan: make(chan bool, 1),
		queue:       newQueue(),
	}, nil
}

// NewTask returns a new task.
func NewTask(fn func(params []interface{})) *Task {
	return &Task{
		fn: fn,
	}
}

// NewWorker returns a new worker.
func newWorker(dp chan *worker) *worker {
	return &worker{
		task:        make(chan Task),
		dormantPool: dp,
		dismissChan: make(chan bool, 1),
	}
}

func newQueue() *queue {
	q := &queue{
		queue:       list.New(),
		enqueue:     make(chan Task),
		dequeue:     make(chan Task),
		dismissChan: make(chan bool, 1),
	}

	go func(q *queue) {
		for {
			if q.queue.Front() == nil {
				select {
				case <-q.dismissChan:
					runtime.Goexit()
				case t := <-q.enqueue:
					q.queue.PushBack(t)
				}
				continue
			}

			select {
			case <-q.dismissChan:
				break
			case t := <-q.enqueue:
				q.queue.PushBack(t)
			case q.dequeue <- q.queue.Front().Value.(Task):
				q.queue.Remove(q.queue.Front())
			}
		}
	}(q)

	return q
}

// QueueTask adds a task to the pools queue.
func (p *Pool) QueueTask(t Task) {
	p.queue.enqueue <- t
}

// Start starts a pools goroutine as well as the goroutines of all its
// workers.
func (p *Pool) Start() error {
	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, newWorker(p.dormantPool))
	}

	for _, w := range p.workers {
		w.Start()
	}
	go func(p *Pool) {
		for {
			select {
			case <-p.dismissChan:
				runtime.Goexit()
			case t := <-p.queue.dequeue:
				select {
				case w := <-p.dormantPool:
					w.setTask(t)
				default:
					// If there currently isn't any dormant worker, requeue the task.
					p.QueueTask(t)
				}
			}
		}
	}(p)
	return nil
}

// Dismiss dismisses the pool and all of its workers.
func (p *Pool) Dismiss() {
	p.dismissChan <- true
	p.queue.dismiss()
	for _, w := range p.workers {
		w.dismiss()
		// Buffer the task channel to ensure the pool's main goroutine doesn't
		// get stuck on sending to a dismissed workers channel.
		w.task = make(chan Task, 1)
	}
}

// SetParams sets the parameters which will be passed to the tasks function.
func (t *Task) SetParams(params []interface{}) {
	t.params = params
}

// dismiss dismisses a worker, ending its goroutine.
func (w *worker) dismiss() {
	w.dismissChan <- true
}

// setTask gives a worker a new task.
func (w *worker) setTask(t Task) {
	w.task <- t
}

// Start starts the goroutine of the worker. To stop the goroutine, call the
// workers dismiss method.
func (w *worker) Start() {
	go func(w *worker) {
		for {
			w.dormantPool <- w
			select {
			case <-w.dismissChan:
				runtime.Goexit()
			case t := <-w.task:
				if t.fn == nil {
					panic("workers: received nil function")
				}
				t.fn(t.params)
			}
		}
	}(w)
}

// dismiss dismisses a queue's goroutine.
func (q *queue) dismiss() {
	q.dismissChan <- true
}
