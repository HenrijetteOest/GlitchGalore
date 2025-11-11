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

/*---------- Node Struct ----------------------------------------------*/
type Node struct {
	pb.UnimplementedRAServer
	Portnumber      string
	Lamport         int32 
	Timestamp		int32 // lamport at time of message sending
	State           State // RELEASE, WANTED, & HELD
	Queue           []*pb.Request
	SystemNodes     map[string]pb.RAClient // portnumber and connection 
	PermissionCount int
}

/*---------- Type State with three options -------------------------------*/
type State string

const (
	Release State = "RELEASE"
	Wanted  State = "WANTED"
	Held    State = "HELD"
)

/*---------- Make Lock ---------------------------------------------------*/
var mu sync.Mutex 
var cond = sync.NewCond(&mu) 

/*---------- Create Global Log variables ----------------------------------*/
var fileLog *log.Logger // For logging to system.log
var termLog *log.Logger // For logging to the terminal





func main() {
	/*-------- Log Functionality  -------------------------*/
	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	/* Create two independent loggers, one for logging to system.log and one for logging in terminal */
	fileLog = log.New(file, "", log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0)

	/*-------- Command Line Arguments --------------------*/
	// Save the arguments to a local array (without)the program path
	args := os.Args[1:]

	nodeId, err := strconv.Atoi(args[0]) //extracts the nodes id from argument list
	if err != nil {
		panic(err)
	}

	myPort := args[nodeId] 
	args = args[1:]        // remove the first index from the array (such that only we only have portnumbers left in the array)

	/*-------- Create a map containing the other relevant Portnumbers ------------*/
	startMap := make(map[string]pb.RAClient)
	for key := range args {
		if args[key] != myPort {
			startMap[args[key]] = nil
		}
	}

	/*-------- Create this Node struct --------------------------*/
	node := &Node{Portnumber: myPort, Lamport: 0, Timestamp: 0, State: Release, Queue: make([]*pb.Request, 0), SystemNodes: startMap, PermissionCount: 0}

	/*-------- Start Server for this Node -----------------------*/
	node.StartServer()
	time.Sleep(10 * time.Second) // sleep to ensure all nodes (servers) are up and running before trying to establish connections

	/*-------- Establish connection to other servers ------------*/
	node.JoinSystem()
	time.Sleep(5 * time.Second)

	/*-------- Start Seeking Access to the Critical Section ------------*/
	// seek access 10 times and sleep up to 5 seconds between each call
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)
		node.SendRequests() // Start the fun
	}

	for {

	}
}

// Establishes the server connection
func (n *Node) StartServer() {
	/*---------- Create Listener on node Port -----------------------*/
	tmp := ":" + n.Portnumber
	listener, err := net.Listen("tcp", tmp) // used to be ":5050"
	if err != nil {
		panic(err)
	}

	/*---------- Make GrpcServer ---------------------------------*/
	grpcServer := grpc.NewServer()
	pb.RegisterRAServer(grpcServer, n) //registers the unimplemented server (implements the server)


	/*---------- Log Status ---------------------------------*/
	// This print technically occurs too soon because we can't place it after err = grpcServer.Serve(listener)...
	fileLog.Printf("Server running on port %s... \n", n.Portnumber)
	termLog.Printf("Server running on port %s... \n", n.Portnumber)

	/*---------- Start the actual Server ---------------------------------*/
	go func(){
		err = grpcServer.Serve(listener) // Activates the server
		if err != nil {                  // Only enters this section if the server could not start
			panic(err)
		}
	}()
}

func (n *Node) JoinSystem() {
	/*---------- Establish Client Connections to all relevant ports --------------*/
	// establish connection to all saved portnumbers in the SystemNodes map
	mu.Lock()
	n.Lamport++
	mu.Unlock()
	
	for key := range n.SystemNodes {
		hostingPort := "localhost:" + key // Format the portnumber correctly

		conn, err := grpc.NewClient(hostingPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not establish a connection / make a new grpc client, error: %v", err)
		}
		
		nodeConn := pb.NewRAClient(conn)
		n.SystemNodes[key] = nodeConn

		/*---------- Log Status for Client -----------------------------------*/
		fileLog.Printf("Node: %s Client connection established on port %s... \n", n.Portnumber, key) // homemade error handling
		termLog.Printf("Client connection established on port %s... \n", key)
	}
}

// Once the node has received permission it calls to enter the CS from here
func (n *Node) SendRequests() {
	/*---------- Check State ---------------------------------------------*/
	if n.State != Release {
		return
	}

	/*---------- Change State and update Lamport -------------------------*/
	mu.Lock()
	n.State = Wanted
	n.Lamport++ // increment Lamport timestamp for this request
	n.Timestamp = n.Lamport
	n.PermissionCount = 0
	mu.Unlock()

	/*----------------------Send Requests to all other Nodes ----------------------*/
	for node := range n.SystemNodes {
		/*---------- Log Status SendRequest ---------------------------------------*/
		fileLog.Printf("Node: %s will ask Node: %s for permission", n.Portnumber, node)
		termLog.Printf("I will request permission from Node: %s", node)

			/*---------- The rpc call RequestAccess -----------------------------------*/
		_, err := n.SystemNodes[node].RequestAccess(
			context.Background(),
			&pb.Request{
				Portnr:  n.Portnumber,
				Lamport: n.Timestamp,	// lamport at time of sending request
			},
		)
		if err != nil { // error handling
			log.Printf("Failed sending a request to node %s: %v", node, err)
		}	
	}

	/*---------- Wait for Permissions -----------------------------------*/
	
	mu.Lock() 		
	for n.PermissionCount < len(n.SystemNodes) { 
		cond.Wait() // Wait for all replies using sync.Cond
	}
	
	n.State = Held 	// Enough Permissions acquired, change state
	mu.Unlock()
	n.CriticalSection()
	
}

