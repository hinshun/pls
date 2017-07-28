package dind

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/hinshun/pls/pkg/failsafe"
	"github.com/hinshun/pls/pkg/namegen"
	"github.com/palantir/stacktrace"
)

const (
	StableDindImageName = "docker:stable-dind"
	DindPort            = 4444
)

type DindSpec struct {
	Name string
}

type Dind struct {
	client.APIClient

	ID       string
	Name     string
	HostAddr string
	HostPort string

	rootCTX context.Context
	rootCLI client.APIClient
}

func NewDind(ctx context.Context, cli client.APIClient, spec DindSpec) (*Dind, error) {
	dindName := spec.Name
	if dindName == "" {
		var err error
		dindName, err = namegen.GetUnusedContainerName(ctx, cli, "dind")
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to generate dind container name.")
		}
	}

	dindTCPPort := fmt.Sprintf("%d/tcp", DindPort)
	exposedPorts, err := dockercli.NewPortSet(dindTCPPort)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create dind port.")
	}

	cfg := &container.Config{
		Image: StableDindImageName,
		Labels: map[string]string{
			"pls": "dind",
		},
		Entrypoint:   []string{"sh"},
		ExposedPorts: exposedPorts,
		OpenStdin:    true,
	}
	hostCfg := &container.HostConfig{
		Privileged:      true,
		PublishAllPorts: true,
	}
	netCfg := &network.NetworkingConfig{}

	dindID, err := dockercli.ContainerCreate(ctx, cli, cfg, hostCfg, netCfg, dindName)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create dind")
	}

	containerJSON, err := cli.ContainerInspect(ctx, dindID)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to inspect dind")
	}

	portBindings, ok := containerJSON.NetworkSettings.Ports[nat.Port(dindTCPPort)]
	if !ok || len(portBindings) == 0 {
		return nil, stacktrace.NewError("failed to get dind host port")
	}

	dind := &Dind{
		ID:       dindID,
		Name:     dindName,
		HostAddr: containerJSON.NetworkSettings.Gateway,
		HostPort: portBindings[0].HostPort,

		rootCTX: ctx,
		rootCLI: cli,
	}

	err = dind.startDaemon()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed start daemon for '%s'", dind.Name)
	}

	dind.APIClient, err = dind.newClient()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create new docker client for '%s'", dind.Name)
	}

	err = dind.Healthcheck()
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to healthcheck '%s' dind daemon after start.", dind.Name)
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
		return stacktrace.Propagate(err, "failed to ping '%s' daemon.", d.Name)
	}
	return nil
}

func (d *Dind) startDaemon() error {
	execResp, err := d.rootCLI.ContainerExecCreate(d.rootCTX, d.ID, types.ExecConfig{
		Cmd: []string{"dockerd", "-H", fmt.Sprintf("tcp://0.0.0.0:%d", DindPort)},
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to create dind daemon exec in container '%s'.", d.Name)
	}

	err = d.rootCLI.ContainerExecStart(d.rootCTX, execResp.ID, types.ExecStartCheck{})
	if err != nil {
		return stacktrace.Propagate(err, "failed to start dind daemon exec in container '%s'.", d.Name)
	}

	return nil
}

func (d *Dind) newClient() (client.APIClient, error) {
	dindEndpoint := fmt.Sprintf("tcp://%s:%s", d.HostAddr, d.HostPort)
	cli, err := client.NewClient(dindEndpoint, "", nil, nil)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create docker client for '%s' daemon.", d.Name)
	}

	return cli, nil
}

// func (d *Dind) StartMITMProxy() error {
// 	execResp, err := d.CLI.ContainerExecCreate(d.CTX, d.ID, types.ExecConfig{
// 		Detach: true,
// 		Cmd:    []string{"dockerd"},
// 	})
// 	if err != nil {
// 		return stacktrace.Propagate(err, "Failed to create dockerd exec in container '%s'.", d.Name)
// 	}

// 	err = d.CLI.ContainerExecStart(d.CTX, execResp.ID, types.ExecStartCheck{})
// 	if err != nil {
// 		return stacktrace.Propagate(err, "Failed to start dockerd exec in container '%s'.", d.Name)
// 	}
// }
