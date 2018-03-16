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
	volumetypes "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/pkg/failsafe"
	"github.com/hinshun/pls/pkg/namegen"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"
)

const (
	MITMProxyImageName          = "mitmproxy/mitmproxy:2.0.2"
	MITMProxyPort               = 8080
	MITMProxyPrefix             = "mitm"
	MITMProxyDefaultCADirectory = "/home/mitmproxy/.mitmproxy"
	MITMProxyDefaultCAFilename  = "mitmproxy-ca-cert.pem"
)

type MITMProxySpec struct {
	Name string
}

type MITMProxy struct {
	ID      string
	Name    string
	Network string

	rootCTX context.Context
	rootCLI client.APIClient
}

func New(ctx context.Context, cli client.APIClient, spec MITMProxySpec) (*MITMProxy, error) {
	proxyName := spec.Name
	if proxyName == "" {
		var err error
		proxyName, err = namegen.GetUnusedContainerName(ctx, cli, MITMProxyPrefix)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate mitmproxy container name")
		}
	}

	logrus.Infof("Creating network '%s'", proxyName)
	networkResp, err := cli.NetworkCreate(ctx, proxyName, types.NetworkCreate{
		Labels: map[string]string{
			"pls": MITMProxyPrefix,
		},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mitmproxy network")
	}

	logrus.Infof("Creating volume '%s'", proxyName)
	_, err = cli.VolumeCreate(ctx, volumetypes.VolumesCreateBody{
		Name: proxyName,
		Labels: map[string]string{
			"pls": MITMProxyPrefix,
		},
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mitmproxy volume")
	}

	cfg := &container.Config{
		Image: MITMProxyImageName,
		Labels: map[string]string{
			"pls": MITMProxyPrefix,
		},
		Cmd:         []string{"mitmproxy", "--insecure"},
		Tty:         true,
		AttachStdin: true,
		OpenStdin:   true,
	}
	hostCfg := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: proxyName,
				Target: MITMProxyDefaultCADirectory,
			},
		},
	}
	netCfg := &network.NetworkingConfig{}

	logrus.Infof("Creating container '%s'", proxyName)
	createResp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, proxyName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create mitmproxy container")
	}

	err = cli.NetworkConnect(ctx, networkResp.ID, createResp.ID, &network.EndpointSettings{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to connect mitmproxy to mitmproxy network")
	}

	logrus.Infof("Starting mitmproxy")
	err = cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to start mitmproxy container")
	}

	return NewFromExisting(ctx, cli, proxyName)
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

func (m *MITMProxy) GetCACertificateTar() (io.ReadCloser, error) {
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
