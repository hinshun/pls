package command

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dind"
	"github.com/palantir/stacktrace"
	"github.com/sirupsen/logrus"

	"gopkg.in/urfave/cli.v2"
)

func PruneDinds(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	pruneFilter := filters.NewArgs()
	pruneFilter.Add("label", fmt.Sprintf("pls=%s", dind.DindPrefix))
	dindContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Quiet:   true,
		All:     true,
		Filters: pruneFilter,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to list dind containers: %s", err)
	}

	if len(dindContainers) == 0 {
		logrus.Info("No dind containers were found")
	} else {
		for _, dindContainer := range dindContainers {
			err = cli.ContainerRemove(ctx, dindContainer.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				return stacktrace.Propagate(err, "failed to remove dind container '%s': %s", dindContainer.ID, err)
			}

			logrus.Infof("Deleted container '%s'", dindContainer.Names)
		}
	}

	return nil
}
