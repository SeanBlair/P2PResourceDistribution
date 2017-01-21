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
	"log"
	// "net"
	"net/rpc"
	"os"
	"strconv"
	// TODO
)

// Resource server type.
type RServer int

// Peer server type
type PeerServer int

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

var (
	numPeers int
	// myID int
	// peersFile string
	serverIpPort string
	sessionID int
	resource Resource
	err error
)
// Main workhorse method.
func main() {
	args := os.Args[1:]

	// Missing command line args.
	if len(args) != 4 {
		fmt.Println("Usage: go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]")
		return
	}

	numPeers, err = strconv.Atoi(args[0])
	if err != nil {
		log.Fatal("Something wrong with [numPeers] arg: ", err)
	}
	myID, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("Something wrong with [peerID] arg: ", err)
	}
	// peersFile := args[2]
	serverIpPort = args[3]

	if myID == 1 {
		initSession()
		GetResource()
	} else {
		listen()
	}

	// client, err := rpc.Dial("tcp", serverIpPort)


	// initArgs := Init{numPeers, ""}

	// err = client.Call("RServer.InitSession", initArgs, &sessionID)
	// if err != nil {
	// 	log.Fatal("RServer.InitSession:", err)
	// }
	// fmt.Println("Server responded with sessionID: ", sessionID)

	// resourceArgs := ResourceRequest{sessionID, ""}

	// err = client.Call("RServer.GetResource", resourceArgs, &resource)
	// if err != nil {
	// 	log.Fatal("RServer.InitSession:", err)
	// }
	// fmt.Println("Server responded with Resource: ", resource)
	// fmt.Println("The Resource string is: ", resource.Resource)
	// fmt.Println("The Resource PeerID is: ", resource.PeerID)	
	// fmt.Println("The Resource NumRemaining is: ", resource.NumRemaining)

	// TODO
}

// sets sessionID, or exits program if error
func initSession() {	
	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial error: ", err)
	}
	initArgs := Init{numPeers, ""}
	err = client.Call("RServer.InitSession", initArgs, &sessionID)
	if err != nil {
		log.Fatal("RServer.InitSession:", err)
	}
	fmt.Println("Server responded with sessionID: ", sessionID)
}

func listen() {
	// should register its rpc's
	// listen on given port for given myID
	// serve...
	fmt.Println("in listen() state....")
	// infinite loop
	for true {

	}
}

func GetResource() {
	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial (in GetResource) error: ", err)
	}
	resourceArgs := ResourceRequest{sessionID, ""}

	err = client.Call("RServer.GetResource", resourceArgs, &resource)
	if err != nil {
		log.Fatal("RServer.InitSession:", err)
	}
	fmt.Println("Server responded with Resource: ", resource)
	fmt.Println("The Resource string is: ", resource.Resource)
	fmt.Println("The Resource PeerID is: ", resource.PeerID)	
	fmt.Println("The Resource NumRemaining is: ", resource.NumRemaining)

	// handle Resource
	listen()

}