package hooks

import (
	"fmt"
	"net/url"
	"strings"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/lieutenant"
)

const (
	defaultElasticUsername = "elastic"
	defaultElasticPassword = "changeme"
)

func CreateAdminAccount(m lieutenant.Interface) error {
	// TODO: use encoding/json to encode payloads
	req, err := m.BuildRequest(
		"POST",
		fmt.Sprintf("/_xpack/security/user/%s", m.Options().SidecarUsername()),
		strings.NewReader(
			fmt.Sprintf(`
			{
				"password": "%s",
				"roles": [ "superuser" ],
				"enabled": true
			}`, m.Options().SidecarPassword()),
		),
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
		return ValidateUserAccount(m)
	}

	return nil
}

func ValidateUserAccount(m lieutenant.Interface) error {
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