func (n *Node) CriticalSection() {
	/*---------- Log Status inside the Critical Section ----------------------------------*/
	fileLog.Printf("Node: %s made it to the Critical Section! MUAHAHAAAA \n", n.Portnumber)
	termLog.Println("I have made it to the Critical Section! MUAHAHAAAA")

	/*---------- Increase the Lamport ----------------------------------------------------*/
	mu.Lock()
	n.Lamport++
	mu.Unlock()

	/*---------- Stay in Critical Section for 4 Seconds ----------------------------------*/
	time.Sleep(3 * time.Second)
	n.ExitCriticalSection()
}

func (n *Node) ExitCriticalSection() {
	if (n.State != Held) {
		fileLog.Printf("FAIL: Node: %s tried to exit the critical section without being in it...", n.Portnumber)
		termLog.Println("FAIL: I tried to exit the critical section without being in it... ")
	}
	/*---------- Log Status on Leaving Critical Section ----------------------------------*/
	fileLog.Printf("Node: %s has left the Critical Section", n.Portnumber)
	termLog.Println("I have left the Critical Section ")

	/*---------- Reset PermissionCount and Clean up --------------------------------------*/
	mu.Lock()             // Release and wake any deferred requests
	n.PermissionCount = 0 // Reset how many Permissions we have received.
	n.State = Release     // No longer inside the Critical Section
	n.Lamport++			  // Increment lamport when leaving the 
	n.Timestamp = n.Lamport	// update the timestamp to no longer be too low
	mu.Unlock()

	/*---------- Reply by Sending Permissions to All Nodes Waiting for my Permission ------*/
	for _, req := range n.Queue {
		_, err := n.SystemNodes[req.Portnr].SendPermission(context.Background(), &pb.Response{Portnr: n.Portnumber, Lamport: n.Timestamp})
		if err != nil {
			log.Printf("Error sending deferred permission to %s, Error: %v", req.Portnr, err)
		}
	}
	n.Queue = nil //empty the queue
}

/*-------------------------------- RPC CALLS ----------------------------------------*/

// The Caller sends a permission, while the receiving Node receives a permission
func (n *Node) SendPermission(ctx context.Context, resp *pb.Response) (*pb.Empty, error) {
	/*---------- Log Status SendPermission ---------------------------------------*/
	fileLog.Printf("Node: %s has received permission from Node: %s", n.Portnumber, resp.GetPortnr())
	termLog.Printf("I have received a permission from Node: %s", resp.GetPortnr())

	/*---------- Increase PermissionCount ----------------------------------------*/
	mu.Lock()
	n.PermissionCount++
	n.Lamport++
	termLog.Printf("Current permission count: %d, total permission count needed: %d", n.PermissionCount, len(n.SystemNodes))
	cond.Broadcast()
	mu.Unlock()
	

	/*---------- End method ------------------------------------------------------*/
	return &pb.Empty{}, nil
}


func (n *Node) RequestAccess(ctx context.Context, req *pb.Request) (*pb.Empty, error) {
	/*---------- Log Status SendPermission ---------------------Portnumber-----------------*/
	fileLog.Printf("Node: %s received request from Node: %s", n.Portnumber, req.Portnr)
	termLog.Printf("I have received a request from Node: %s", req.Portnr)

	/*---------- Convert PortnumbeTimestamp)ings to Integers -----------------------------*/
	myPortnr, _ := strconv.Atoi(n.Portnumber)
	reqPortnr, _ := strconv.Atoi(req.GetPortnr())

	mu.Lock()
	if req.Lamport > n.Lamport {
		n.Lamport = req.Lamport
	}
	n.Lamport++ 
	mu.Unlock()

	// UPDATED the lamport to timestamp at the time of request
	/*---------- Compare Whether to Immediately Send a Respond or wait --------------------*/
	// append to queue if we are in the critical section,
	// or we want to and our request has a higher priority than the others node, or if both request have the same priority and the requesting node has a higher portnumber
	if n.State == Held || (n.State == Wanted && (n.Timestamp < req.Lamport || (n.Timestamp == req.Lamport && myPortnr < reqPortnr))) {
		mu.Lock()
		n.Queue = append(n.Queue, req)
		mu.Unlock()
	} else {
		/*---------- Send Immediate Reply -------------------------------------------------*/
		_, err := n.SystemNodes[req.GetPortnr()].SendPermission(context.Background(), &pb.Response{Portnr: n.Portnumber, Lamport: n.Timestamp})
		if err != nil {
			// this error occurs if we can't send permission to a node because there is something wrong with the receiving node
			// TODO remove the node from the SystemNodes map (this might be more relevant to add to the sendRequest() error section )
			log.Printf("error trying to send permission to node: %s  %v", req.GetPortnr(), err)
		}
	}
	/*---------- End Method without Returning Anything -------------------------------------*/
	return &pb.Empty{}, nil
}