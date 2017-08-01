package command

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/hinshun/pls/docker/dockercli"
	"github.com/hinshun/pls/docker/mitmproxy"
	"github.com/palantir/stacktrace"

	"gopkg.in/urfave/cli.v2"
)

func ListMITMProxies(c *cli.Context) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		return stacktrace.Propagate(err, "failed to create docker client from env: %s", err)
	}

	listFilter := filters.NewArgs()
	listFilter.Add("label", fmt.Sprintf("pls=%s", mitmproxy.MITMProxyPrefix))
	mitmProxyContainers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		All:     true,
		Filters: listFilter,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to list mitmproxy containers")
	}

	err = dockercli.PrintContainers(mitmProxyContainers)
	if err != nil {
		return stacktrace.Propagate(err, "failed to write out mitmproxy  containers")
	}

	return nil
}
