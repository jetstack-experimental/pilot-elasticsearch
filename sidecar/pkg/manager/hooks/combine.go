package hooks

import (
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
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
