package sshconn

import (
	"sync"
)

// SendChanOnWGWait returns a channel which is passed `true` after `wg.Wait()` finishes.
func SendChanOnWGWait(wg *sync.WaitGroup) chan bool {
	c := make(chan bool, 1)
	go func(wg *sync.WaitGroup, c chan bool) {
		wg.Wait()
		c <- true
	}(wg, c)
	return c
}
