package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	Hostnames []map[string][]string
	m         sync.Mutex
)

func main() {
	ctx := context.Background()
	var dnsServerIp string
	var networkMask string
	var threads int
	flag.StringVar(&dnsServerIp, "d", "", "Specify and local DNS server ip address. Example: 192.168.1.155")
	flag.StringVar(&networkMask, "n", "", "CIDR notation of a newtork to scan. Example: 192.168.255.255/24")
	flag.IntVar(&threads, "t", 1, "Number of threads")
	flag.Parse()

	if dnsServerIp == "" {
		fmt.Printf("Missing dns server argument, Example: -d 192.168.1.155\n\n")
		os.Exit(1)
	}
	if networkMask == "" {
		fmt.Printf("Missing network mask. Example: -n 192.168.255.255/24\n\n")
		os.Exit(1)
	}

	r := &net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * time.Duration(1000),
			}
			return d.DialContext(ctx, network, fmt.Sprintf("%s:53", dnsServerIp))
		},
	}

	_, ipnet, err := net.ParseCIDR(networkMask)
	if err != nil {
		fmt.Printf("ERROR parsing network mask: %v", err.Error())
	}

	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)
	var addresses []string
	// loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		addresses = append(addresses, ip.String())
	}

	ipAddressSlices := getIPSlices(threads, addresses)
	fmt.Printf("Number of slices: %d\n", len(ipAddressSlices))
	timeStart := time.Now()
	wg := new(sync.WaitGroup)
	for _, ipSlice := range ipAddressSlices {
		wg.Add(1)
		go getHostNames(ipSlice, r, wg, ctx)
	}
	wg.Wait()
	timeFinish := time.Now()
	hosts, _ := json.Marshal(Hostnames)
	fmt.Printf("\nTime: %v\nHostnames: %v", timeFinish.Sub(timeStart), string(hosts))
}

func getIPSlices(threads int, addresses []string) [][]string {
	var ipAddressSlices [][]string

	for i := 0; i < threads; i++ {
		min := i * len(addresses) / threads
		max := ((i + 1) * len(addresses)) / threads

		ipAddressSlices = append(ipAddressSlices, addresses[min:max])
	}
	return ipAddressSlices
}

func getHostNames(addresses []string, r *net.Resolver, wg *sync.WaitGroup, ctx context.Context) {
	defer wg.Done()
	for _, ipAddress := range addresses {
		names, err := r.LookupAddr(ctx, ipAddress)
		if err != nil && strings.Contains(err.Error(), "no such host") {
			continue
		}
		if err != nil {
			continue
		}
		if len(names) == 0 {
			continue
		}
		m.Lock()
		fmt.Printf("\nHost found: %v ip: %v", names, ipAddress)
		Hostnames = append(Hostnames, map[string][]string{ipAddress: names})
		m.Unlock()
	}
}
