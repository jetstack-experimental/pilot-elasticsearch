package pilot

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/jetstack-experimental/navigator/pkg/apis/marshal/v1alpha1"

	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager/hooks"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/manager/hooks/events"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/probe"
	"github.com/jetstack-experimental/pilot-elasticsearch/pkg/util"
)

var (
	apiServerHost                  string
	namespace                      string
	podName                        string
	roles                          []util.Role
	plugins                        []v1alpha1.ElasticsearchClusterPlugin
	controllerName, controllerKind string

	esSidecarUsername = "_sidecar"
	esSidecarPassword string
	clusterURL        string

	pluginsFlag string
	rolesFlag   string

	startCmd = &cobra.Command{
		Use:   "start",
		Short: "starts the elasticsearch pilot",
		Run: func(cmd *cobra.Command, args []string) {
			if err := parsePluginsFlag(); err != nil {
				log.Fatalf("error parsing plugins list: %s", err.Error())
			}

			if err := parseRolesFlag(); err != nil {
				log.Fatalf("error parsing roles list: %s", err.Error())
			}

			kubeClient, err := util.NewKubernetesClient(apiServerHost)

			if err != nil {
				log.Fatalf("error creating kubernetes client: %s", err.Error())
			}

			managerOpts, err := manager.NewOptions(
				manager.SetControllerKind(controllerKind),
				manager.SetControllerName(controllerName),
				manager.SetPodName(podName),
				manager.SetNamespace(namespace),
				manager.SetRoles(roles),
				manager.SetClusterURL(clusterURL),
			)

			if err != nil {
				log.Fatalf("error constructing manager options: %s", err.Error())
			}

			m := manager.NewManager(
				managerOpts,
				kubeClient,
			)

			// Install plugins before Elasticsearch starts
			m.RegisterHooks(manager.PhasePreStart,
				hooks.InstallPlugins(plugins...),
			)

			// Ensure user exists for the pilot
			// m.RegisterHooks(manager.PhasePostStart,
			// 	hooks.AllowErrors(
			// 		// Only run on data nodes as a shard is required to exist
			// 		// in order to write user data
			// 		hooks.OnlyRoles(
			// 			hooks.Retry(
			// 				hooks.EnsureAccount(esSidecarUsername, esSidecarPassword, "superuser"),
			// 				time.Second*2, // wait 2 seconds between each attempt
			// 				10,            // try 10 times
			// 			),
			// 			util.RoleData,
			// 		),
			// 	),
			// )

			m.RegisterHooks(manager.PhasePostStart,
				hooks.OnlyRoles(
					hooks.AcceptShards,
					util.RoleData,
				),
			)

			m.RegisterHooks(manager.PhasePreStop,
				hooks.OnlyRoles(
					hooks.OnEvent(
						events.ScaleDownEvent,
						hooks.DrainShards,
					),
					util.RoleData,
				),
			)

			// Start readiness checker
			go (&probe.Listener{
				Port:  12001,
				Check: m.ReadinessCheck(),
			}).Listen()

			// Start liveness checker
			go (&probe.Listener{
				Port:  12000,
				Check: m.LivenessCheck(),
			}).Listen()

			// TODO: Once the ES process is exited, we're immediately exiting the main process
			// without allowing time for postStop hooks to run. We should block until postStop hooks
			// are complete
			if err := m.Run(); err != nil {
				log.Fatalf("error running elasticsearch: %s", err.Error())
			}
		},
	}
)

func parsePluginsFlag() error {
	if len(pluginsFlag) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(pluginsFlag), &plugins)
}

func parseRolesFlag() error {
	if len(rolesFlag) == 0 {
		return nil
	}
	return json.Unmarshal([]byte(rolesFlag), &roles)
}

func init() {
	// StartCmd flags
	startCmd.PersistentFlags().StringVar(&podName, "podName", "", "The name of this pod")
	startCmd.PersistentFlags().StringVar(&namespace, "namespace", "", "The namespace this node is running in")
	startCmd.PersistentFlags().StringVar(&apiServerHost, "apiServerHost", "", "Kubernetes apiserver host address (overrides autodetection)")
	startCmd.PersistentFlags().StringVarP(&pluginsFlag, "plugins", "p", "[]", "List of Elasticsearch plugins to install")
	startCmd.PersistentFlags().StringVarP(&rolesFlag, "roles", "r", `["client"]`, "The role of this Elasticsearch node")
	startCmd.PersistentFlags().StringVar(&controllerName, "controllerName", "", "Name of the controller managing this node")
	startCmd.PersistentFlags().StringVar(&controllerKind, "controllerKind", "", "Kind of the controller managing this node")
	startCmd.PersistentFlags().StringVar(&clusterURL, "clusterURL", "", "URL for communicating with Elasticsearch client nodes")
	startCmd.PersistentFlags().StringVar(&esSidecarPassword, "sidecarUserPassword", "insecure", "The password to use for the sidecars ElasticSearch user account")

	rootCmd.AddCommand(startCmd)
}
