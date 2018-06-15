package lru

type option interface {
	apply(*LRU)
}

type optionFn func(*LRU)

func (f optionFn) apply(l *LRU) {
	f(l)
}

func withCapacity(cap int) option {
	return optionFn(func(l *LRU) {
		l.cap = cap
	})
}

func withShards(n int) option {
	return optionFn(func(l *LRU) {
		l.nshards = n
	})
}
