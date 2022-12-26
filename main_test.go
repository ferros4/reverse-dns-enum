package main

import (
	"context"
	"net"
	"sync"
	"testing"
)

//////////////////////////////////////////////
////     Unit tests created by ChatGPT   /////
//////////////////////////////////////////////

func TestGetIPSlices(t *testing.T) {
	// Test with a single thread
	threads := 1
	addresses := []string{"1.1.1.1", "2.2.2.2", "3.3.3.3", "4.4.4.4"}
	ipAddressSlices := getIPSlices(threads, addresses)

	// Check that the result is a slice with a single element
	if len(ipAddressSlices) != 1 {
		t.Errorf("Expected 1 slice, got %v", len(ipAddressSlices))
	}

	// Check that the single slice contains all of the addresses
	if len(ipAddressSlices[0]) != len(addresses) {
		t.Errorf("Expected slice to contain %v addresses, got %v", len(addresses), len(ipAddressSlices[0]))
	}

	// Test with multiple threads
	threads = 2
	ipAddressSlices = getIPSlices(threads, addresses)

	// Check that the result is a slice with two elements
	if len(ipAddressSlices) != 2 {
		t.Errorf("Expected 2 slices, got %v", len(ipAddressSlices))
	}

	// Check that the first slice contains half of the addresses
	if len(ipAddressSlices[0]) != len(addresses)/2 {
		t.Errorf("Expected first slice to contain %v addresses, got %v", len(addresses)/2, len(ipAddressSlices[0]))
	}

	// Check that the second slice contains the remaining addresses
	if len(ipAddressSlices[1]) != len(addresses)/2 {
		t.Errorf("Expected second slice to contain %v addresses, got %v", len(addresses)/2, len(ipAddressSlices[1]))
	}
}

func TestGetAddresses(t *testing.T) {
	// Test with a range of 1 address
	start := uint32(0x7f000001)
	finish := uint32(0x7f000001)
	addresses := getaddresses(start, finish)

	// Check that the result is a slice with a single element
	if len(addresses) != 1 {
		t.Errorf("Expected 1 address, got %v", len(addresses))
	}

	// Check that the single address is correct
	if addresses[0] != "127.0.0.1" {
		t.Errorf("Expected address to be 127.0.0.1, got %v", addresses[0])
	}

	// Test with a range of 2 addresses
	start = uint32(0x7f000001)
	finish = uint32(0x7f000002)
	addresses = getaddresses(start, finish)

	// Check that the result is a slice with two elements
	if len(addresses) != 2 {
		t.Errorf("Expected 2 addresses, got %v", len(addresses))
	}

	// Check that the first address is correct
	if addresses[0] != "127.0.0.1" {
		t.Errorf("Expected first address to be 127.0.0.1, got %v", addresses[0])
	}

	// Check that the second address is correct
	if addresses[1] != "127.0.0.2" {
		t.Errorf("Expected second address to be 127.0.0.2, got %v", addresses[1])
	}
}

func TestParseCIDRNotation(t *testing.T) {
	// Test with a valid CIDR notation
	networkMask := "192.168.1.0/24"
	start, finish := parseCIDRNotation(networkMask)
	if start != 3232235776 {
		t.Errorf("Expected start to be 3232235776, got %v", start)
	}
	if finish != 3232236031 {
		t.Errorf("Expected finish to be 3232236031, got %v", finish)
	}

	// Test with an invalid CIDR notation
	networkMask = "192.168.1.0/33"
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected a panic, got nil")
		}
	}()
	a, b := parseCIDRNotation(networkMask)
	if a == 0 && b == 0 {
		t.Errorf("Expected a panic but got 0, 0")
	}
}

func TestGetHostNamesSuccess(t *testing.T) {
	// Test with a valid address and resolver
	addresses := []string{"8.8.8.8"}
	r := net.Resolver{}
	wg := sync.WaitGroup{}
	ctx := context.Background()
	wg.Add(1)
	getHostNames(addresses, &r, &wg, ctx)
	if len(Hostnames) != 1 {
		t.Errorf("Expected Hostnames to have 1 entry, got %v", len(Hostnames))
	}
	if len(Hostnames[0]["8.8.8.8"]) == 0 {
		t.Error("Expected Hostnames[0]['8.8.8.8'] to have a name, got 0")
	}
}

func TestGetHostNamesInvalid(t *testing.T) {
	// Test with an invalid address and resolver
	addresses := []string{"invalid"}
	r := net.Resolver{}
	wg := sync.WaitGroup{}
	ctx := context.Background()
	wg.Add(1)
	getHostNames(addresses, &r, &wg, ctx)
	if len(Hostnames) != 1 {
		t.Errorf("Expected Hostnames to have 1 entry, got %v", len(Hostnames))
	}
}

func TestGetHostNamesWgNil(t *testing.T) {
	addresses := []string{}
	r := net.Resolver{}
	ctx := context.Background()
	// Test with a nil WaitGroup
	getHostNames(addresses, &r, nil, ctx)
	if len(Hostnames) != 1 {
		t.Errorf("Expected Hostnames to have 1 entry, got %v", len(Hostnames))
	}
}

