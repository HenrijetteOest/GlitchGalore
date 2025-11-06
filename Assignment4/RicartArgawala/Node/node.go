package main

import (
	"log"
	"time"

)

var fileLog *log.Logger
var termLog *log.Logger

type NodeServer struct {
	pb.UnimplementedNodeServiceServer
	//id		int32	// for identifying a node
	//lamport int32		// for priority
	// queue of nodes before me 
}

func criticalSection() {
	//log.printf("Node: %d has accessed the critical section", id)

	//log.printf("Node: %d is leaving the critical section", id)
}

/*
func sendRequest() {


}
*/

func main() {
	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	fileLog = log.New(file, "",log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) 

	/* Lamport Timestamp */
	// localLamport = 0

	//  Change the first string "localhost:5050" to be compatible with the actual port number
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()
	client := pb.NewRicartAgrawalaServiceClient(conn)
}

func (s *NodeServer) start_server() {
    // Have a list of ports (or increment the port by)
	//var port = ":5000"

	// change the port number for each node
	// Try for some time to find the port, if it doesn't find the port
	// increment port number and try again
	//listener, err := net.Listen("tcp", port) //hope it works with the variable instead of string
	//if err != nil {
	//	log.Fatalf("Did not work in server")
	//}

	grpcServer := grpc.NewServer()
	pb.RegisterRicartAgrawalaServiceServer(grpcServer, s) //registers the unimplemented server (implements the server)

	//fileLog.Printf("/ Node %d as Server / Startup / %d", s.id, s.lamport)		//changes might not work immediately (needs the Id to function)
	//termLog.Printf("Node %d as Server running on port %s... \n", s.id, port)	//changes might not work immediately (needs the Id to function)

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Did not work in server")
	}
}

func createNode() {
	var portInt = 500
	var portString = ":" + 500

	for () {
		listener, err := net.Listen("tcp", portString) //hope it works with the variable instead of string
		time.Sleep(5 * time.Second)
		if err != nil {
			log.Fatalf("Did not work in server")
		}
		port + 1
		portString = ":" + portInt
	}
}

/* Things to log
 - Send request
 - receive request
 - receive response
 - actually accessing critical section
 - exiting critical section

*/


/*  FREEZER 

// method for stopping the server after some time
// Below method inspired by https://github.com/grpc/grpc-go/blob/master/examples/features/gracefulstop/server/main.go
go func() {
	time.Sleep(45 * time.Second) // waits 45 seconds before starting shut down
	grpcServer.Stop()
	fileLog.Printf("/ Server / shutdowned / %d", s.lamport)
	termLog.Println("Server force shutdown")
}()


*/