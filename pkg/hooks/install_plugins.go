package hooks

import (
	"fmt"

	"github.com/jetstack-experimental/navigator/pkg/apis/navigator/v1alpha1"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/managers"
)

// InstallPlugins returns a hook that will install the plugins specified in
// plugins. If any of the plugins fail to install, an error will be returned
func InstallPlugins(plugins ...v1alpha1.ElasticsearchClusterPlugin) managers.Hook {
	return managers.NewHook(func(m managers.PilotClient) error {
		for _, plugin := range plugins {
			if err := m.InstallPlugin(plugin); err != nil {
				return fmt.Errorf("error installing %s: %s", plugin, err.Error())
			}
		}
		return nil
	}, managers.Version5)
}
