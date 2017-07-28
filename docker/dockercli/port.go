package dockercli

import (
	"strings"

	"github.com/docker/go-connections/nat"
	"github.com/palantir/stacktrace"
)

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
