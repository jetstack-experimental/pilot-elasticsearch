package lieutenant

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"

	"gitlab.jetstack.net/marshal/lieutenant-elastic-search/sidecar/pkg/es"
)

const (
	esHost = "127.0.0.1"
	esPort = 9200
)

type Interface interface {
	// Options returns a set of options for this manager
	Options() Options
	// ESClient returns an HTTP client for communicating with Elasticsearch
	ESClient() *http.Client
	// KubeClient returns a kubernetes Clientset that can be used to
	// communicate with the clusters apiserver
	KubeClient() *kubernetes.Clientset
	// BuildRequest creates a Request type used to talk to Elasticsearch
	BuildRequest(method, path string, body io.Reader) (*http.Request, error)
	// RegisterHook will register a hook to execute in a particular phase
	RegisterHook(Phase, Hook)
	// Run will start Elasticsearch with the managers provided configuration.
	// It will block until Elasticsearch is exited, for whatever reason
	Run() error
	// Shutdown will handle a shutdown signal. It will block until it is safe to shut
	// down this node and will handle starting migrations of data etc.
	Shutdown() error
}

var _ Interface = &Manager{}

type Manager struct {
	options    Options
	kubeClient *kubernetes.Clientset

	hookLock sync.RWMutex
	hooks    map[Phase][]Hook

	esCmd *exec.Cmd
}

func NewManager(opts Options, kubeClient *kubernetes.Clientset) Interface {
	return &Manager{
		options:    opts,
		kubeClient: kubeClient,
		hooks:      make(map[Phase][]Hook),
	}
}

func (m *Manager) Options() Options {
	return m.options
}

func (m *Manager) ESClient() *http.Client {
	return http.DefaultClient
}

func (m *Manager) KubeClient() *kubernetes.Clientset {
	return m.kubeClient
}

func (m *Manager) RegisterHook(p Phase, h Hook) {
	m.hookLock.Lock()
	defer m.hookLock.Unlock()
	hooks := []Hook{}
	if existingHooks, ok := m.hooks[p]; ok {
		hooks = existingHooks
	}
	hooks = append(hooks, h)
	m.hooks[p] = hooks
}

func (m *Manager) BuildRequest(method, path string, body io.Reader) (*http.Request, error) {
	// TODO: refactor scheme & host out of this method
	builtURL := url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s:%d", esHost, esPort),
		Path:   path,
		User:   url.UserPassword(m.Options().SidecarUsername(), m.Options().SidecarPassword()),
	}

	return http.NewRequest(method, builtURL.String(), body)
}

func (m *Manager) Run() error {
	if err := m.transitionPhase(PhasePreStart); err != nil {
		return fmt.Errorf("error running: %s", err.Error())
	}

	m.esCmd = exec.Command(m.Options().ElasticsearchBin())
	m.esCmd.Stdout = os.Stdout
	m.esCmd.Stderr = os.Stderr
	m.esCmd.Env = append(os.Environ(), es.Env(m.Options().Role())...)

	go m.handleSignals()
	go m.firePostStart()

	return m.esCmd.Run()
}

func (m *Manager) firePostStart() {
	for {
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
	m.hookLock.RLock()
	defer m.hookLock.RUnlock()
	if hooks, ok := m.hooks[p]; ok {
		for _, h := range hooks {
			err := h(m)

			if err != nil {
				return fmt.Errorf("error running hook for phase '%s': %s", p, err.Error())
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

	for _ = range sigChan {
		defer os.Exit(1)
		if err := m.Shutdown(); err != nil {
			log.Fatalf("error shutting down: %s", err.Error())
		}
	}
}
