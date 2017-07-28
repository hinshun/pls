package dockercli

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/palantir/stacktrace"
)

func ContainerCreate(ctx context.Context, cli client.APIClient, cfg *container.Config, hostCfg *container.HostConfig, netCfg *network.NetworkingConfig, containerName string) (string, error) {
	_, _, err := cli.ImageInspectWithRaw(ctx, cfg.Image)
	if err != nil {
		if !client.IsErrNotFound(err) {
			return "", stacktrace.Propagate(err, "failed to inspect image '%s'.", cfg.Image)
		}

		imageStream, err := cli.ImagePull(ctx, cfg.Image, types.ImagePullOptions{})
		if err != nil {
			return "", stacktrace.Propagate(err, "failed to pull image '%s'.", cfg.Image)
		}

		loadResp, err := cli.ImageLoad(ctx, imageStream, true)
		if err != nil {
			return "", stacktrace.Propagate(err, "failed to load image '%s'.", cfg.Image)
		}
		defer loadResp.Body.Close()
	}

	createResp, err := cli.ContainerCreate(ctx, cfg, hostCfg, netCfg, containerName)
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to create container '%s'.", containerName)
	}

	err = cli.ContainerStart(ctx, createResp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", stacktrace.Propagate(err, "failed to start container '%s'.", containerName)
	}

	return createResp.ID, nil
}
