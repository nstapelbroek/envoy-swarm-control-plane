package snapshot

import (
	"context"
	"sync"
	"testing"
	"time"

	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	corelog "github.com/envoyproxy/go-control-plane/pkg/log"
	"github.com/nstapelbroek/envoy-swarm-control-plane/pkg/logger"
	"github.com/stretchr/testify/assert"
)

type testLogger struct{}

func (testLogger) Debugf(string, ...interface{})            {}
func (testLogger) Infof(string, ...interface{})             {}
func (testLogger) Warnf(string, ...interface{})             {}
func (testLogger) Errorf(string, ...interface{})            {}
func (testLogger) Fatalf(string, ...interface{})            {}
func (testLogger) Panicf(string, ...interface{})            {}
func (l testLogger) WithFields(logger.Fields) logger.Logger { return l }

type mockADS struct {
	mu        sync.Mutex
	callCount int
	done      chan struct{}
}

func (m *mockADS) Provide(context.Context) ([]types.Resource, []types.Resource, error) {
	m.mu.Lock()
	m.callCount++
	m.mu.Unlock()
	select {
	case m.done <- struct{}{}:
	default:
	}
	return nil, nil, nil
}

func (m *mockADS) calls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.callCount
}

type mockSDS struct{}

func (mockSDS) Provide(context.Context) ([]types.Resource, error) { return nil, nil }
func (mockSDS) HasValidCertificate(*route.VirtualHost) bool       { return false }
func (mockSDS) GetCertificateConfig(*route.VirtualHost) *auth.SdsSecretConfig {
	return nil
}

func newTestManager(ads *mockADS) *Manager {
	snapCache := cache.NewSnapshotCache(false, StaticHash{}, corelog.LoggerFuncs{})
	return NewManager(ads, mockSDS{}, snapCache, testLogger{})
}

func TestListenCoalescesBurstEvents(t *testing.T) {
	ads := &mockADS{done: make(chan struct{}, 1)}
	mgr := newTestManager(ads)

	events := make(chan UpdateReason, 10)
	for i := 0; i < 5; i++ {
		events <- UpdateReason("burst event")
	}

	go mgr.Listen(events)

	select {
	case <-ads.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for discovery to run")
	}

	close(events)

	assert.Equal(t, 1, ads.calls(), "expected burst events to be coalesced into a single discovery run")
}

func TestListenProcessesMultipleBurstsSeparately(t *testing.T) {
	ads := &mockADS{done: make(chan struct{}, 1)}
	mgr := newTestManager(ads)

	events := make(chan UpdateReason, 20)
	go mgr.Listen(events)

	for i := 0; i < 3; i++ {
		events <- UpdateReason("burst 1")
	}
	waitDiscovery(t, ads)
	assert.Equal(t, 1, ads.calls())

	for i := 0; i < 3; i++ {
		events <- UpdateReason("burst 2")
	}
	waitDiscovery(t, ads)
	assert.Equal(t, 2, ads.calls())

	close(events)
}

func waitDiscovery(t *testing.T, ads *mockADS) {
	t.Helper()
	select {
	case <-ads.done:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for discovery to run")
	}
}
