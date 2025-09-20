package operator

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/hashicorp/mdns"
	"github.com/whytheplatypus/switchboard/config"
)

func Listen(ctx context.Context) <-chan *mdns.ServiceEntry {
	// Make a channel for results and start listening
	entries := make(chan *mdns.ServiceEntry, 5)

	// Start the lookup
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		defer close(entries)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				mdns.Lookup(fmt.Sprintf("%s", config.ServiceName), entries)
			}
		}
	}()

	return entries
}

func Connect(entry *mdns.ServiceEntry) error {
	if !strings.Contains(entry.Name, config.ServiceName) {
		return ErrUnknownEntry
	}
	return register(entry)
}

func register(entry *mdns.ServiceEntry) error {
	var addr string
	if v4 := entry.AddrV4; v4 != nil {
		addr = v4.String()
	} else if v6 := entry.AddrV6; v6 != nil {
		addr = "[" + v6.String() + "]"
	} else if host := entry.Host; host != "" {
		addr = host
	} else {
		return fmt.Errorf("no address found for entry %+v", entry)
	}
	port := 80
	if entry.Port != 0 {
		port = entry.Port
	}
	u, err := url.Parse(fmt.Sprintf("http://%s:%d", addr, port))
	if err != nil {
		return err
	}
	defaultRouter.register(entry.InfoFields[0], u)
	return nil
}
