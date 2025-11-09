package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"context"
	"strconv"
	"time"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "RA/grpc"

)

type Node struct {
	pb.UnimplementedRAServer
	Portnumber  string
	Lamport     int32 // sendRequest, receiveRequest (?), access Critical section
	Queue       []pb.Request
	State       State                  // RELEASE, WANTED, & HELD
	SystemNodes map[string]pb.RAClient // portnumber and proto Ricart Agrawala client
}

type State string

const (
	Release State = "RELEASE"
	Wanted  State = "WANTED"
	Held    State = "HELD"
)

var mu sync.Mutex
var cond = sync.NewCond(&mu)

// Make Client first then server
func (n *Node) JoinSystem() {
	// client connection establishes
	for key := range n.SystemNodes {
		hostingPort := "localhost:" + key

		conn, err := grpc.NewClient(hostingPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Not working, %v", err)
		}

		defer conn.Close()
		client := pb.NewRAClient(conn)
		n.SystemNodes[key] = client

		fmt.Printf("Client connection established on port %s... \n", key) // homemade error handling
	}

}

func (n *Node) CriticalSection() {
	fmt.Printf("there is established accessed the critical section node: %s", n.Portnumber)
	n.Lamport++

	fmt.Printf("there is no more access to critical section node: %s", n.Portnumber)
}

func (n *Node) RequestCriticalSectionAccess(ctx context.Context, req *pb.Request) (*pb.Response, error) {

	if n.State == Held || (n.State == Wanted || n.Lamport < req.Lamport) {

		n.Queue = append(n.Queue, *req)

	} else {
		return &pb.Response{Permission: true}, nil
	} 
	mu.Lock()
	for n.State != Release{
		cond.Wait()
	}
	// TODO empty the queue here as well
	mu.Unlock()
	return &pb.Response{Permission: true}, nil 
}

func (n *Node) RequestAccess() {
	n.State = Wanted
	n.SendRequests()

	if n.State == Held {
		n.CriticalSection()
		n.State = Release
		cond.Signal()  // empty the queue (once the functionality has been added properly)
	}
}

func (n *Node) SendRequests() {
	waitGroup := sync.WaitGroup{}
	n.Lamport++
	
	for node, _ := range n.SystemNodes {

		waitGroup.Add(1)

		go func(nodeConnection string) {
			defer waitGroup.Done()
			
			_, err := n.SystemNodes[nodeConnection].RequestCriticalSectionAccess(context.Background(), &pb.Request{
				Portnr: n.Portnumber, 
				Lamport: n.Lamport,
			})
			if err != nil {
				log.Printf("error trying to send request to node: %s  %v", nodeConnection, err)
			}
		}(node)
	}
	go func() {
		waitGroup.Wait()
		n.State = Held
	}()
}


func (n *Node) StartServer() {
	// Server section
	tmp := ":" + n.Portnumber
	listener, err := net.Listen("tcp", tmp) //":5050"
	if err != nil {
		log.Fatalf("Did not work in server, error: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRAServer(grpcServer, n) //registers the unimplemented server (implements the server)

	fmt.Printf("Server running on port %s... \n", n.Portnumber)

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Could not work in server")
	}

	for { 

	}
}

func main() {
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
	node := &Node{Portnumber: myPort, Lamport: 0, Queue: make([]pb.Request, 0), State: Release, SystemNodes: startMap}
	go node.StartServer()
	time.Sleep(1 * time.Second)
	node.JoinSystem()

	for {

	}
}
