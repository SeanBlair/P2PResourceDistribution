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
	"net"
	"net/rpc"
	"os"
	"strconv"
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

// Peer rpc stuffs

// Peer server type
type PeerServer int

// The string that peer must host
type HostRequest struct {
	TheString string
}

// The Host rpc method
// Prints arg to console
func (t *PeerServer) Host(arg *HostRequest, success *bool) (error) {
	fmt.Println("Received this string to Host: ", arg.TheString)
	*success = true
	return nil
}

// TODO the Exit RPC

// TODO the GetNextResource RPC

var (
	numPeers int
	myID int
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
	myID, err = strconv.Atoi(args[1])
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

}

// sets sessionID, or exits program if error
func initSession() {
	// TODO  add sleep for 2 seconds!	
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
	// TODO close??
}

func listen() {
	// listen on given port for given myID
	fmt.Println("in listen() state....")

	peerServer := new(PeerServer)
	rpc.Register(peerServer)

	// TODO need to change this to determine what to listen
	// to depending on myID and peersFile
	listener, err := net.Listen("tcp", "localhost:2222")
  	if err != nil {
    	log.Fatal("Error in net.Listen() in listen(): ", err)
  	}

  	rpc.Accept(listener)

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

	// if for myID, print to console, else call Host to appropriate address
	// if last, call Exit to all but myself, then exit
	// else delegate (call GetResource to next peer in line) and listen()

	// trying peer rpc..
	client, err = rpc.Dial("tcp", "localhost:2222")
	if err != nil {
		log.Fatal("rpc.Dial(tcp, peer (localhost:2222) Error: ", err)
	}
	hostRequest := HostRequest{resource.Resource}
	var successful bool
	err = client.Call("PeerServer.Host", &hostRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.Host failed... : ", err)
	}
	fmt.Println("Was my peer RPC successful?? Answer: ", successful)
	fmt.Println("Bye, bye!")


	// listen()

}