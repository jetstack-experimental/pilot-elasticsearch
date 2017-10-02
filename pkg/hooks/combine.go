package hooks

import (
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

func Combine(hooks ...managers.Hook) managers.Hook {
	return managers.NewHook(func(m managers.PilotClient) error {
		for _, h := range hooks {
			if err := h.Execute(m); err != nil {
				return err
			}
		}
		return nil
	}, managers.AllVersions...)
}
