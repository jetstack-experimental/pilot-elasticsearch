package manager

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"
	"gopkg.in/olivere/elastic.v5"
	"k8s.io/client-go/kubernetes"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/es"
	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/probe"
)

const (
	esHost = "127.0.0.1"
	esPort = 9200
)

// Interface describes a manager for an Elasticsearch process
type Interface interface {
	Client() (*elastic.Client, error)

	// Options returns a set of options for this manager
	Options() Options
	// ESClient returns an HTTP client for communicating with Elasticsearch
	ESClient() *http.Client
	// KubeClient returns a kubernetes Clientset that can be used to
	// communicate with the clusters apiserver
	KubeClient() *kubernetes.Clientset
	// BuildRequest creates a Request type used to talk to Elasticsearch
	BuildRequest(method, path, query string, body io.Reader) (*http.Request, error)
	// RegisterHook will register a hook to execute in a particular phase
	RegisterHooks(Phase, ...Hook)
	// Phase returns the current phase of the Elasticsearch process
	Phase() Phase
	// Run will start Elasticsearch with the managers provided configuration.
	// It will block until Elasticsearch is exited, for whatever reason, It is
	// responsible for firing preStart and postStart hooks.
	Run() error
	// Shutdown will handle a shutdown signal. It will block until it is safe to shut
	// down this node and will fire preStop and postStop hooks.
	Shutdown() error
	ReadinessCheck() probe.Check
	LivenessCheck() probe.Check
}

var _ Interface = &Manager{}

// Manager is the default implementation of an Elasticsearch process manager
type Manager struct {
	esClient   *elastic.Client
	options    Options
	kubeClient *kubernetes.Clientset

	hookLock sync.RWMutex
	hooks    map[Phase][]Hook
	phase    Phase

	esCmd *exec.Cmd
}

// NewManager constructs a new Manager instance with the given Options
// and Kubernetes API client
func NewManager(opts Options, kubeClient *kubernetes.Clientset) Interface {
	return &Manager{
		options:    opts,
		kubeClient: kubeClient,
		hooks:      make(map[Phase][]Hook),
	}
}

func (m *Manager) Client() (*elastic.Client, error) {
	// TODO: Mutex around the esClient
	if m.esClient == nil {
		cl, err := elastic.NewClient(elastic.SetURL("http://127.0.0.1:9200/"))

		if err != nil {
			return nil, err
		}

		m.esClient = cl
	}
	return m.esClient, nil
}

// Options returns a set of options for this manager
func (m *Manager) Options() Options {
	return m.options
}

// ESClient returns an HTTP client for communicating with Elasticsearch
func (m *Manager) ESClient() *http.Client {
	return http.DefaultClient
}

// KubeClient returns a kubernetes Clientset that can be used to
// communicate with the clusters apiserver
func (m *Manager) KubeClient() *kubernetes.Clientset {
	return m.kubeClient
}

// BuildRequest builds an authenticated http.Request for the Elasticsearch cluster
func (m *Manager) BuildRequest(method, path, query string, body io.Reader) (*http.Request, error) {
	// TODO: refactor scheme & host out of this method
	builtURL := url.URL{
		Scheme:   m.Options().ClusterURL().Scheme,
		Host:     m.Options().ClusterURL().Host,
		RawQuery: query,
		Path:     path,
	}

	log.Debugf("created request: %s", builtURL.String())

	if m.Options().SidecarUsername() != "" {
		builtURL.User = url.UserPassword(m.Options().SidecarUsername(), m.Options().SidecarPassword())
	}

	return http.NewRequest(method, builtURL.String(), body)
}

// RegisterHooks will register a hook to execute in a particular phase
func (m *Manager) RegisterHooks(p Phase, h ...Hook) {
	m.hookLock.Lock()
	defer m.hookLock.Unlock()
	hooks := []Hook{}
	if existingHooks, ok := m.hooks[p]; ok {
		hooks = existingHooks
	}
	hooks = append(hooks, h...)
	m.hooks[p] = hooks
}

