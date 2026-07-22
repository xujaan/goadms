package service

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"log/slog"
)

type ScanResult struct {
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Open    bool   `json:"open"`
	Error   string `json:"error,omitempty"`
}

type ScannerService struct {
	logger *slog.Logger
}

func NewScannerService(logger *slog.Logger) *ScannerService {
	return &ScannerService{logger: logger}
}

// ScanSubnet scans a subnet for open port 4370 (ZKTeco).
// subnet format: "192.168.1.0/24" or "192.168.1"
func (s *ScannerService) ScanSubnet(ctx context.Context, subnet string) ([]ScanResult, error) {
	// Normalize subnet.
	if !strings.Contains(subnet, "/") {
		// Assume /24
		subnet = subnet + "/24"
	}

	_, ipnet, err := net.ParseCIDR(subnet)
	if err != nil {
		// Try as IP base.
		ip := net.ParseIP(subnet)
		if ip == nil {
			return nil, fmt.Errorf("invalid subnet: %s", subnet)
		}
		_, ipnet, _ = net.ParseCIDR(ip.String() + "/24")
	}

	var ips []net.IP
	for ip := ipnet.IP.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ipCopy := make(net.IP, len(ip))
		copy(ipCopy, ip)
		ips = append(ips, ipCopy)
	}

	// Skip network and broadcast addresses.
	if len(ips) > 2 {
		ips = ips[1 : len(ips)-1]
	}

	var results []ScanResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Limit concurrency.
	sem := make(chan struct{}, 50)

	for _, ip := range ips {
		wg.Add(1)
		go func(ip net.IP) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			addr := net.JoinHostPort(ip.String(), "4370")
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			r := ScanResult{IP: ip.String(), Port: 4370}
			if err == nil {
				conn.Close()
				r.Open = true
			} else {
				r.Error = err.Error()
			}

			mu.Lock()
			results = append(results, r)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return results, nil
}

// DetectSingle scans a single IP on port 4370.
func (s *ScannerService) DetectSingle(ip string) ScanResult {
	addr := net.JoinHostPort(ip, "4370")
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	r := ScanResult{IP: ip, Port: 4370}
	if err == nil {
		conn.Close()
		r.Open = true
	} else {
		r.Error = err.Error()
	}
	return r
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
