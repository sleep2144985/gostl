package container

type BaseIterator interface {
}

type ConstIterator interface {
	Next() ConstIterator
	Value() interface{}
	Equal(ConstIterator) bool
}

type Iterator interface {
	Next() Iterator
	Value() interface{}
	Equal(Iterator) bool
	Set(data interface{}) error
}
