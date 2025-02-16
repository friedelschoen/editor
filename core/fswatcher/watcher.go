package fswatcher

type Watcher interface {
	Add(name string) error
	Remove(name string) error
	Events() <-chan any
	OpMask() *Op
	Close() error
}

type Event struct {
	Op   Op
	Name string
}

type Op uint16

func (op Op) HasAny(op2 Op) bool { return op&op2 != 0 }
func (op *Op) Add(op2 Op)        { *op |= op2 }
func (op *Op) Remove(op2 Op)     { *op &^= op2 }

const (
	Attrib Op = 1 << iota
	Create
	Modify // write, truncate
	Remove
	Rename

	AllOps Op = Attrib | Create | Modify | Remove | Rename
)
