package hooks

import (
	log "github.com/Sirupsen/logrus"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/util"
)

// OnlyRoles will only execute the Hook if the node is of one
// of the specified roles
func OnlyRoles(h manager.Hook, roles ...util.Role) manager.Hook {
	return func(m manager.Interface) error {
		for _, r := range roles {
			for _, clusterRole := range m.Options().Roles() {
				if clusterRole == r {
					log.Debugf("executing hook for node with role: %s", clusterRole)
					return h(m)
				}
			}
		}
		return nil
	}
}
