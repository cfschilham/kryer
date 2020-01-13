package workers

import (
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	poolSize := 10
	taskQty := 20

	task := NewTask(func(params []interface{}) {
		var (
			n1 = params[0].(int)
			n2 = params[1].(int)
			c  = params[2].(chan int)
		)
		c <- n1 * n2
	})

	pool, err := NewPool(poolSize)
	if err != nil {
		t.Fatalf("TestAll: error creating pool: %s", err.Error())
	}

	c := make(chan int, taskQty)
	for i := 0; i < taskQty; i++ {
		task.SetParams([]interface{}{8, 8, c})
		pool.QueueTask(*task)
	}

	if err := pool.Start(); err != nil {
		t.Fatalf("TestAll: error starting pool: %s", err.Error())
	}

	pool.Wait()

	for i := 0; i < taskQty; i++ {
		if <-c != 64 {
			t.Fatalf("TestAll: unexpected result received")
		}
	}

	for i := 0; i < taskQty; i++ {
		task.SetParams([]interface{}{8, 8, c})
		pool.QueueTask(*task)
	}
	time.Sleep(time.Millisecond * 50)
	select {
	case <-c:
		t.Fatalf("TestAll: able to continue after dimissal")
	default:
		return
	}
}
