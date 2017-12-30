package retry

import (
	"sync"
	"time"

	"github.com/wpajqz/linker"
)

type ReTry struct {
	items             sync.Map
	timeout, interval time.Duration
}

type Item struct {
	channel    string
	ctx        linker.Context
	value      interface{}
	retryTimes int
}

func NewRetry(interval, timeout time.Duration) *ReTry {
	return &ReTry{items: sync.Map{}, interval: interval, timeout: timeout}
}

func (rt *ReTry) Put(key interface{}, value *Item) {
	rt.items.Store(key, value)

	t1 := time.NewTimer(rt.interval)
	t2 := time.NewTimer(rt.timeout)
	for {
		select {
		case <-t1.C:
			if v, ok := rt.items.Load(key); ok {
				if i, ok := v.(*Item); ok {
					if i.retryTimes == 3 {
						rt.Delete(key)
						return
					}

					i.ctx.Write(i.channel, i.value)
					i.retryTimes++
				}
			} else {
				return
			}
		case <-t2.C:
			rt.Delete(key)
			return
		}
	}
}

func (rt *ReTry) Delete(key interface{}) {
	rt.items.Delete(key)
}
