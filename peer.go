/*
Implements the solution to assignment 2 for UBC CS 416 2016 W2.

Usage:
$ go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]

Example:
$ go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]

*/

package main

import (
	"bufio"
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

type NextResourceRequest struct {
	SessID int
}

// The Host rpc method
// Prints arg to console
func (t *PeerServer) Host(arg *HostRequest, success *bool) error {
	fmt.Println("Received this string to Host: ", arg.TheString)
	*success = true
	return nil
}

// TODO the Exit RPC

// TODO the GetNextResource RPC

func (t *PeerServer) GetNextResource(arg *NextResourceRequest, success *bool) error {
	fmt.Println("Received this SessionID: ", arg.SessID, " to GetNextResource with")
	sessionID = arg.SessID
	getResource()
	*success = true
	return nil
}

var (
	numPeers int
	myID     int
	peersFile string
	serverIpPort string
	sessionID    int
	resource     Resource
	peerAddresses []string
	err          error
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
	peersFile = args[2]
	serverIpPort = args[3]

	setPeerAddresses()

	if myID == 1 {
		initSession()
		getResource()
	} else {
		listen()
	}

}

func setPeerAddresses() {
	file, err := os.Open(peersFile)
    if err != nil {
        log.Fatal("Error while opening [peersFile]: ", err)
    }
    defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        peerAddresses = append(peerAddresses, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

    fmt.Println("The full contents of the peerAddresses slice is: ", peerAddresses)
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

	myAddress := peerAddresses[myID -1]

	listener, err := net.Listen("tcp", myAddress)
	if err != nil {
		log.Fatal("Error in net.Listen() in listen(): ", err)
	}

	// rpc.Accept(listener)

	for {
    conn, err := listener.Accept()
    if err != nil {
      fmt.Println("Error in listener.Accept() in listen: ", err)
    }

    go rpc.ServeConn(conn)
   }
    fmt.Println("at end of listen() method")
  
}

func getResource() {
	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial (in getResource) error: ", err)
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

	//TODO	
	// if last, call Exit to all but myself, then exit
	// else delegate (call GetResource to next peer in line) and listen()

	// if for myID, print to console, else call Host to appropriate address
	if resource.PeerID == myID {
		fmt.Println(resource.Resource)
	} else {
		hostResource(resource.PeerID, resource.Resource)
	}

	if resource.NumRemaining > 0 {
		nextPeerAddress := getNextPeerAddress()
		go getNextResource(nextPeerAddress)
		listen()
	} else {
		exitProgram()
	}

	fmt.Println("Bye, bye!!! :)")
}

// call ExitProgram RPC for all peers exept for me
// then exits program (me)
func exitProgram() {
	fmt.Println("in exitProgram()..... (not implemented) ")
}

// returns peerAddress that follows the address corresponding to 
// myID in peersFile
func getNextPeerAddress() (string) {
	if myID == len(peerAddresses) {
		return peerAddresses[0]
	} else {
		return peerAddresses[myID]
	}
}

// calls the PeerServer.GetNextAddress RPC with given IP:Port
func getNextResource(peerAddress string) {
	client, err := rpc.Dial("tcp", peerAddress)
	if err != nil {
		log.Fatal("rpc.Dial(tcp, ", peerAddress, ") failed in getNextResource(): ", err)
	}

	nextResourceRequest := NextResourceRequest{sessionID}
	var successful bool
	err = client.Call("PeerServer.GetNextResource", &nextResourceRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.GetNextResource ...) failed in getNextResource() for peerAddress: ", peerAddress)
	}

	fmt.Println("Was my peer RPC (GetNextResource) successful?? Answer: ", successful)

	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in getNextResource", err)
	}
}

// calls the PeerServer.Host RPC to the given peerID

func hostResource(peer int, resourceString string) {
	// peerAddresses represents a zero indexed array
	address := peerAddresses[peer -1]

	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("rpc.Dial(tcp, ", address, ") failed in hostResource(): ", err)
	}

	hostRequest := HostRequest{resourceString}
	var successful bool
	err = client.Call("PeerServer.Host", &hostRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.Host ...) failed in hostRequest() for peerID: ", peer, " and resource: ", resourceString)
	}
	fmt.Println("Was my peer RPC (Host) successful?? Answer: ", successful)

// TODO: close connection???	
	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in hostResource", err)
	}
}
