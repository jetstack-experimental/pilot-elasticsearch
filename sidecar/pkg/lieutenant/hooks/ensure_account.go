package hooks

import (
	"encoding/json"
	"fmt"
	"net/url"

	"bytes"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant"
)

const (
	defaultElasticUsername = "elastic"
	defaultElasticPassword = "changeme"
)

// EnsureAccount will ensure an account with the given username and password
// exists, creating one if neccessary. It uses the default 'elastic' username
// and password to create the account if neccessary, so can be used for
// bootstrapping clusters
func EnsureAccount(user, pass string, roles ...string) lieutenant.Hook {
	return func(m lieutenant.Interface) error {
		data, err := json.Marshal(struct {
			Password string   `json:"password"`
			Roles    []string `json:"roles"`
			Enabled  bool     `json:"enabled"`
		}{
			Password: pass,
			Roles:    roles,
			Enabled:  true,
		})

		if err != nil {
			return fmt.Errorf("error encoding payload: %s", err.Error())
		}

		req, err := m.BuildRequest(
			"POST",
			fmt.Sprintf("/_xpack/security/user/%s", user),
			bytes.NewReader(data),
		)

		if err != nil {
			return fmt.Errorf("error constructing request: %s", err.Error())
		}

		// here we override the sidecar auth details to use the superuser account
		req.URL.User = url.UserPassword(defaultElasticUsername, defaultElasticPassword)

		resp, err := m.ESClient().Do(req)

		if err != nil {
			return fmt.Errorf("error creating sidecar user: %s", err.Error())
		}

		if resp.StatusCode >= 500 || resp.StatusCode < 200 || (resp.StatusCode >= 300 && resp.StatusCode < 400) {
			return fmt.Errorf("error creating sidecar user: code %d", resp.StatusCode)
		}

		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return validateUserAccount(m)
		}

		return nil
	}
}

func validateUserAccount(m lieutenant.Interface) error {
	req, err := m.BuildRequest(
		"GET",
		fmt.Sprintf("/_xpack/security/user/%s", m.Options().SidecarUsername()),
		nil,
	)

	if err != nil {
		return fmt.Errorf("error creating validation request: %s", err.Error())
	}

	resp, err := m.ESClient().Do(req)

	if err != nil {
		return fmt.Errorf("error validating sidecar user: %s", err.Error())
	}

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		return fmt.Errorf("error validating sidecar user: code %d", resp.StatusCode)
	}

	return nil
}
