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
	Lamport         int32 // sendRequest, receiveRequest (?), access Critical section
	State           State // RELEASE, WANTED, & HELD
	Queue           []*pb.Request
	SystemNodes     map[string]pb.RAClient // portnumber and proto Ricart Agrawala client
	PermissionCount int
}

/*---------- Type State with three options -----------------------------*/
type State string

const (
	Release State = "RELEASE"
	Wanted  State = "WANTED"
	Held    State = "HELD"
)

/*---------- Make Lock and Condition Variable -----------------------------*/
var mu sync.Mutex // for deferring responding to requests
var cond = sync.NewCond(&mu)

/*---------- Create Global Log variables ----------------------------------*/
var fileLog *log.Logger // For logging to system.log
var termLog *log.Logger // For logging to the terminal

// prints that connection established to Critical Section
func (n *Node) CriticalSection() {
	/*---------- Log Status inside the Critical Section ----------------------------------*/
	fileLog.Printf("Node: %s made it to the Critical Section! MUAHAHAAAA \n", n.Portnumber)
	termLog.Println("I have made it to the Critical Section! MUAHAHAAAA")

	/*---------- Increase the Lamport ----------------------------------------------------*/
	mu.Lock()
	n.Lamport++
	mu.Unlock()

	/*---------- Stay in Critical Section for 4 Seconds ----------------------------------*/
	time.Sleep(4 * time.Second)
}

func (n *Node) EnterCriticalSection() {
	if n.State != Held {
		fileLog.Printf("Node: %s tried to enter Critical Section but state was not HELD", n.Portnumber)
		termLog.Println("I tried to enter the Critical Section in a not HELD state")
		return
	}

	/*---------- Enter the Critical Section -------------------*/
	n.CriticalSection()

	/*---------- Log Status on Leaving Critical Section -------------------*/
	fileLog.Printf("Node: %s has left the Critical Section", n.Portnumber)
	termLog.Println("I have left the Critical Section ")

	/*---------- Reset PermissionCount and Clean up -------------------*/
	mu.Lock()             // Release and wake any deferred requests
	n.PermissionCount = 0 // Reset how many Permissions we have received.
	n.State = Release     // No longer inside the Critical Section
	mu.Unlock()

	/*---------- Send Permissions to All Nodes Waiting for my Permission --------*/
	for _, req := range n.Queue {
		_, err := n.SystemNodes[req.Portnr].SendPermission(context.Background(), &pb.Response{Portnr: n.Portnumber, Lamport: n.Lamport})
		if err != nil {
			log.Printf("Error sending deferred permission to %s, Error: %v", req.Portnr, err)
		}
		//n.Queue = n.Queue[1:] // Remove Request from queue
	}
	n.Queue = nil //empty the queue
}

// RequestCriticalSectionAccess is our rpc call
func (n *Node) RequestAccess(ctx context.Context, req *pb.Request) (*pb.Empty, error) {
	/*---------- Log Status SendPermission ------------------------------------------------*/
	fileLog.Printf("Node: %s received request from Node: %s", n.Portnumber, req.Portnr)
	termLog.Printf("I have received a request from Node: %s", req.Portnr)

	/*---------- Convert Portnumbers from Strings to Integers -----------------------------*/
	myPortnr, _ := strconv.Atoi(n.Portnumber)
	reqPortnr, _ := strconv.Atoi(req.GetPortnr())

	/*---------- Compare Whether to Immediately Send a Respond or wait --------------------*/
	// append to queue if we are in the critical section,
	// or we want to and we either have higher lamport priority or the same lamport with higher portnumber
	if n.State == Held || (n.State == Wanted && (n.Lamport < req.Lamport || (n.Lamport == req.Lamport && myPortnr > reqPortnr))) { //
		mu.Lock()
		n.Queue = append(n.Queue, req)
		mu.Unlock()
	} else {
		/*---------- Send Immediate Reply -------------------------------------------------*/
		_, err := n.SystemNodes[req.GetPortnr()].SendPermission(context.Background(), &pb.Response{Portnr: n.Portnumber, Lamport: n.Lamport})
		if err != nil {
			// this error occurs if we can't send permission to a node because there is something wrong with the receiving node
			// TODO remove the node from the SystemNodes map (this might be more relevant to add to the sendRequest() error section )
			log.Printf("error trying to send permission to node: %s  %v", req.GetPortnr(), err)
		}
	}
	/*---------- End Method without Returning Anything -------------------------------------*/
	return &pb.Empty{}, nil
}

