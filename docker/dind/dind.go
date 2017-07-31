package dind

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/hinshun/pls/docker/mitmproxy"
	"github.com/hinshun/pls/pkg/failsafe"
	"github.com/hinshun/pls/pkg/namegen"
	"github.com/palantir/stacktrace"
)

const (
	DindImageName              = "docker:stable-dind"
	DindPort                   = 2375
	DindPrefix                 = "dind"
	DockerSocketPath           = "/var/run/docker.sock"
	SystemCertificateDirectory = "/usr/local/share/ca-certificates"
)

type DindSpec struct {
	Name          string
	MITMProxyName string
}

type Dind struct {
	client.APIClient

	ID   string
	Name string

	rootCTX context.Context
	rootCLI client.APIClient
}

func New(ctx context.Context, cli client.APIClient, spec DindSpec) (*Dind, error) {
	dindName := spec.Name
	if dindName == "" {
		var err error
		dindName, err = namegen.GetUnusedContainerName(ctx, cli, DindPrefix)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate dind container name")
		}
	}

	var (
		dindCmd   []string
		mitmProxy *mitmproxy.MITMProxy
	)
	if spec.MITMProxyName != "" {
		var err error
		mitmProxy, err = mitmproxy.NewFromExisting(ctx, cli, spec.MITMProxyName)
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create reference from existing mitmproxy container")
		}

		dindCmd = append(dindCmd, "update-ca-certificates;", fmt.Sprintf("HTTPS_PROXY=%s:%d", mitmProxy.Name, mitmproxy.MITMProxyPort))
	}
	dindCmd = append(dindCmd, "dockerd", "-H", fmt.Sprintf("unix://%s", DockerSocketPath), "-H", fmt.Sprintf("tcp://0.0.0.0:%d", DindPort))

	dindTCPPort := fmt.Sprintf("%d/tcp", DindPort)
	exposedPorts, err := dockercli.NewPortSet(dindTCPPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create dind port")
	}

	cfg := &container.Config{
		Image: DindImageName,
		Labels: map[string]string{
			"pls": DindPrefix,
		},
		Entrypoint:   []string{"sh"},
		Cmd:          append([]string{"-c"}, strings.Join(dindCmd, " ")),
		ExposedPorts: exposedPorts,
	}
	hostCfg := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
	}
	netCfg := &network.NetworkingConfig{}

	createResp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, dindName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create dind container")
	}

	dind := &Dind{
		ID:   createResp.ID,
		Name: dindName,

		rootCTX: ctx,
		rootCLI: cli,
	}

	if spec.MITMProxyName != "" {
		err = cli.NetworkConnect(ctx, mitmProxy.Network, dind.ID, &network.EndpointSettings{})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to connect dind to mitmproxy network")
		}

		caStream, err := mitmProxy.GetCACertificate()
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to get mitmproxy ca certificate")
		}

		err = cli.CopyToContainer(ctx, dind.ID, SystemCertificateDirectory, caStream, types.CopyToContainerOptions{})
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to copy mitmproxy ca certificate to system certificate directory")
		}
	}

	err = dind.startDaemon()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to start daemon")
	}

	return dind, nil
}

func (d *Dind) Healthcheck() error {
	retryPolicy := failsafe.NewRetryPolicy().WithDelay(time.Second)
	err := failsafe.New(retryPolicy).Run(d.rootCTX, func() error {
		_, err := d.Ping(d.rootCTX)
		return err
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to ping dind daemon")
	}
	return nil
}

func (d *Dind) startDaemon() error {
	err := d.rootCLI.ContainerStart(d.rootCTX, d.ID, types.ContainerStartOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "failed to start dind")
	}

	d.APIClient, err = d.newClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create new docker client for dind")
	}

	err = d.Healthcheck()
	if err != nil {
		return stacktrace.Propagate(err, "failed to healthcheck daemon after start")
	}

	return nil
}

func (d *Dind) newClient() (client.APIClient, error) {
	containerJSON, err := d.rootCLI.ContainerInspect(d.rootCTX, d.ID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to inspect dind container")
	}

	hostBinding, err := dockercli.GetHostBinding(containerJSON, DindPort)
	if err != nil {
		return nil, stacktrace.NewError("failed to get dind host binding")
	}

	dindEndpoint := fmt.Sprintf("tcp://%s:%s", hostBinding.HostAddr, hostBinding.HostPort)
	cli, err := client.NewClient(dindEndpoint, "", nil, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create docker client for '%s' daemon", d.Name)
	}

	return cli, nil
}
