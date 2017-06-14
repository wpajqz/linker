package session

import (
	"sync"

	"github.com/wpajqz/linker"
)

type (
	// socket信息、在线状态
	Session struct {
		Ctx    *linker.Context
		Status int // 0:不在线;1:在线
	}

	// session的宿主平台mobile、pad、pc等。
	Platform struct {
		lock sync.RWMutex
		data map[string]Session
	}
)

func (p Platform) Get(key string) (Session, bool) {
	p.lock.RLock()
	v, ok := p.data[key]
	p.lock.RUnlock()
	return v, ok
}

func (p Platform) Set(key string, value Session) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.data[key] = value
}

func (p Platform) Delete(key string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	delete(p.data, key)
}

func (p Platform) IsExist(key string) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, ok := p.data[key]
	return ok
}

func (p Platform) AllPlatform() map[string]Session {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.data
}
