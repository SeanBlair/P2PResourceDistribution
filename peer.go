/*
Implements the solution to assignment 2 for UBC CS 416 2016 W2.

Usage:
$ go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]

Example:
$ go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]

*/

package main

import (
	"fmt"
	"net"
	"net/rpc"
	"os"
	// TODO
)

// Resource server type.
type RServer int

// Request that peer sends in call to RServer.InitSession
// Return value is int
type Init struct {
	NumPeers int
	IPaddr   string // Must be set to ""
}

// Request that peer sends in call to RServer.GetResource
type ResourceRequest struct {
	SessionID int
	IPaddr    string // Must be set to ""
}

// Response that the server returns from RServer.GetResource
type Resource struct {
	Resource     string
	PeerID       int
	NumRemaining int
}

// Main workhorse method.
func main() {
	args := os.Args[1:]

	// Missing command line args.
	if len(args) != 4 {
		fmt.Println("Usage: go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]")
		return
	}

	// TODO
}