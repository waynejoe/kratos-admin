package utils

type Set[T comparable] struct {
	data map[T]struct{}
}

func (s *Set[T]) Add(list ...T) {
	for _, item := range list {
		s.data[item] = struct{}{}
	}
}

func (s *Set[T]) Remove(e T) {
	delete(s.data, e)
}

func (s *Set[T]) Has(e T) bool {
	_, ok := s.data[e]
	return ok
}

func (s *Set[T]) List() []T {
	list := make([]T, 0, len(s.data))
	for item := range s.data {
		list = append(list, item)
	}
	return list
}

func NewSet[T comparable](list ...T) *Set[T] {
	set := &Set[T]{data: make(map[T]struct{})}
	set.Add(list...)
	return set
}
