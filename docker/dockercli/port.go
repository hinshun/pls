package dockercli

import (
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
)

type HostBinding struct {
	HostAddr string
	HostPort string
}

func GetHostBinding(containerJSON types.ContainerJSON, port int) (*HostBinding, error) {
	tcpPort := fmt.Sprintf("%d/tcp", port)
	portBindings, ok := containerJSON.NetworkSettings.Ports[nat.Port(tcpPort)]
	if !ok || len(portBindings) == 0 {
		return nil, stacktrace.NewError("failed to get host port")
	}

	return &HostBinding{
		HostAddr: containerJSON.NetworkSettings.Gateway,
		HostPort: portBindings[0].HostPort,
	}, nil
}

func NewPortSet(ports ...string) (nat.PortSet, error) {
	portSet := make(nat.PortSet)
	for _, port := range ports {
		parts := strings.Split(port, "/")
		if len(parts) != 2 {
			return nil, stacktrace.NewError("port must be of format [port]/[proto]")
		}

		natPort, err := nat.NewPort(parts[1], parts[0])
		if err != nil {
			return nil, stacktrace.Propagate(err, "failed to create nat.Port")
		}
		portSet[natPort] = struct{}{}
	}

	return portSet, nil
}
