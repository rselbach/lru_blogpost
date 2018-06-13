package lru

import (
	"container/list"
	"sync"
)

// LRU is a last-recently-used list of items
type LRU struct {
	cap int // the max number of items to hold

	sync.Mutex                               // protects the idem and list
	idx        map[interface{}]*list.Element // the index for our list
	l          *list.List                    // the actual list holding the data
}

type entry struct {
	key, val interface{}
}

// New creates a new LRU with the provided capacity. If cap less than 1, then the LRU
// grows indefinitely
func New(cap int) *LRU {
	l := &LRU{
		cap: cap,
		l:   list.New(),
		idx: make(map[interface{}]*list.Element, cap+1),
	}
	return l
}

// Len returns the number of items currently in the LRU
func (l *LRU) Len() int { l.lazyInit(); return l.l.Len() }

// this initializes some fields at first use. Helpful to
// allow us to use the empty value of LRU
func (l *LRU) lazyInit() {
	if l.l == nil {
		l.l = list.New()
		l.idx = make(map[interface{}]*list.Element, l.cap+1)
	}
}

// Add will insert a new keyval pair to the LRU
func (l *LRU) Add(k, v interface{}) {
	l.Lock()
	defer l.Unlock()
	l.lazyInit()

	// first let's see if we already have this key
	if le, ok := l.idx[k]; ok {
		// update the entry and move it to the front
		le.Value.(*entry).val = v
		l.l.MoveToFront(le)
		return
	}
	l.idx[k] = l.l.PushFront(&entry{key: k, val: v})

	if l.cap > 0 && l.l.Len() > l.cap {
		l.removeOldest()
	}
	return
}

// PeekFront will return the element at the front of the queue without modifying
// it in anyway
func (l *LRU) PeekFront() (key, val interface{}) {
	l.Lock()
	defer l.Unlock()
	l.lazyInit()
	if l.l.Len() == 0 {
		return nil, nil
	}

	le := l.l.Front()
	return le.Value.(*entry).key, le.Value.(*entry).val
}

// Get will try to retrieve a value from the given key. The second return is
// true if the key was found.
func (l *LRU) Get(key interface{}) (value interface{}, ok bool) {
	l.Lock()
	defer l.Unlock()
	l.lazyInit()
	if le, found := l.idx[key]; found {
		l.l.MoveToFront(le)
		return le.Value.(*entry).val, true
	}
	return nil, false
}

func (l *LRU) removeOldest() (key, val interface{}) {
	le := l.l.Back()
	if le == nil {
		return
	}
	return l.remove(le)
}

func (l *LRU) remove(le *list.Element) (key, val interface{}) {
	e := le.Value.(*entry)
	l.l.Remove(le)
	delete(l.idx, e.key)
	return e.key, e.val
}

// Remove will remove the given key from the LRU
func (l *LRU) Remove(key interface{}) {
	l.Lock()
	defer l.Unlock()
	l.lazyInit()
	le, ok := l.idx[key]
	if !ok {
		return
	}
	l.remove(le)
}

// TraverseFunc is the function called for each element when
// traversing an LRU
type TraverseFunc func(key, val interface{}) bool

// Traverse will call fn for each element in the LRU, from most recently used to
// least. If fn returns false, the traverse stops
func (l *LRU) Traverse(fn TraverseFunc) {
	l.Lock()
	defer l.Unlock()
	le := l.l.Front()
	for {
		if le == nil {
			break
		}

		e := le.Value.(*entry)
		if !fn(e.key, e.val) {
			break
		}
		le = le.Next()
	}
}

// TraverseReverse will call fn for each element in the LRU, from least recently used to
// most. If fn returns false, the traverse stops
func (l *LRU) TraverseReverse(fn TraverseFunc) {
	l.Lock()
	defer l.Unlock()
	le := l.l.Back()
	for {
		if le == nil {
			break
		}

		e := le.Value.(*entry)
		if !fn(e.key, e.val) {
			break
		}
		le = le.Prev()
	}
}
