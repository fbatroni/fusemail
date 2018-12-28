package utils

/*
DynamicLocks provides mutex locking/unlocking on given identifier by using a
sync.Map containing the identifier as key, and custom mutex-embedded structure
with lock count to keep track when to delete the mutex from the map.
Locking an already-locked id will block, like regular mutex.
*/

import (
	"sync"
	"sync/atomic"
)

// mutexCounter is internal used by DynamicLocks; it keeps
// reference count of how many locks are (or waiting to be) acquired.
type mutexCounter struct {
	*sync.Mutex
	c int32
}

// newMutexCounter is the factory method of mutexCounter.
func newMutexCounter() *mutexCounter {
	return &mutexCounter{Mutex: &sync.Mutex{}, c: 1}
}

// DynamicLocks performs mutex lock/unlock based on input identifier.
// Input identifier can be the same or different.
type DynamicLocks struct {
	lockMap sync.Map
	len     int64
}

// Lock performs a mutex lock by input id
func (m *DynamicLocks) Lock(id interface{}) {
	v, exist := m.lockMap.LoadOrStore(id, newMutexCounter())
	mc, _ := v.(*mutexCounter)
	if exist {
		atomic.AddInt32(&mc.c, 1)
	} else {
		atomic.AddInt64(&m.len, 1)
	}
	mc.Lock()
}

// Unlock performs a mutex unlock by input id
func (m *DynamicLocks) Unlock(id interface{}) {
	v, exist := m.lockMap.Load(id)
	if !exist {
		return
	}
	mc, _ := v.(*mutexCounter)
	mc.Unlock()

	if atomic.AddInt32(&mc.c, -1) == 0 {
		m.lockMap.Delete(id)
		atomic.AddInt64(&m.len, -1) // FIXME: un-sync'd with above delete
	}
}

// Len returns number of mutexes currenly exist
// NOTE: Len() is mainly for unit-test and debug, it has synchronize issue within Unlock()
func (m *DynamicLocks) Len() int64 {
	return atomic.LoadInt64(&m.len)
}
