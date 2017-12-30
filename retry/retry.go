package retry

import (
	"sync"
	"time"

	"github.com/wpajqz/linker"
)

type ReTry struct {
	items             sync.Map
	timeout, interval time.Duration
	times             int
}

type Item struct {
	Channel    string
	Ctx        linker.Context
	Value      interface{}
	retryTimes int
}

func NewRetry(interval, timeout time.Duration, times int) *ReTry {
	return &ReTry{items: sync.Map{}, interval: interval, timeout: timeout, times: times}
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
					if i.retryTimes == rt.times {
						rt.Delete(key)
						return
					}

					i.Ctx.Write(i.Channel, i.Value)
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
