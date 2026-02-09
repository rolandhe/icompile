package golang

type LinkedSetString struct {
	list []string
	got  map[string]struct{}
}

func NewLinkedSetString() *LinkedSetString {
	return &LinkedSetString{
		list: make([]string, 0),
		got:  make(map[string]struct{}),
	}
}

func (lset *LinkedSetString) add(name string) {
	if _, ok := lset.got[name]; ok {
		return
	}
	lset.list = append(lset.list, name)
}

func (lset *LinkedSetString) getList() []string {
	return lset.list
}

type linkedMapContainer[T any] struct {
	list []string
	got  map[string]T
}

func newLinkedMapContainer[T any]() *linkedMapContainer[T] {
	return &linkedMapContainer[T]{
		list: make([]string, 0),
		got:  make(map[string]T),
	}
}

func (lmap *linkedMapContainer[T]) add(name string, value T) {
	if _, ok := lmap.got[name]; !ok {
		lmap.list = append(lmap.list, name)
	}
	lmap.got[name] = value
}
func (lmap *linkedMapContainer[T]) getList() []string {
	return lmap.list
}
func (lmap *linkedMapContainer[T]) getValue(name string) T {
	return lmap.got[name]
}
