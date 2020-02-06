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
	close       chan bool

	// 0 = unstarted
	// 1 = running
	// 2 = closed
	state uint8
}

// Task represents the combination of a function and parameters to
// be executed by a worker. You should use type casting in your Task function
// to use the parameters. Example of this:
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
	task chan Task

	// Workers will transmit a pointer to themselves to this channel
	// when they are about to listen for incoming tasks on their task
	// channel.
	dormantPool chan *worker
	close       chan bool
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

	return &Pool{
		size:        size,
		dormantPool: make(chan *worker, size),
		close:       make(chan bool),
		queue:       newQueue(),
	}, nil
}

// QueueTask adds a task to the pools queue.
func (p *Pool) QueueTask(t Task) error {
	if p.state == 2 {
		return errors.New("workers: cannot queue tasks on a closed pool")
	}
	p.queue.enqueue <- t
	return nil
}

// Start starts a pool, this includes a goroutine for all workers and two more for
// managing and assigning tasks to them.
func (p *Pool) Start() error {
	if p.state != 0 {
		return errors.New("workers: cannot restart a running/closed pool")
	}
	p.state = 1

	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, newWorker(p.dormantPool))
	}
	for _, w := range p.workers {
		w.start()
	}
	p.start()
	return nil
}

// newWorker returns a new worker.
func newWorker(dp chan *worker) *worker {
	return &worker{
		task:        make(chan Task),
		dormantPool: dp,
		close:       make(chan bool),
	}
}

// newQueue returns a new queue.
func newQueue() *queue {
	q := &queue{
		queued:  list.New(),
		enqueue: make(chan Task),
		dequeue: make(chan Task),
		close:   make(chan bool),
	}

	// A queue is started immediately on creation to be able to queue tasks
	// before the pool is started.
	q.start()
	return q
}

// start starts the goroutine of the pool. It is stopped when the pool is closed.
func (p *Pool) start() {
	go func() {
		for {
			w := <-p.dormantPool
			t := <-p.queue.dequeue
			select {
			case <-p.close:
				runtime.Goexit()
			case w.task <- t:
			}
		}
	}()
}

// start starts the goroutine of the worker. It is stopped when its pool is closed.
func (w *worker) start() {
	go func() {
		for {
			w.dormantPool <- w
			select {
			case <-w.close:
				runtime.Goexit()
			case t := <-w.task:
				t.Fn(t.Params)
			}
		}
	}()
}

// start starts the goroutine of the queue. It is stopped when its pool is closed.
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

// Close closes the pool and ends all associated goroutines.
func (p *Pool) Close() error {
	if p.state == 2 {
		return errors.New("workers: cannot close an already closed pool")
	}
	p.state = 2

	for _, w := range p.workers {
		w.Close()
	}

	// dormantPool chan must be closed to prevent the pools goroutine from getting
	// stuck on receiving.
	close(p.dormantPool)
	p.queue.Close()
	p.close <- true
	return nil
}

// Close closes a workers goroutine.
func (w *worker) Close() error {
	w.close <- true
	return nil
}

// Close closes a queue's goroutine.
func (q *queue) Close() error {
	q.close <- true

	// dequeue chan must be closed to prevent the pools goroutine from getting
	// stuck on receiving.
	close(q.dequeue)
	return nil
}
