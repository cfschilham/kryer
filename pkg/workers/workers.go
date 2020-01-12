package workers

import (
	"errors"
	"sync"
)

type Pool struct {
	size    int
	workers []*Worker
	queue   []*Task
	QueueWG sync.WaitGroup
	dismiss chan bool
}

type Task struct {
	fn     func(params []interface{})
	params []interface{}
}

type Worker struct {
	task    chan *Task
	TaskWG  sync.WaitGroup
	dismiss chan bool
}

// NewPool returns a new worker pool.
func NewPool(size int) (*Pool, error) {
	if size < 1 {
		return nil, errors.New("workers: pool size cannot be < 1")
	}

	return &Pool{
		size:    size,
		dismiss: make(chan bool),
		QueueWG: sync.WaitGroup{},
	}, nil
}

// NewTask returns a new task.
func NewTask(fn func(params []interface{})) *Task {
	return &Task{
		fn: fn,
	}
}

// NewWorker returns a new worker.
func NewWorker() *Worker {
	return &Worker{
		task:    make(chan *Task),
		TaskWG:  sync.WaitGroup{},
		dismiss: make(chan bool),
	}
}

func (p *Pool) QueueTask(t *Task) {
	p.QueueWG.Add(1)
	p.queue = append(p.queue, t)
}

func (p *Pool) Start() error {
	for i := 0; i < p.size; i++ {
		p.workers = append(p.workers, NewWorker())
	}

	for _, w := range p.workers {
		w.Start()
	}
	go func(p *Pool) {
		for {
			select {
			case <-p.dismiss:
				break
			default:
				p.assignTaskFromQueue()
			}
		}
	}(p)
	return nil
}

func (p *Pool) Dismiss() {
	for _, w := range p.workers {
		w.Dismiss()
	}
	p.dismiss <- true
}

// Wait waits for the pools queue to empty and then for all workers to finish. Do
// not queue tasks at the same time as this might cause unwanted behaviour.
func (p *Pool) Wait() {
	p.QueueWG.Wait()
	for _, w := range p.workers {
		w.TaskWG.Wait()
	}
}

func (p *Pool) assignTaskFromQueue() {
	if len(p.queue) <= 0 {
		return
	}

	t := p.queue[0]
	for _, w := range p.workers {
		if err := w.SetTask(t); err == nil {
			if len(p.queue) > 1 {
				p.queue = p.queue[1:]
			} else {
				p.queue = nil
			}
			p.QueueWG.Done()
			return
		}
	}
}

func (t *Task) SetParams(params []interface{}) {
	t.params = params
}

func (w *Worker) Dismiss() {
	w.dismiss <- true
}

func (w *Worker) SetTask(t *Task) error {
	select {
	case w.task <- t:
		w.TaskWG.Add(1)
		return nil
	default:
		return errors.New("workers: worker is busy")
	}
}

func (w *Worker) Start() {
	go func(w *Worker) {
		for {
			select {
			case <-w.dismiss:
				break
			case task := <-w.task:
				task.fn(task.params)
				w.TaskWG.Done()
			}
		}
	}(w)
}
