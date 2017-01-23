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
	"time"
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

// Peer server type
type PeerServer int

// Request that peer sends in call to PeerServer.Host
type HostRequest struct {
	TheString string
}

// Request that peer sends in call to PeerServer.GetNextResource
type NextResourceRequest struct {
	SessID int
}

// Request that peer sends in call to PeerServer.ExitProgram
type ExitRequest struct {
	Request bool
}

var (
	// Number of Peers in the system
	numPeers int
	// Unique peerID (1 <= x <= numPeers)
	myID int
	// Path to file containing all peer ip_port. Line i corresponds to peer with peerID i
	peersFile string
	// Ip_port of functional server
	serverIpPort string
	// ID corresponding to current resource sharing session
	sessionID int
	// Resource that server provides as a result of RServer.GetResource
	resource Resource
	// Addresses in given peersFile
	peerAddresses []string
	err           error
	// true if in listen state, false otherwise
	isListen bool
	// true if in exit state, false otherwise
	isExit bool
)

// The PeerServer.Host RPC method
// Prints arg to console
func (t *PeerServer) Host(arg *HostRequest, success *bool) error {
	fmt.Println(arg.TheString)
	*success = true
	return nil
}

// The PeerServer.ExitProgram RPC method
// Triggers peer to exit immediately
func (t *PeerServer) ExitProgram(arg *ExitRequest, success *bool) error {
	isExit = true
	*success = true
	return nil
}

// The PeerServer.GetNextResource RPC method
// Makes peer call RServer.GetResource and decide what to do next
func (t *PeerServer) GetNextResource(arg *NextResourceRequest, success *bool) error {
	sessionID = arg.SessID
	isListen = false
	*success = true
	return nil
}

// Main workhorse method. (Program entry point)
func main() {
	args := os.Args[1:]
	// Missing command line args.
	if len(args) != 4 {
		fmt.Println("Usage: go run peer.go [numPeers] [peerID] [peersFile] [server ip:port]")
		return
	}
	numPeers, err = strconv.Atoi(args[0])
	if err != nil {
		log.Fatal(err)
	}
	myID, err = strconv.Atoi(args[1])
	if err != nil {
		log.Fatal(err)
	}
	peersFile = args[2]
	serverIpPort = args[3]
	setPeerAddresses()

	isExit = false

	// Peer with peerID == 1 starts the conversation with the server
	if myID == 1 {
		// To satisfy the minumum time before all parts of the system are running
		time.Sleep(2 * time.Second)
		// Set sessionID
		initSession()
		isListen = false
	} else {
		// All other peer are in listen (PeerServer) state
		isListen = true
	}

	// Main logic infinite loop
	for true {
		if isExit {
			os.Exit(0)
		} else {
			if isListen {
				listenState()
			} else {
				getResource()
			}
		}
	}
}

// Opens peersFile and extracts each address into peerAddresses []string
func setPeerAddresses() {
	file, err := os.Open(peersFile)
	if err != nil {
		log.Fatal("Error while opening [peersFile]: ", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerAddresses = append(peerAddresses, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	file.Close()
	return
}

// Sets sessionID by calling RServer.InitSesssion
func initSession() {
	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial() error in initSession(): ", err)
	}
	initArgs := Init{numPeers, ""}
	err = client.Call("RServer.InitSession", initArgs, &sessionID)
	if err != nil {
		log.Fatal("client.Call(RServer.InitSession) error: ", err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("client.Close() error in initSession(): ", err)
	}
	return
}

// Makes peer an RPC PeerServer listening on its corresponding
// ip:port in peersFile. Returns after serving exactly one RPC
func listenState() {
	peerServer := new(PeerServer)
	rpc.Register(peerServer)
	myAddress := peerAddresses[myID-1]

	listener, err := net.Listen("tcp", myAddress)
	if err != nil {
		log.Fatal("Error in net.Listen() in listenState(): ", err)
	}
	// Blocks until a request arrives
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal("Error in listener.Accept() in listenState(): ", err)
	}
	rpc.ServeConn(conn)
	conn.Close()
	listener.Close()
	return
}

// Calls RServer.GetResource and decides what to do next
func getResource() {
	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial error in getResource: ", err)
	}
	resourceArgs := ResourceRequest{sessionID, ""}
	err = client.Call("RServer.GetResource", resourceArgs, &resource)
	if err != nil {
		log.Fatal("client.Call(RServer.GetResource,..) error in getResource()", err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("client.Close() error in getResource: ", err)
	}
	// True if there are more resources at RServer
	isRemaining := resource.NumRemaining > 0
	// If for myID, print to console, else call PeerServer.Host to correct peer address
	if resource.PeerID == myID {
		fmt.Println(resource.Resource)
	} else {
		hostResource(resource.PeerID, resource.Resource)
		// Stall to give peer time to get into listenState for next request
		time.Sleep(time.Second)
	}
	// Nil Resource to avoid weird bug were numRemaing would not be set to
	// zero if already non-zero...
	resource = Resource{"", 0, 0}
	if isRemaining {
		// Delegate next call to RServer
		nextPeerAddress := getNextPeerAddress()
		getNextResource(nextPeerAddress)
		isListen = true
	} else {
		exitAllPeers()
	}
	return
}

// Call exitProgram for all peers except self, then exits
func exitAllPeers() {
	for i, address := range peerAddresses {
		if i+1 != myID {
			exitProgram(address)
		}
	}
	os.Exit(0)
}

// Call PeerServer.ExitProgram to peer listening on given ipPort
func exitProgram(ipPort string) {
	client, err := rpc.Dial("tcp", ipPort)
	if err != nil {
		log.Fatal("rpc.Dial() error in exitProgram(): ", err)
	}
	exitRequest := ExitRequest{true}
	var successful bool
	err = client.Call("PeerServer.ExitProgram", &exitRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.ExitProgram) error in exitProgram(): ", err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("client.Close() error in exitProgram(): ", err)
	}
	return
}

// Returns peerAddress that follows the address corresponding to
// myID in ascending order in peersFile
func getNextPeerAddress() string {
	if myID == len(peerAddresses) {
		return peerAddresses[0]
	} else {
		return peerAddresses[myID]
	}
}

// Calls PeerServer.GetNextAddress with given peerAddress
func getNextResource(peerAddress string) {
	client, err := rpc.Dial("tcp", peerAddress)
	if err != nil {
		log.Fatal("rpc.Dial() error in getNextResource(): ", err)
	}
	nextResourceRequest := NextResourceRequest{sessionID}
	var successful bool
	err = client.Call("PeerServer.GetNextResource", &nextResourceRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.GetNextResource) error in getNextResource(): ", err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("client.Close() error in getNextResource: ", err)
	}
	return
}

// Calls PeerServer.Host to peer with given peerID
func hostResource(peer int, resourceString string) {
	address := peerAddresses[peer-1]
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		log.Fatal("rpc.Dial() error in hostResource(): ", err)
	}
	hostRequest := HostRequest{resourceString}
	var successful bool
	err = client.Call("PeerServer.Host", &hostRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.Host) error in hostRequest(): ", err)
	}
	err = client.Close()
	if err != nil {
		log.Fatal("client.Close() error in hostResource", err)
	}
	return
}
