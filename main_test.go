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
