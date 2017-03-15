package es

import (
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/util"
)

// Env returns the environment variables to be used when running ElasticSearch
// with the provided Roles
// TODO: refactor this into the Manager
func Env(roles []util.Role) []string {
	env := []string{
		"DISCOVERY_PROVIDER=kubernetes",
		"ES_JAVA_OPTS=-Djava.net.preferIPv4Stack=true",
	}

	for _, role := range roles {
		switch role {
		case util.RoleMaster:
			env = append(env, "NODE_MASTER=true", "NODE_DATA=false")
		case util.RoleData:
			env = append(env, "NODE_DATA=true", "NODE_MASTER=false")
		case util.RoleClient:
			env = append(env, "NODE_DATA=false", "NODE_MASTER=false")
		}
	}

	return env
}
