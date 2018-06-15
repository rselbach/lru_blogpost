package lru

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// a shard of the list
type shard struct {
	cap        int
	len        int32
	sync.Mutex                               // protects the idem and list
	idx        map[interface{}]*list.Element // the index for our list
	l          *list.List                    // the actual list holding the data
}

// newShard creates a new shard
func newShard(cap int) *shard {
	s := &shard{
		cap: cap,
		l:   list.New(),
		idx: make(map[interface{}]*list.Element, cap+1),
	}
	return s
}

// Len returns the number of items currently in the LRU
func (s *shard) Len() int { return int(atomic.LoadInt32(&s.len)) }

// add will insert a new keyval pair to the shard
func (s *shard) add(k, v interface{}) {
	s.Lock()
	defer s.Unlock()

	// first let's see if we already have this key
	if le, ok := s.idx[k]; ok {
		// update the entry and move it to the front
		le.Value.(*entry).val = v
		s.l.MoveToFront(le)
		return
	}
	s.idx[k] = s.l.PushFront(&entry{key: k, val: v})
	atomic.AddInt32(&s.len, 1)

	if s.cap > 0 && s.Len() > s.cap {
		s.removeOldest()
	}
	return
}

// front will return the element at the front of the queue without modifying
// it in anyway
func (s *shard) front() (key, val interface{}) {
	s.Lock()
	defer s.Unlock()

	if s.Len() == 0 {
		return nil, nil
	}

	le := s.l.Front()
	return le.Value.(*entry).key, le.Value.(*entry).val
}

// get will try to retrieve a value from the given key. The second return is
// true if the key was found.
func (s *shard) get(key interface{}) (value interface{}, ok bool) {
	s.Lock()
	defer s.Unlock()

	if le, found := s.idx[key]; found {
		s.l.MoveToFront(le)
		return le.Value.(*entry).val, true
	}
	return nil, false
}

func (s *shard) removeOldest() (key, val interface{}) {
	le := s.l.Back()
	if le == nil {
		return
	}
	return s.removeElement(le)
}

func (s *shard) removeElement(le *list.Element) (key, val interface{}) {
	e := le.Value.(*entry)
	s.l.Remove(le)
	delete(s.idx, e.key)
	atomic.AddInt32(&s.len, -1)
	return e.key, e.val
}

// removeKey will remove the given key from the LRU
func (s *shard) removeKey(key interface{}) {
	s.Lock()
	defer s.Unlock()

	le, ok := s.idx[key]
	if !ok {
		return
	}
	s.removeElement(le)
}
