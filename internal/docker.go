package internal

import (
	"context"
	"github.com/containous/traefik/v2/pkg/config/dynamic"
	"github.com/containous/traefik/v2/pkg/provider/docker"
	"github.com/containous/traefik/v2/pkg/safe"
)

func TestDockerIntegration() error {
	p := docker.Provider{}
	p.SetDefaults()
	err := p.Init()
	if err != nil {
		return err
	}

	configChan := make(chan dynamic.Message)
	pool := safe.NewPool(context.Background())
	_ = p.Provide(configChan, pool)

	henk := <- configChan
	println(henk.Configuration)
	return nil
}
