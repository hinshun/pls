package dockercli

import (
	"os"

	"github.com/docker/cli/cli/command/formatter"
	"github.com/docker/docker/api/types"
)

func PrintContainers(containers []types.Container) error {
	containerCtx := formatter.Context{
		Output: os.Stdout,
		Format: formatter.NewContainerFormat(formatter.TableFormatKey, false, false),
		Trunc:  false,
	}
	return formatter.ContainerWrite(containerCtx, containers)
}
