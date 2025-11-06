package main

import (
	"log"
	"os"
	"fmt"
	"net"
	"strconv"
	"container/list"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "RA/grpc"

)


type Node struct {
	pb.UnimplementedRAServer
	portnumber string;
	lamport int32;
	queue *list.List;
	State state; 	// RELEASE, WANTED, & HELD

}

type state string

const (
	Release state = "RELEASE" 	
	Wanted 	state = "WANTED"	
	Held	state = "HELD"
)

// Make Client first then server
func (s *Node) JoinSystem() {	//doesn't feel like the right way to include the pointer

	// Client section
	hostingPort := "localhost:" + s.portnumber

	// hostingPort in the grpc call below, used to be "localhost:5050"
	conn, err := grpc.NewClient(hostingPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()
	client := pb.NewRAClient(conn)

	fmt.Printf("Client connection established on port %s... \n", s.portnumber)
	fmt.Printf("Client... %v \n", client)

	// Server section
	tmp := ":" + s.portnumber
	listener, err := net.Listen("tcp", tmp) //":5050"
	if err != nil {
		log.Fatalf("Did not work in server, erro: ", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRAServer(grpcServer, s) //registers the unimplemented server (implements the server)

	fmt.Printf("Server running on port %s...", s.portnumber)

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Could not work in server")
	}
}

func main() {
	// Get the command line arguments
	args := os.Args[1:] //this is without the program path in the argument array
	input := args[0]
	myId, err := strconv.Atoi(input)
	if err != nil {
		panic(err)
	}
	myPort := args[myId] // heres to hoping this will be seen as a int

	fmt.Printf("My id: %d, Myport: %s \n", myId, myPort)
	// potentially 
	server := &Node{portnumber: myPort, lamport: 0, queue: list.New(), State: Release} 
	server.JoinSystem()

}