// sending a permission reson
func (n *Node) SendPermission(ctx context.Context, resp *pb.Response) (*pb.Empty, error) {
	/*---------- Log Status SendPermission ---------------------------------------*/
	fileLog.Printf("Node: %s has received permission from Node: %s", n.Portnumber, resp.GetPortnr())
	termLog.Printf("I have received a permission")

	/*---------- Increase PermissionCount ----------------------------------------*/
	mu.Lock()
	n.PermissionCount++
	termLog.Printf("Current permission count: %d, total permission count needed: %d", n.PermissionCount, len(n.SystemNodes))
	cond.Broadcast() // wake those who are waiting for permissions to see if we have received enough permissions
	mu.Unlock()

	/*---------- End method ----------------------------------------*/
	return &pb.Empty{}, nil
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
	mu.Unlock()

	/*---------- Asynchronously Send Requests to all other Nodes ----------------------*/
	for node := range n.SystemNodes {
		go func(nodeConnection string) { // why is this inside a go func?
			/*---------- Log Status SendRequest ---------------------------------------*/
			fileLog.Printf("Node: %s will ask Node: %s for permission", n.Portnumber, nodeConnection)
			termLog.Printf("I will request permission from Node: %s", nodeConnection)

			/*---------- The rpc call RequestAccess -----------------------------------*/
			_, err := n.SystemNodes[nodeConnection].RequestAccess(
				context.Background(),
				&pb.Request{
					Portnr:  n.Portnumber,
					Lamport: n.Lamport,
				},
			)
			if err != nil { // error handling
				log.Printf("Failed sending a request to node %s: %v", nodeConnection, err)
			}
		}(node)
	}

	/*---------- Wait for Permissions -----------------------------------*/
	mu.Lock() // Wait for all replies using sync.Cond
	for int(n.PermissionCount) < len(n.SystemNodes) {
		cond.Wait() // goroutine sleeps until cond.Broadcast() or cond.Signal() is called
	}

	n.State = Held // Enough Permissions acquired, change state
	mu.Unlock()

	/*---------- Enter Critical Section -----------------------------------*/
	n.EnterCriticalSection()
}

// Make a Client connection and save it to the SystemNodes map
func (n *Node) JoinSystem() {
	/*---------- Establish Client Connections to all relevant ports --------------*/
	// establish connection to all saved portnumbers in the SystemNodes map
	n.Lamport++
	for key := range n.SystemNodes {
		hostingPort := "localhost:" + key // Format the portnumber correctly

		conn, err := grpc.NewClient(hostingPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Could not establish a connection / make a new grpc client, error: %v", err)
		}

		/*---------- Use Connection to make a Client -------------------------*/
		n.SystemNodes[key]:= pb.NewRAClient(conn)

		NodeConn := pb.NewRAClient(conn)
		n.SystemNodes[key] = NodeConn

		/*---------- Log Status for Client -----------------------------------*/
		fileLog.Printf("Node: %s Client connection established on port %s... \n", n.Portnumber, key) // homemade error handling
		termLog.Printf("Client connection established on port %s... \n", key)
	}
}

// Establishes the server connection
func (n *Node) StartServer() {
	/*---------- Create Listener on node Port -----------------------*/
	tmp := ":" + n.Portnumber
	listener, err := net.Listen("tcp", tmp) // used to be ":5050"
	if err != nil {
		log.Fatalf("Did not make the server listener, error: %v", err)
	}

	/*---------- Make GrpcServer ---------------------------------*/
	grpcServer := grpc.NewServer()
	pb.RegisterRAServer(grpcServer, n) //registers the unimplemented server (implements the server)

	/*---------- Log Status? ---------------------------------*/
	fileLog.Printf("Server running on port %s... \n", n.Portnumber)
	termLog.Printf("Server running on port %s... \n", n.Portnumber)

	/*---------- Start the actual Server ---------------------------------*/
	go func() {
		err = grpcServer.Serve(listener) //activates the server
		if err != nil {                  // Only enters this section if the server could not start
			panic(err)
		}
	}()
}

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

	myPort := args[nodeId] // heres to hoping this will be seen as a int
	args = args[1:]        // remove the first index from the array (such that only we only have portnumbers left in the array)

	/*-------- Create a map containing the other relevant Portnumbers ------------*/
	startMap := make(map[string]pb.RAClient)
	for key := range args {
		if args[key] != myPort {
			startMap[args[key]] = nil
		}
	}

	/*-------- Create this Node struct --------------------------*/
	node := &Node{Portnumber: myPort, Lamport: 0, State: Release, Queue: make([]*pb.Request, 0), SystemNodes: startMap, PermissionCount: 0}

	/*-------- Start Server for this Node -----------------------*/
	node.StartServer()
	time.Sleep(10 * time.Second) // sleep so we can create more servers.

	/*-------- Establish connection to other servers ------------*/
	node.JoinSystem()
	time.Sleep(5 * time.Second)

	/*-------- Start Seeking Access to the Critical Section ------------*/
	// seek access 10 times and sleep up to 5 seconds between each call
	for i := 0; i < 10; i++ {
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)
		node.SendRequests()
	}

	for {
	}
}
