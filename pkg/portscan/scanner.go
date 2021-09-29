package portscan

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

type PortScanner struct {
	Ips   []net.IP
	Ports []int
	Lock  *semaphore.Weighted
}

func New(ips []net.IP, ports []int) *PortScanner {
	return &PortScanner{
		Ips:   ips,
		Ports: ports,
		//maybe change this later...
		Lock: semaphore.NewWeighted(256),
	}
}

func ScanPort(target string, timeout time.Duration) bool {
	conn, err := net.DialTimeout("tcp", target, timeout)
	// i don't like this part...if there's a better way we should do it...

	if err != nil {
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			return ScanPort(target, timeout)
		} else {
			return false
		}
	}

	conn.Close()
	return true
}

func (ps *PortScanner) Start(timeout time.Duration) chan string {
	results := make(chan string, 256)

	go func() {
		wg := sync.WaitGroup{}

		defer func() {
			wg.Wait()
			close(results)
		}()

		for _, ip := range ps.Ips {
			for _, port := range ps.Ports {
				ps.Lock.Acquire(context.TODO(), 1) //nolint: errcheck

				wg.Add(1)

                addr := fmt.Sprintf("%s:%d", ip.String(), port)
				go func(addr string) {
					defer ps.Lock.Release(1)

					defer wg.Done()

					if ScanPort(addr, timeout) {
						results <- addr
                    }
				}(addr)
			}
		}
	}()

	return results
}