package managers

type hook struct {
	fn        func(PilotClient) error
	supported []Version
}

func (f *hook) Execute(p PilotClient) error {
	return f.fn(p)
}

func (f *hook) Supported(v Version) bool {
	for _, vs := range f.supported {
		if vs == v {
			return true
		}
	}
	return false
}

// NewHook returns a new hook with the given hook function and supported versions
func NewHook(fn func(PilotClient) error, supported ...Version) Hook {
	return &hook{fn, supported}
}
