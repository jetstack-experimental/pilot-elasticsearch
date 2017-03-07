package hooks

import (
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/manager"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

// OnlyRoles will only execute the Hook if the node is of one
// of the specified roles
func OnlyRoles(h manager.Hook, roles ...util.Role) manager.Hook {
	return func(m manager.Interface) error {
		for _, r := range roles {
			if m.Options().Role() == r {
				return h(m)
			}
		}
		return nil
	}
}
