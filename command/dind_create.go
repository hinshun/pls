package command

import (
	"context"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dind"
	"github.com/palantir/stacktrace"

	"gopkg.in/urfave/cli.v2"
)

func CreateDind(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	spec := dind.DindSpec{}

	dind, err := dind.NewDind(ctx, cli, spec)
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
