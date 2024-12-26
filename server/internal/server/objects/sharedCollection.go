package objects

import "sync"

// a genereic type . thread safe map ob objects with auti incrementing id

type SharedCollection[T any] struct {
	objectMap map[uint64]T
	nextId    uint64
	mapMux    sync.Mutex
}

func NewSharedCollection[T any](capacity ...int) *SharedCollection[T] {
	var newObjMap map[uint64]T
	if len(capacity) > 0 {
		newObjMap = make(map[uint64]T, capacity[0])
	} else {
		newObjMap = make(map[uint64]T)
	}

	return &SharedCollection[T]{
		objectMap: newObjMap,
		nextId:    1,
	}
}

func (s *SharedCollection[T]) Add(obj T, ids ...uint64) uint64 {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()

	thisId := s.nextId
	if len(ids) > 0 {
		thisId = ids[0]
	}
	s.objectMap[thisId] = obj
	s.nextId++
	return thisId
}

func (s *SharedCollection[T]) Remove(id uint64) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()
	delete(s.objectMap, id)
}

func (s *SharedCollection[T]) ForEach(callback func(uint64, T)) {
	s.mapMux.Lock()
	// Create a local copy while holding the lock.
	localCopy := make(map[uint64]T, len(s.objectMap))

	for id, obj := range s.objectMap {
		localCopy[id] = obj
	}
	s.mapMux.Unlock()
	// Iterate over the local copy without holding the lock.
	for id, obj := range localCopy {
		callback(id, obj)
	}

}

func (s *SharedCollection[T]) Get(id uint64) (T, bool) {
	s.mapMux.Lock()
	defer s.mapMux.Unlock()
	obj, ok := s.objectMap[id]
	return obj, ok
}

func (s *SharedCollection[T]) Lenght() int {
	return len(s.objectMap)
}