func (m *Manager) Phase() Phase {
	return m.phase
}

// Run will start Elasticsearch with the managers provided configuration.
// It will block until Elasticsearch is exited, for whatever reason, It is
// responsible for firing preStart and postStart hooks.
func (m *Manager) Run() error {
	if err := m.transitionPhase(PhasePreStart); err != nil {
		return fmt.Errorf("error running: %s", err.Error())
	}

	m.esCmd = exec.Command(m.Options().ElasticsearchBin())
	m.esCmd.Stdout = os.Stdout
	m.esCmd.Stderr = os.Stderr
	m.esCmd.Env = append(os.Environ(), es.Env(m.Options().Roles())...)

	go m.handleSignals()
	go m.firePostStart()

	log.Debugf("starting elasticsearch...")
	return m.esCmd.Run()
}

// Shutdown will handle a shutdown signal. It will block until it is safe to shut
// down this node and will fire preStop and postStop hooks.
func (m *Manager) Shutdown() error {
	log.Infof("initiating shutdown...")
	defer os.Exit(1)
	if err := m.transitionPhase(PhasePreStop); err != nil {
		return fmt.Errorf("error running lieutenant pre stop hooks: %s", err.Error())
	}

	if m.esCmd != nil {
		log.Debugf("sending SIGTERM to elasticsearch process")
		m.esCmd.Process.Signal(syscall.SIGTERM)
		log.Debugf("waiting for elasticsearch process to exit...")
		state, err := m.esCmd.Process.Wait()

		if err != nil {
			return fmt.Errorf("elasticsearch exited with error: %s", err.Error())
		}

		if !state.Exited() {
			return fmt.Errorf("warning: elasticsearch has not exited")
		}
	}

	if err := m.transitionPhase(PhasePostStop); err != nil {
		return fmt.Errorf("error running lieutenant post stop hooks: %s", err.Error())
	}

	return nil
}

// firePostStart will wait until the Elasticsearch process is accessible,
// and then fire the postStart hooks
func (m *Manager) firePostStart() {
	for {
		log.Debugf("waiting for port 9200 before executing post-start hook...")
		if m.listening() {
			break
		}

		time.Sleep(time.Second * 1)
	}

	if err := m.transitionPhase(PhasePostStart); err != nil {
		log.Printf("error transitioning to post-start phase: %s", err.Error())
		// TODO: notice this error
		m.Shutdown()
	}
}

// listening will return true if the Elasticsearch process is accessible
// on the HTTP client port
func (m *Manager) listening() bool {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", esHost, esPort))
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

// transitionPhase will fire all hooks for the given Phase p
func (m *Manager) transitionPhase(p Phase) error {
	log.Debugf("transitioning to phase '%s'", p)
	m.hookLock.RLock()
	defer m.hookLock.RUnlock()
	if hooks, ok := m.hooks[p]; ok {
		log.Debugf("executing %d '%s' hooks", len(hooks), p)
		for _, h := range hooks {
			log.Debugf("executing hook...")
			err := h(m)

			if err != nil {
				err = fmt.Errorf("error running hook for phase '%s': %s", p, err.Error())
				log.Warnf(err.Error())
				return err
			}
		}
	}
	return nil
}

// handleSignals is responsible for firing shutdown events and handling
// any OS signals.
// TODO: Refactor this elsewhere
func (m *Manager) handleSignals() {
	sigChan := make(chan os.Signal, 1)
	defer close(sigChan)
	defer signal.Stop(sigChan)

	signal.Ignore(
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT,
	)

	for sig := range sigChan {
		log.Debugf("got signal: '%s'", sig.String())
		defer os.Exit(1)
		if err := m.Shutdown(); err != nil {
			log.Fatalf("error shutting down: %s", err.Error())
		}
	}
}
