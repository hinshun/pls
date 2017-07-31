package dockercli

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/palantir/stacktrace"
)

func LazyImageLoad(ctx context.Context, cli client.APIClient, image string) error {
	_, _, err := cli.ImageInspectWithRaw(ctx, image)
	if err != nil {
		if !client.IsErrNotFound(err) {
			return stacktrace.Propagate(err, "failed to inspect image '%s'.", image)
		}

		// Dockerd doesn't have the image, so we pull it down.
		imageStream, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			return stacktrace.Propagate(err, "failed to pull image '%s'.", image)
		}

		// Load into target dockerd.
		loadResp, err := cli.ImageLoad(ctx, imageStream, true)
		if err != nil {
			return stacktrace.Propagate(err, "failed to load image '%s'.", image)
		}
		defer loadResp.Body.Close()
	}

	return nil
}
