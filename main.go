package main

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

// Progress Bar code used from:
// https://www.pixelstech.net/article/1596946473-A-simple-example-on-implementing-progress-bar-in-GoLang

type Bar struct {
	percent int64  // progress percentage
	cur     int64  // current progress
	total   int64  // total value for progress
	rate    string // the actual progress bar to be printed
	graph   string // the fill value for progress bar
}

var (
	Hostnames         []map[string][]string
	DNSServers        []string
	m                 sync.Mutex
	bar               Bar
	addressesComplete int64
	numOfAddresses    int64
)

func main() {
	ctx := context.Background()
	dnsServerIp, networkMask, findDnsServers, threads := getCommandLineFlags()

	start, finish := parseCIDRNotation(networkMask)
	addresses := getaddresses(start, finish)
	numOfAddresses = int64(len(addresses))
	ipAddressSlices := getIPSlices(threads, addresses)
	go calculateBar()
	if findDnsServers {
		findDns(ipAddressSlices)
	} else {
		doOutput(ipAddressSlices, dnsServerIp, ctx)
	}

}

func calculateBar() {

	addressesComplete = 0
	bar.NewOption(0, numOfAddresses)

	for {
		if addressesComplete >= numOfAddresses {
			break
		}
		time.Sleep(100 * time.Millisecond)
		bar.Play(addressesComplete)
	}
}

func (bar *Bar) NewOption(start, total int64) {
	bar.cur = start
	bar.total = total
	if bar.graph == "" {
		bar.graph = "=>"
	}
	bar.percent = bar.getPercent()
	for i := 0; i < int(bar.percent); i += 2 {
		bar.rate += bar.graph // initial progress position
	}
}

func (bar *Bar) getPercent() int64 {
	return int64((float32(bar.cur) / float32(bar.total)) * 50)
}

func (bar *Bar) Play(cur int64) {
	bar.cur = cur
	last := bar.percent
	bar.percent = bar.getPercent()
	if bar.percent != last {
		var i int64 = 0
		for ; i < bar.percent-last; i++ {
			bar.rate += bar.graph
		}
		fmt.Printf("\r[%-50s]%3d%% %8d/%d", bar.rate, bar.percent*2, bar.cur, bar.total)
	}
}

func (bar *Bar) Finish() {
	fmt.Println()
}

func findDns(ipAddressSlices [][]string) {
	timeStart := time.Now()
	wg := new(sync.WaitGroup)
	for _, ipSlice := range ipAddressSlices {
		wg.Add(1)
		go getDns(ipSlice, wg)
	}
	wg.Wait()
	bar.Play(numOfAddresses)
	bar.Finish()
	timeFinish := time.Now()
	hosts, _ := json.Marshal(DNSServers)
	fmt.Printf("\nTime: %v\nDNS Servers: %v", timeFinish.Sub(timeStart), string(hosts))
}

func getDns(ipSlice []string, wg *sync.WaitGroup) {
	if wg != nil {
		defer wg.Done()
	}
	newWG := new(sync.WaitGroup)
	for _, address := range ipSlice {
		newWG.Add(1)
		go doDnsOutput(address, newWG)
	}
	newWG.Wait()
}

func doDnsOutput(address string, wg *sync.WaitGroup) {
	defer wg.Done()
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:53", address), time.Second*time.Duration(1))
	addressesComplete++
	if err != nil && conn == nil {
		return
	}
	_ = conn.Close()
	//fmt.Printf("Found DNS server at:%v\n", address)
	m.Lock()
	DNSServers = append(DNSServers, fmt.Sprintf("%v", address))
	m.Unlock()
}

func doOutput(ipAddressSlices [][]string, dnsServerIp string, ctx context.Context) {
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
	timeStart := time.Now()
	wg := new(sync.WaitGroup)
	for _, ipSlice := range ipAddressSlices {
		wg.Add(1)
		go getHostNames(ipSlice, r, wg, ctx)
	}
	wg.Wait()
	bar.Play(numOfAddresses)
	bar.Finish()
	timeFinish := time.Now()
	hosts, _ := json.Marshal(Hostnames)
	fmt.Printf("\nTime: %v\nHostnames: %v", timeFinish.Sub(timeStart), string(hosts))
}

func parseCIDRNotation(networkMask string) (uint32, uint32) {
	_, ipnet, err := net.ParseCIDR(networkMask)
	if err != nil {
		panic(fmt.Sprintf("Error: failed to parse network mask: %v", err.Error()))
	}

	// convert IPNet struct mask and address to uint32
	// network is BigEndian
	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipnet.IP)

	// find the final address
	finish := (start & mask) | (mask ^ 0xffffffff)
	return start, finish
}

func getCommandLineFlags() (string, string, bool, int) {
	var dnsServerIp string
	var networkMask string
	var threads int
	var findDNSServers bool
	flag.StringVar(&networkMask, "n", "", "CIDR notation of a newtork to scan. Example: 192.168.255.255/24")
	flag.BoolVar(&findDNSServers, "f", false, "If true search network for DNS servers and output")
	flag.StringVar(&dnsServerIp, "d", "", "Specify and local DNS server ip address. Example: 192.168.1.155")
	flag.IntVar(&threads, "t", 1, "Number of threads")
	flag.Parse()

	if networkMask == "" {
		fmt.Printf("Missing network mask. Example: -n 192.168.255.255/24\n")
		os.Exit(1)
	}

	if dnsServerIp == "" && findDNSServers == false {
		fmt.Printf("Missing dns server argument, Example: -d 192.168.1.155\n")
		os.Exit(1)
	}

	if reflect.TypeOf(threads).Kind().String() != "int" || threads < 1 {
		fmt.Printf("Number of threads must be a positive integer\n")
		os.Exit(1)
	}
	return dnsServerIp, networkMask, findDNSServers, threads
}

func getaddresses(start uint32, finish uint32) []string {
	var addresses []string
	// loop through addresses as uint32
	for i := start; i <= finish; i++ {
		// convert back to net.IP
		ip := make(net.IP, 4)
		binary.BigEndian.PutUint32(ip, i)
		addresses = append(addresses, ip.String())
	}
	return addresses
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
	if wg != nil {
		defer wg.Done()
	}
	for _, ipAddress := range addresses {
		names, err := r.LookupAddr(ctx, ipAddress)
		addressesComplete++
		if err != nil || len(names) == 0 {
			continue
		}
		m.Lock()
		//fmt.Printf("\nHost found: %v ip: %v", names, ipAddress)
		addressesComplete++
		Hostnames = append(Hostnames, map[string][]string{ipAddress: names})
		m.Unlock()
	}
}
