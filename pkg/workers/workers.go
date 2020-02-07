package workers

import (
	"container/list"
	"errors"
	"runtime"
)

// Pool represents a pool of multiple managed workers.
type Pool struct {
	size    int
	workers []*worker
	queue   *queue

	// 0 = unstarted
	// 1 = running
	// 2 = closed
	state uint8
}

// Task represents the combination of a function and parameters to be executed
// by a worker. You should use type casting in your Task function to use the
// parameters. Example of this:
//
//	Task{
//		Fn: func(params []interface{}) {
//			var (
//				n1 = params[0].(int)
//				n2 = params[1].(int)
//			)
//
//			fmt.Printf("Sum: %d", n1+n2)
//		},
//		Params: []interface{}{5, 5}
//	}
//
type Task struct {
	Fn     func(params []interface{})
	Params []interface{}
}

type worker struct {
	task  chan Task
	close chan bool
}

type queue struct {
	queued *list.List
	enqueue,
	dequeue chan Task
	close chan bool
}

// NewPool returns a new worker pool.
func NewPool(size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("workers: pool size cannot be < 1")
	}
	p := &Pool{
		size: size,
		queue: &queue{
			queued:  list.New(),
			enqueue: make(chan Task),
			dequeue: make(chan Task),
			close:   make(chan bool),
		},
	}
	p.queue.start()
	return p, nil
}

// Queue queues a task on a pools queue.
func (p *Pool) Queue(t Task) error {
	if p.state == 2 {
		return errors.New("workers: cannot queue tasks on a closed pool")
	}
	p.queue.enqueue <- t
	return nil
}

// Start starts a pool. This includes one goroutine for each worker and one for
// the queue.
func (p *Pool) Start() error {
	if p.state != 0 {
		return errors.New("workers: cannot restart a running/closed pool")
	}
	p.state = 1

	for i := 0; i < p.size; i++ {
		w := &worker{
			task:  p.queue.dequeue,
			close: make(chan bool),
		}
		p.workers = append(p.workers, w)
		w.start()
	}
	return nil
}

// start starts the goroutine of the worker. It is stopped when it receives on
// its close channel.
func (w *worker) start() {
	go func() {
		for {
			select {
			case <-w.close:
				runtime.Goexit()
			case t := <-w.task:
				t.Fn(t.Params)
			}
		}
	}()
}

// start starts the goroutine of the queue. It is stopped when it receives on
// its close channel.
func (q *queue) start() {
	go func() {
		for {
			if q.queued.Front() == nil {
				select {
				case <-q.close:
					runtime.Goexit()
				case t := <-q.enqueue:
					q.queued.PushBack(t)
				}
			}

			select {
			case <-q.close:
				runtime.Goexit()
			case t := <-q.enqueue:
				q.queued.PushBack(t)
			case q.dequeue <- q.queued.Front().Value.(Task):
				q.queued.Remove(q.queued.Front())
			}
		}
	}()
}

// Close closes the pool and stops all associated goroutines.
func (p *Pool) Close() error {
	if p.state == 2 {
		return errors.New("workers: cannot close an already closed pool")
	}
	p.state = 2

	p.queue.Close()
	for _, w := range p.workers {
		w.Close()
	}
	return nil
}

// Close stops a workers goroutine.
func (w *worker) Close() error {
	w.close <- true
	return nil
}

// Close stops a queue's goroutine.
func (q *queue) Close() error {
	q.close <- true
	return nil
}
