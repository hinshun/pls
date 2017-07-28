package namegen

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/hinshun/pls/pkg/failsafe"
)

func GetUnusedContainerName(ctx context.Context, cli client.APIClient, prefix string) (string, error) {
	var (
		containerName string
		retryPolicy   = failsafe.NewRetryPolicy()
	)

	err := failsafe.New(retryPolicy).Run(ctx, func() error {
		containerName = fmt.Sprintf("%s-%s", prefix, GetRandomName())
		_, err := cli.ContainerInspect(ctx, containerName)
		if err != nil {
			if client.IsErrNotFound(err) {
				retryPolicy.Cancel()
				return nil
			}
			return err
		}
		return fmt.Errorf("container name '%s' already in use", containerName)
	})
	if err != nil {
		return containerName, err
	}

	return containerName, nil
}
