package mappool

import "sync"

var tagMapPool = sync.Pool{
	New: func() any {
		// The buffer must be at least a block long.
		tagMap := make(map[string]struct{}, 5)
		return tagMap
	},
}

func Get() map[string]struct{} {
	return tagMapPool.Get().(map[string]struct{})
}

func Put(m map[string]struct{}) {
	for key := range m {
		delete(m, key)
	}

	tagMapPool.Put(m)
}
