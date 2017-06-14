package session

import (
	"sync"
)

const (
	OFFLINE = 0
	ONLINE  = 1
)

var (
	defaultSession = make(map[string]Platform)
	mutex          sync.RWMutex
)

func Get(key string) Platform {
	mutex.RLock()
	defer mutex.RUnlock()

	return defaultSession[key]
}

func Set(key, platform string, session Session) {
	mutex.Lock()
	defer mutex.Unlock()

	if v, ok := defaultSession[key]; ok {
		v.Set(platform, session)
		defaultSession[key] = v
	} else {
		p := Platform{lock: sync.RWMutex{}, data: make(map[string]Session)}
		p.Set(platform, session)
		defaultSession[key] = p
	}
}

func IsExist(key, platform string) bool {
	mutex.RLock()
	defer mutex.RUnlock()

	if v, ok := defaultSession[key]; ok {
		return v.IsExist(platform)
	}

	return false
}

func Delete(key, platform string) {
	mutex.Lock()
	defer mutex.Unlock()

	if v, ok := defaultSession[key]; ok {
		v.Delete(platform)
	}
}

func Default() map[string]Platform {
	return defaultSession
}
