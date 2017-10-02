package managers

import "fmt"

var (
	factories map[Version]FactoryFn
)

type FactoryFn func(Options) (Manager, error)

func Register(v Version, fn FactoryFn) {
	factories[v] = fn
}

func New(o Options, v Version) (Manager, error) {
	if f, ok := factories[v]; ok {
		return f(o)
	}
	return nil, fmt.Errorf("unrecognised version '%s'", string(v))
}
