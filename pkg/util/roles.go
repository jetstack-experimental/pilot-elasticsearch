package util

import "fmt"

// Role represents the role of an Elasticsearch node
type Role string

const (
	// RoleData should be set for ES data nodes
	RoleData Role = "data"
	// RoleClient should be set for ES client nodes
	RoleClient Role = "client"
	// RoleMaster should be set for ES master nodes
	RoleMaster Role = "master"
)

// RoleVar is an implementation of pflags.Value with validation
// for node rules
type RoleVar struct {
	role Role
}

// String returns the role set in this flag
func (r *RoleVar) String() string {
	if r.role == "" {
		return string(RoleClient)
	}

	return string(r.role)
}

// Set sets the role of this flag, additionally validating the provided value
func (r *RoleVar) Set(s string) error {
	switch Role(s) {
	case RoleData:
		r.role = RoleData
	case RoleClient:
		r.role = RoleClient
	case RoleMaster:
		r.role = RoleMaster
	default:
		return fmt.Errorf("role should be one of '%s', '%s' or '%s'", RoleData, RoleClient, RoleMaster)
	}

	return nil
}

// Type returns a description of this var type
func (r *RoleVar) Type() string {
	return "nodeRole"
}
