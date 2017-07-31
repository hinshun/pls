package mitmproxy

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/pkg/failsafe"
	"github.com/hinshun/pls/pkg/namegen"
	"github.com/palantir/stacktrace"
)

const (
	MITMProxyImageName          = "mitmproxy/mitmproxy"
	MITMProxyPort               = 8080
	MITMProxyPrefix             = "mitm"
	MITMProxyDefaultCADirectory = "/home/mitmproxy/.mitmproxy"
	MITMProxyDefaultCAFilename  = "mitmproxy-ca-cert.pem"
)

type MITMProxy struct {
	ID      string
	Name    string
	Network string

	rootCTX context.Context
	rootCLI client.APIClient
}

func New(ctx context.Context, cli client.APIClient) (*MITMProxy, error) {
	mitmProxyName, err := namegen.GetUnusedContainerName(ctx, cli, MITMProxyPrefix)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to generate mitmproxy container name")
	}

	_, err = cli.NetworkCreate(ctx, mitmProxyName, types.NetworkCreate{
		Labels: map[string]string{
			"pls": MITMProxyPrefix,
		},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mitmproxy network")
	}

	cfg := &container.Config{
		Image: MITMProxyImageName,
		Labels: map[string]string{
			"pls": MITMProxyPrefix,
		},
		Cmd:       []string{"mitmdump"},
		OpenStdin: true,
	}
	hostCfg := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: mitmProxyName,
				Target: MITMProxyDefaultCADirectory,
			},
		},
	}
	netCfg := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			mitmProxyName: {},
		},
	}

	createResp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, mitmProxyName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mitmproxy container")
	}

	err = cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to start mitmproxy container")
	}

	return NewFromExisting(ctx, cli, mitmProxyName)
}

func NewFromExisting(ctx context.Context, cli client.APIClient, containerName string) (*MITMProxy, error) {
	containerJSON, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return nil, stacktrace.NewError("failed to inspect mitmproxy container")
	}

	_, networkExist := containerJSON.NetworkSettings.Networks[containerName]
	if !networkExist {
		return nil, stacktrace.NewError("failed to get mitmproxy network")
	}

	return &MITMProxy{
		ID:      containerJSON.ID,
		Name:    containerName,
		Network: containerName,

		rootCTX: ctx,
		rootCLI: cli,
	}, nil
}

func (m *MITMProxy) GetCACertificate() (io.ReadCloser, error) {
	caPath := fmt.Sprintf("%s/%s", MITMProxyDefaultCADirectory, MITMProxyDefaultCAFilename)

	var caStream io.ReadCloser

	retryPolicy := failsafe.NewRetryPolicy().WithDelay(time.Second)
	err := failsafe.New(retryPolicy).Run(m.rootCTX, func() error {
		var err error
		caStream, _, err = m.rootCLI.CopyFromContainer(m.rootCTX, m.ID, caPath)
		return err
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to copy ca certificate from mitmproxy container")
	}

	return caStream, nil
}
