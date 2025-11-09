package main

import (
	//"fmt"
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "RA/grpc"
)

type Node struct {
	pb.UnimplementedRAServer
	Portnumber      string
	Lamport         int32 // sendRequest, receiveRequest (?), access Critical section
	Queue           []*pb.Request
	State           State                  // RELEASE, WANTED, & HELD
	SystemNodes     map[string]pb.RAClient // portnumber and proto Ricart Agrawala client
	PermissionCount int32
}

type State string

const (
	Release State = "RELEASE"
	Wanted  State = "WANTED"
	Held    State = "HELD"
)

// for deferring responding to requests
var mu sync.Mutex
var cond = sync.NewCond(&mu)

// For logging to system.log and to the terminal
var fileLog *log.Logger
var termLog *log.Logger

// prints that connection established to Critical Section
func (n *Node) CriticalSection() {
	fileLog.Printf("Node: %s made it to the Critical Section! MUAHAHAAAA \n", n.Portnumber)
	termLog.Printf("Node: %s made it to the Critical Section! MUAHAHAAAA \n", n.Portnumber)
	mu.Lock()
	n.Lamport++
	mu.Unlock()
	time.Sleep(4 * time.Second)
}

// RequestCriticalSectionAccess is our rpc call
func (n *Node) RequestCriticalSectionAccess(ctx context.Context, req *pb.Request) (*pb.Empty, error) {
	fileLog.Printf("Node: %s received request from Node: %s", n.Portnumber, req.Portnr)
	termLog.Printf("I have received a request from Node: %s", req.Portnr)

	myPortnr, _ := strconv.Atoi(n.Portnumber)
	reqPortnr, _ := strconv.Atoi(req.GetPortnr())
	// append to queue if we are in the critical section, or we want to and we either have higher lamport priority or the same lamport with higher portnumber
	if n.State == Held || (n.State == Wanted && (n.Lamport < req.Lamport || (n.Lamport == req.Lamport && myPortnr > reqPortnr))) { //
		mu.Lock()
		n.Queue = append(n.Queue, req)
		mu.Unlock()
	} else {
		_, err := n.SystemNodes[req.GetPortnr()].SendPermission(context.Background(), &pb.Response{Permission: true, Portnr: n.Portnumber})
		if err != nil {
			log.Printf("error trying to send request to node: %s  %v", req.GetPortnr(), err)
		}
	}
	return &pb.Empty{}, nil
}

// sending a permission reson
func (n *Node) SendPermission(ctx context.Context, resp *pb.Response) (*pb.Empty, error) {
	fileLog.Printf("Node: %s has received permission from Node: %s", n.Portnumber, resp.GetPortnr())
	termLog.Printf("I have received a permission")

	mu.Lock()
	n.PermissionCount++
	termLog.Printf("Current permission count: %d, total permission count needed: %d", n.PermissionCount, len(n.SystemNodes))
	cond.Broadcast()

	mu.Unlock()

	return &pb.Empty{}, nil
}

// Called from main
// Check the node is RELEASE and neither WANTED nor HELD
func (n *Node) RequestAccess() {
	if n.State == Release {
		//n.State = Wanted
		n.SendRequests()
	}
}

