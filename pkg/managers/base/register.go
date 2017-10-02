package base

import (
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

func init() {
	managers.Register(managers.VersionBase, func(managers.Options) (managers.Manager, error) {
		return &Base{}, nil
	})
}
