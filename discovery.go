package srvdisco

import (
	"fmt"
	"net"
	"net/url"
)

// Discover does a DNS SRV lookup on the specified domain,
// for the specified service, and returns a string array of URLs to use
func Discover(domain, service, scheme string) (urls []string, err error) {

	errorChan := make(chan error)
	discoChan := make(chan url.URL)

	go DiscoverChan(domain, service, scheme, discoChan, errorChan)

	for {
		select {
		case err = <-errorChan:
			if err.Error() == "Complete" {
				err = nil
			}
			return
		case u := <-discoChan:
			urls = append(urls, u.String())
		}
	}

	return
}

// DiscoverChan does a DNS SRV lookup on the specified domain,
// for the specified service, and streams url.URLs to use via discoChan,
// and errors over errorChan, closing both when done. If errorChan
// receives any messages, that signals the end of streams. An error
// of "Complete" is a non-error case. (Yeah, needs reworking)
func DiscoverChan(domain, service, scheme string, discoChan chan url.URL, errorChan chan error) {
	defer close(errorChan)
	defer close(discoChan)

	_, addrs, err := net.LookupSRV(service, "tcp", domain)
	if err != nil {
		errorChan <- err
		return
	}

	for _, srv := range addrs {
		u := &url.URL{
			Scheme: scheme,
			Host:   net.JoinHostPort(srv.Target, fmt.Sprintf("%d", srv.Port)),
		}
		discoChan <- &u
	}

	errorChan <- fmt.Errorf("Complete")

}