// Uses a Wait Group to ask all other nodes permission to enter the CS concurrently
// Once the node has received permission it calls to enter the CS from here
func (n *Node) SendRequests() {
	mu.Lock()
	n.State = Wanted
	n.Lamport++ // increment Lamport timestamp for this request
	mu.Unlock()

	// Send requests to all other nodes asynchronously
	for nodeAddr := range n.SystemNodes {
		go func(nodeConnection string) {
			fileLog.Printf("Node: %s will ask Node: %s for permission", n.Portnumber, nodeConnection)
			termLog.Printf("I will request permission from Node: %s", nodeConnection)

			_, err := n.SystemNodes[nodeConnection].RequestCriticalSectionAccess(
				context.Background(),
				&pb.Request{
					Portnr:  n.Portnumber,
					Lamport: n.Lamport,
				},
			)
			if err != nil {
				log.Printf("error trying to send request to node %s: %v", nodeConnection, err)
			}
		}(nodeAddr)
	}

	// Wait for all replies using sync.Cond
	mu.Lock()
	for int(n.PermissionCount) < len(n.SystemNodes) {
		cond.Wait() // goroutine sleeps until cond.Broadcast() is called
	}
	n.State = Held
	mu.Unlock()

	// Critical section
	n.CriticalSection()
	fileLog.Printf("Node: %s has left the Critical Section", n.Portnumber)
	termLog.Println("I have left the Critical Section ")

	// Release and wake any deferred requests
	mu.Lock()
	n.PermissionCount = 0
	n.State = Release
	cond.Broadcast() // notify all waiting handlers
	mu.Unlock()

	for _, req := range n.Queue {
		_, err := n.SystemNodes[req.Portnr].SendPermission(context.Background(), &pb.Response{Permission: true, Portnr: n.Portnumber})
		if err != nil {
			log.Printf("Error sending deferred permission to %s: %v", req.Portnr, err)
		}
		n.Queue = n.Queue[1:]
	}
	//n.Queue = n.Queue
}

// Make a Client connection and save it to the SystemNodes map
func (n *Node) JoinSystem() {
	// client connection establishes
	for key := range n.SystemNodes {
		hostingPort := "localhost:" + key

		conn, err := grpc.NewClient(hostingPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Not working, %v", err)
		}

		//defer conn.Close()
		client := pb.NewRAClient(conn)
		n.SystemNodes[key] = client

		fileLog.Printf("Node: %s Client connection established on port %s... \n", n.Portnumber, key) // homemade error handling
		termLog.Printf("Client connection established on port %s... \n", key)
	}
}

// Establishes the server connection
func (n *Node) StartServer() {
	// Server section
	tmp := ":" + n.Portnumber
	listener, err := net.Listen("tcp", tmp) //":5050"
	if err != nil {
		log.Fatalf("Did not work in server, error: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRAServer(grpcServer, n) //registers the unimplemented server (implements the server)

	fileLog.Printf("Server running on port %s... \n", n.Portnumber)
	termLog.Printf("Server running on port %s... \n", n.Portnumber)

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Could not work in server")
	}

	for {

	}
}

func main() {
	// The logging system
	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	/* Create two independent loggers, one for logging to system.log and one for logging in terminal */
	fileLog = log.New(file, "", log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0)

	// Get the command line arguments
	args := os.Args[1:] //this is without the program path in the argument array
	//fmt.Printf("The args: %v \n", args) //homemade error handling

	nodeId, err := strconv.Atoi(args[0]) //extracts the nodes id from argument list
	if err != nil {
		panic(err)
	}

	myPort := args[nodeId] // heres to hoping this will be seen as a int
	args = args[1:]        // remove the first index from the array

	startMap := make(map[string]pb.RAClient)

	for key := range args {
		if args[key] != myPort {
			startMap[args[key]] = nil
			//fmt.Printf("The key is: %d \n", key)   // homemade for errorhandling
			//fmt.Printf("the value is: %s \n", args[key])
		}
	}

	// fmt.Printf("My id: %d, Myport: %s \n", nodeId, myPort)     // homemade error handling
	node := &Node{Portnumber: myPort, Lamport: 0, Queue: make([]*pb.Request, 0), State: Release, SystemNodes: startMap, PermissionCount: 0}
	//node := &Node{Portnumber: myPort, Lamport: 0, State: Release, SystemNodes: startMap}
	go node.StartServer()
	time.Sleep(10 * time.Second)
	node.JoinSystem()
	time.Sleep(5 * time.Second)

	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)
		node.RequestAccess()
	}

	for {

	}
}
