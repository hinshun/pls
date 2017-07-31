package command

import (
	"context"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/mitmproxy"
	"github.com/palantir/stacktrace"

	"gopkg.in/urfave/cli.v2"
)

func PruneMITMProxies(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	pruneFilter := filters.NewArgs()
	pruneFilter.Add("label", fmt.Sprintf("pls=%s", mitmproxy.MITMProxyPrefix))
	mitmProxyContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Quiet:   true,
		All:     true,
		Filters: pruneFilter,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to list mitmproxy containers: %s", err)
	}

	if len(mitmProxyContainers) == 0 {
		logrus.Info("No mitmproxy containers were found")
	} else {
		for _, mitmProxyContainer := range mitmProxyContainers {
			err = cli.ContainerRemove(ctx, mitmProxyContainer.ID, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				return stacktrace.Propagate(err, "failed to remove mitmproxy container '%s': %s", mitmProxyContainer.ID, err)
			}

			logrus.Infof("Deleted container '%s'", mitmProxyContainer.Names)
		}
	}

	networkReport, err := cli.NetworksPrune(ctx, pruneFilter)
	if err != nil {
		return stacktrace.Propagate(err, "failed to prune mitmproxy networks")
	}

	for _, network := range networkReport.NetworksDeleted {
		logrus.Infof("Deleted network '%s'", network)
	}

	return nil
}
