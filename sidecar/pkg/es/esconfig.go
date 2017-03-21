package es

import (
	"fmt"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

// Env returns the environment variables to be used when running ElasticSearch
// with the provided Roles
// TODO: refactor this into the Manager
func Env(roles []util.Role) []string {
	env := []string{
		"ES_JAVA_OPTS=-Djava.net.preferIPv4Stack=true -Des.cgroups.hierarchy.override=/",
		fmt.Sprintf("NODE_MASTER=%v", contains(roles, util.RoleMaster)),
		fmt.Sprintf("NODE_INGEST=%v", contains(roles, util.RoleClient)),
		fmt.Sprintf("NODE_DATA=%v", contains(roles, util.RoleData)),
	}

	return env
}

func contains(roles []util.Role, role util.Role) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
