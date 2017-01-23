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

type ExitRequest struct {
	Request bool
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
	isListen bool
	isExit bool
)

// The Host rpc method
// Prints arg to console
func (t *PeerServer) Host(arg *HostRequest, success *bool) error {
	fmt.Println(arg.TheString)
	*success = true
	return nil
}

// TODO the Exit RPC

func (t *PeerServer) ExitProgram(arg *ExitRequest, success *bool) error {
	isExit = true
	*success = true
	return nil
}

// TODO the GetNextResource RPC

func (t *PeerServer) GetNextResource(arg *NextResourceRequest, success *bool) error {
	// fmt.Println("Received this SessionID: ", arg.SessID, " to GetNextResource with")
	sessionID = arg.SessID
	// getResource()
	isListen = false
	*success = true
	return nil
}



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

	isExit = false

	if myID == 1 {
		// to satisfy the minumum time before all parts of the system are running
		time.Sleep(2 * time.Second)
		// set sessionID
		initSession()
		// getResource()
		isListen = false
	} else {
		isListen = true
	}

	// main logic infinite loop
	for true {
		if isExit {
			os.Exit(0)
		} else {	
			if isListen {
				listenState()
			}else {
				getResource()
			}
		}
	}

	// TODO should not need this...
	fmt.Println("Exiting main because reached the end...")
	return
}

func setPeerAddresses() {
	file, err := os.Open(peersFile)
    if err != nil {
        log.Fatal("Error while opening [peersFile]: ", err)
    }
    // defer file.Close()

    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        peerAddresses = append(peerAddresses, scanner.Text())
    }
    if err := scanner.Err(); err != nil {
        log.Fatal(err)
    }

	file.Close()    
    return

    // fmt.Println("The full contents of the peerAddresses slice is: ", peerAddresses)
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
	// fmt.Println("Server responded with sessionID: ", sessionID)
	// TODO close??
	return
}

func listenState() {
	// listen on given port for given myID
	// fmt.Println("in listenState() state....")

	peerServer := new(PeerServer)
	rpc.Register(peerServer)

	myAddress := peerAddresses[myID - 1]

	listener, err := net.Listen("tcp", myAddress)
	if err != nil {
		log.Fatal("Error in net.Listen() in listenState(): ", err)
	}

	// blocks until a request arrives
	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error in listener.Accept() in listenState(): ", err)
	}
	rpc.ServeConn(conn)

    // fmt.Println("at end of listenState() method")

    // TODO: should the connection be closed??
    conn.Close()

    // TODO: should listener be closed??
    // appears to be required with my implementation...
    listener.Close()
    return
}

func getResource() {

	client, err := rpc.Dial("tcp", serverIpPort)
	if err != nil {
		log.Fatal("rpc.Dial (in getResource) error: ", err)
	}
	resourceArgs := ResourceRequest{sessionID, ""}

	err = client.Call("RServer.GetResource", resourceArgs, &resource)
	if err != nil {
		log.Fatal("Error in call to client.Call(RServer.GetResource, ", resourceArgs, "...)", err)
	}

	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in getResource", err)
	}

	// handle Resource

	//TODO	
	// if last, call Exit to all but myself, then exit
	// else delegate (call GetResource to next peer in line) and listen()

	// store numRemaining before calling hostResource... (bug)
	isRemaining := resource.NumRemaining > 0

	// if for myID, print to console, else call Host to appropriate address
	if resource.PeerID == myID {
		fmt.Println(resource.Resource)
	} else {
		hostResource(resource.PeerID, resource.Resource)
		// give peer some time to host the resource in order to be
		// be in listenState in time for next request.
		time.Sleep(time.Second)
	}

	// nil Resource to avoid weird bug were numRemaing would not be set to 
	// zero if already non-zero...
	resource = Resource{"",0,0}

	if isRemaining {
		nextPeerAddress := getNextPeerAddress()
		getNextResource(nextPeerAddress)
		// listen()
		isListen = true
	} else {
		exitAllPeers()
		log.Fatal("Exiting program after call to exitAllPeers()... Wut!!")
	}
	return
}

// call ExitProgram RPC for all peers exept for me
// then exits program (me)
func exitAllPeers() {
	for i, address := range peerAddresses {
		if i + 1 != myID {
			exitProgram(address)
		}
	}
	os.Exit(0)
}

func exitProgram(ipPort string) {
	client, err := rpc.Dial("tcp", ipPort)
	if err != nil {
		log.Fatal("rpc.Dial(tcp, ", ipPort, ") failed in exitProgram(): ", err)
	}

	exitRequest := ExitRequest{true}
	var successful bool
	err = client.Call("PeerServer.ExitProgram", &exitRequest, &successful)
	if err != nil {
		log.Fatal("client.Call(PeerServer.ExitProgram ...) failed in exitProgram() for peerAddress: ", ipPort)
	}

	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in getNextResource", err)
	}
	return
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

	// fmt.Println("Was my peer RPC (GetNextResource) successful?? Answer: ", successful)

	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in getNextResource", err)
	}
	return
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
	// fmt.Println("Was my peer RPC (Host) successful?? Answer: ", successful)
	
	err = client.Close()
	if err != nil {
		log.Fatal("Error while closing tcp connection in hostResource", err)
	}
	return
}
