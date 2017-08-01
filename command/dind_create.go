package command

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dind"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"

	"gopkg.in/urfave/cli.v2"
)

func CreateDind(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	err = dockercli.LazyImageLoad(ctx, cli, dind.DindImageName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to load dind image")
	}

	spec := dind.DindSpec{
		Name:                  c.String("name"),
		MITMProxyName:         c.String("mitm"),
		RegistryServerAddress: c.String("registry"),
		RegistryUsername:      c.String("username"),
		RegistryPassword:      c.String("password"),
	}

	dind, err := dind.New(ctx, cli, spec)
	if err != nil {
		return stacktrace.Propagate(err, "failed to create new dind")
	}

	logrus.Infof("Created dind container '%s'", dind.Name)

	dindContainers, err := dind.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return stacktrace.Propagate(err, "failed to get dind containers")
	}

	logrus.Infof("Dind containers: %s", dindContainers)

	return nil
}
