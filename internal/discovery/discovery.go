package discovery

import (
	"context"

	"github.com/grandcat/zeroconf"
)

// Service represents the discovery component.
type Service interface {
	// Register announces the service on the local network.
	// It returns a function to shutdown/unregister the service.
	Register(instanceName string, port int) (func(), error)

	// ResolveTarget finds the IP and Port of a target instance on the network.
	ResolveTarget(ctx context.Context, instanceName string) (ip string, port int, err error)
}

// zeroconfService is the mDNS implementation of discovery.Service.
type zeroconfService struct {
	serviceType string
	domain      string
}

// New creates a new mDNS discovery service.
func New() Service {
	return &zeroconfService{
		serviceType: "_lan-notifier._tcp",
		domain:      "local.",
	}
}

func (s *zeroconfService) Register(instanceName string, port int) (func(), error) {
	server, err := zeroconf.Register(instanceName, s.serviceType, s.domain, port, []string{"txtv=0", "lo=1", "la=2"}, nil)
	if err != nil {
		return nil, err
	}

	return server.Shutdown, nil
}

func (s *zeroconfService) ResolveTarget(ctx context.Context, instanceName string) (string, int, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return "", 0, err
	}

	entries := make(chan *zeroconf.ServiceEntry)
	var resolvedIP string
	var resolvedPort int

	// The context can be cancelled early if we find the target
	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if entry.Instance == instanceName {
				if len(entry.AddrIPv4) > 0 {
					resolvedIP = entry.AddrIPv4[0].String()
					resolvedPort = entry.Port
					cancel() // Stop searching
				}
			}
		}
	}(entries)

	err = resolver.Browse(innerCtx, s.serviceType, s.domain, entries)
	if err != nil {
		return "", 0, err
	}

	<-innerCtx.Done()

	// If the original context timed out and we didn't find an IP, return an error
	if resolvedIP == "" {
		if ctx.Err() == context.DeadlineExceeded {
			return "", 0, context.DeadlineExceeded
		}
		return "", 0, context.Canceled
	}

	return resolvedIP, resolvedPort, nil
}
