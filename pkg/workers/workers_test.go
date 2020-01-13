package workers

import (
	"sync"
	"testing"
)

func TestPool(t *testing.T) {
	poolSize := 20
	taskQty := 20

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
		t.Errorf("TestPool: error creating pool: %s", err.Error())
		return
	}

	c := make(chan int, taskQty)
	wg := &sync.WaitGroup{}
	wg.Add(taskQty)
	task.SetParams([]interface{}{8, 8, c, wg})
	for i := 0; i < taskQty; i++ {
		pool.QueueTask(*task)
	}

	if err := pool.Start(); err != nil {
		t.Errorf("TestPool: error starting pool: %s", err.Error())
		return
	}

	wg.Wait()

	if len(c) != taskQty {
		t.Errorf("TestPool: did not receive the right amount of results, want: %d got: %d", taskQty, len(c))
		return
	}

	for i := 0; i < taskQty; i++ {
		select {
		case n := <-c:
			if n != 64 {
				t.Errorf("TestPool: unexpected result received, want: %d got %d", 64, n)
			}
		default:
			t.Errorf("TestPool: could not receive from channel c")
			return
		}
	}

	wg.Add(taskQty)
	for i := 0; i < taskQty; i++ {
		pool.QueueTask(*task)
	}

	wg.Wait()

	if len(c) != taskQty {
		t.Errorf("TestPool: did not receive enough results, want: %d got: %d", taskQty, len(c))
		return
	}

	for i := 0; i < taskQty; i++ {
		select {
		case n := <-c:
			if n != 64 {
				t.Errorf("TestPool: unexpected result received, want: %d got %d", 64, n)
			}
		default:
			t.Errorf("TestPool: could not receive from channel c")
			return
		}
	}

	pool.Dismiss()
}
