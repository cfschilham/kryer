package workers

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestPool(t *testing.T) {
	var (
		poolSize = 4
		taskQty  = 20
		n1       = 8
		n2       = 8
		want     = 64
		c        = make(chan int, taskQty)
		wg       = &sync.WaitGroup{}
	)

	task := NewTask(func(params []interface{}) {
		var (
			n1 = params[0].(int)
			n2 = params[1].(int)
			c  = params[2].(chan int)
			wg = params[3].(*sync.WaitGroup)
		)
		c <- n1 * n2
		wg.Done()
	})

	pool, err := NewPool(poolSize)
	if err != nil {
		t.Errorf("error creating pool: %s", err.Error())
		return
	}

	wg.Add(taskQty)
	task.SetParams([]interface{}{n1, n2, c, wg})
	for i := 0; i < taskQty; i++ {
		pool.QueueTask(*task)
	}

	RoutinesBeforeStart := runtime.NumGoroutine()
	pool.Start()

	wg.Wait()

	if len(c) != taskQty {
		t.Errorf("did not receive the right amount of results, want: %d got: %d", taskQty, len(c))
		return
	}

	for i := 0; i < taskQty; i++ {
		select {
		case got := <-c:
			if got != want {
				t.Errorf("unexpected result received, want: %d got %d", want, got)
			}
		default:
			t.Errorf("could not receive from channel c")
			return
		}
	}

	wg.Add(taskQty)
	for i := 0; i < taskQty; i++ {
		pool.QueueTask(*task)
	}

	wg.Wait()

	if len(c) != taskQty {
		t.Errorf("did not receive enough results, want: %d got: %d", taskQty, len(c))
		return
	}

	for i := 0; i < taskQty; i++ {
		select {
		case got := <-c:
			if got != want {
				t.Errorf("unexpected result received, want: %d got %d", want, got)
			}
		default:
			t.Errorf("could not receive from channel c")
			return
		}
	}

	pool.Dismiss()
	time.Sleep(time.Millisecond * 10) // give some time for goroutines to exit
	if runtime.NumGoroutine() > RoutinesBeforeStart {
		t.Errorf("failed to stop all goroutines of a dismissed pool, try giving more time")
	}
}

func TestPoolEarlyDismiss(t *testing.T) {
	poolSize := 4
	taskQty := 20

	task := NewTask(func(params []interface{}) {
		time.Sleep(time.Millisecond * 8)
	})

	pool, err := NewPool(poolSize)
	if err != nil {
		t.Errorf("error creating pool: %s", err.Error())
		return
	}

	for i := 0; i < taskQty; i++ {
		pool.QueueTask(*task)
	}

	RoutinesBeforeStart := runtime.NumGoroutine()
	pool.Start()

	pool.Dismiss()
	time.Sleep(time.Millisecond * 10) // give some time for goroutines to exit
	if runtime.NumGoroutine() > RoutinesBeforeStart {
		t.Errorf("failed to stop all goroutines of a dismissed pool, try giving more time")
	}
}

func BenchmarkPool(b *testing.B) {
	var (
		poolSize = runtime.NumCPU()
		taskQty  = runtime.NumCPU() * b.N * 2
		wg       = &sync.WaitGroup{}
	)

	task := NewTask(func(params []interface{}) {
		var (
			wg = params[0].(*sync.WaitGroup)
		)
		bcrypt.GenerateFromPassword([]byte("benchmark"), 12)
		wg.Done()
	})

	pool, err := NewPool(poolSize)
	if err != nil {
		b.Errorf("error creating pool: %s", err.Error())
		return
	}
	wg.Add(taskQty)
	for i := 0; i < taskQty; i++ {
		task.SetParams([]interface{}{wg})
		pool.QueueTask(*task)
	}

	b.StartTimer()
	pool.Start()
	wg.Wait()
	b.StopTimer()
	pool.Dismiss()
}
