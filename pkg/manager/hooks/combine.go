package hooks

import (
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
)

func Combine(hooks ...manager.Hook) manager.Hook {
	return func(m manager.Interface) error {
		for _, h := range hooks {
			if err := h(m); err != nil {
				return err
			}
		}
		return nil
	}
}