func TestDoDnsOutput(t *testing.T) {
	// Test valid DNS server
	var wg sync.WaitGroup
	wg.Add(1)
	go doDnsOutput("8.8.8.8", &wg)
	wg.Wait()
	if len(DNSServers) != 1 || DNSServers[0] != "8.8.8.8" {
		t.Errorf("Expected DNSServers to be [8.8.8.8], got %v", DNSServers)
	}

	// Test invalid DNS server
	DNSServers = []string{}
	wg.Add(1)
	go doDnsOutput("invalid", &wg)
	wg.Wait()
	if len(DNSServers) != 0 {
		t.Errorf("Expected DNSServers to be [], got %v", DNSServers)
	}
}

func TestGetDns(t *testing.T) {
	// Test valid DNS server addresses
	DNSServers = []string{}
	ipSlice := []string{"8.8.8.8", "1.1.1.1"}
	var wg sync.WaitGroup
	wg.Add(1)
	go getDns(ipSlice, &wg)
	wg.Wait()
	if len(DNSServers) != 2 || !contains(DNSServers, "8.8.8.8") || !contains(DNSServers, "1.1.1.1") {
		t.Errorf("Expected DNSServers to be [8.8.8.8, 1.1.1.1], got %v", DNSServers)
	}

	// Test invalid DNS server addresses
	DNSServers = []string{}
	ipSlice = []string{"invalid", "1.1.1.1"}
	wg.Add(1)
	go getDns(ipSlice, &wg)
	wg.Wait()
	if len(DNSServers) != 1 || DNSServers[0] != "1.1.1.1" {
		t.Errorf("Expected DNSServers to be [1.1.1.1], got %v", DNSServers)
	}

	// Test empty DNS server addresses
	DNSServers = []string{}
	ipSlice = []string{}
	wg.Add(1)
	go getDns(ipSlice, &wg)
	wg.Wait()
	if len(DNSServers) != 0 {
		t.Errorf("Expected DNSServers to be [], got %v", DNSServers)
	}
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestFindDns(t *testing.T) {
	// Test valid DNS server addresses
	DNSServers = []string{}
	ipAddressSlices := [][]string{{"8.8.8.8", "1.1.1.1"}}
	findDns(ipAddressSlices)
	if len(DNSServers) != 2 || !contains(DNSServers, "8.8.8.8") || !contains(DNSServers, "1.1.1.1") {
		t.Errorf("Expected DNSServers to be [8.8.8.8, 1.1.1.1], got %v", DNSServers)
	}

	// Test invalid DNS server addresses
	DNSServers = []string{}
	ipAddressSlices = [][]string{{"invalid", "1.1.1.1"}}
	findDns(ipAddressSlices)
	if len(DNSServers) != 1 || DNSServers[0] != "1.1.1.1" {
		t.Errorf("Expected DNSServers to be [1.1.1.1], got %v", DNSServers)
	}

	// Test empty DNS server addresses
	DNSServers = []string{}
	ipAddressSlices = [][]string{}
	findDns(ipAddressSlices)
	if len(DNSServers) != 0 {
		t.Errorf("Expected DNSServers to be [], got %v", DNSServers)
	}
}

func TestDoOutput(t *testing.T) {
	// Test valid DNS server and IP addresses
	Hostnames = []map[string][]string{}
	ipAddressSlices := [][]string{{"8.8.8.8", "1.1.1.1"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	doOutput(ipAddressSlices, "8.8.8.8", ctx)
	if len(Hostnames) != 2 || !containsString(Hostnames, "dns.google") || !containsString(Hostnames, "one.one.one.one") {
		t.Errorf("Expected Hostnames to be [dns.google, one.one.one.one], got %v", Hostnames)
	}

	// Test invalid DNS server and valid IP addresses
	Hostnames = []map[string][]string{}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	doOutput(ipAddressSlices, "invalid", ctx)
	if len(Hostnames) != 0 {
		t.Errorf("Expected Hostnames to be [], got %v", Hostnames)
	}

	// Test valid DNS server and invalid IP addresses
	Hostnames = []map[string][]string{}
	ipAddressSlices = [][]string{{"invalid", "1.1.1.1"}}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	doOutput(ipAddressSlices, "8.8.8.8", ctx)
	if len(Hostnames) != 1 || !containsString(Hostnames, "one.one.one.one") {
		t.Errorf("Expected Hostnames to be [one.one.one.one], got %v", Hostnames)
	}

	// Test empty IP addresses
	Hostnames = []map[string][]string{}
	ipAddressSlices = [][]string{}
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	doOutput(ipAddressSlices, "8.8.8.8", ctx)
	if len(Hostnames) != 0 {
		t.Errorf("Expected Hostnames to be [], got %v", Hostnames)
	}
}

func containsString(slice []map[string][]string, elem string) bool {
	for _, item := range slice {
		for _, strSlice := range item {
			for _, str := range strSlice {
				if str == elem {
					return true
				}
			}
		}
	}
	return false
}
