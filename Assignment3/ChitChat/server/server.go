package main

import (
	"log"
	"net"
	"fmt"

	"google.golang.org/grpc"

	pb "ChitChat/grpc" //pb used to be proto

)


// userStreams is an array of active streams
// meant to receive new streams in JoinChat when a client joins
// and decrement when in LeaveChat, when a client leaves.
type ChitChatServer struct {
	pb.UnimplementedChitChatServiceServer
	userStreams []pb.ChitChatService_JoinChatServer
}


// The join chat method, whereby a client joins the chat 
// The server then saves the client stream in an array of streams
// Also calls the local method Broadcast() which then sends a message
// to all active clients that "x has joined the chat"  
func (s *ChitChatServer) JoinChat(req *pb.UserLamport, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams = append(s.userStreams, stream)
	
	fmt.Printf("arrayet er opdateret: %v \n", s.userStreams) 	//homemade debug statement
	s.Broadcast(req)
	log.Printf("Broadcasting burde være færdigt...")	//homemade debug statement

	// From go tour website:
	// The select statement lets a goroutine wait on multiple communication operations
	// A select blocks until one of its cases can run, then it executes that case. It chooses one at random if multiple are ready. 
	select{}	// homemade from chat

    // return nil	// we never make it here due to the select{} block above (i think)
}

// Broadcasts to all clients in the list of user streams
// The userStreams array is updated in JoinChat to contain multiple streams 
func (s *ChitChatServer) Broadcast (req *pb.UserLamport) error {
	fmt.Println("Number of clients with 'open' streams: ", len(s.userStreams))
	
	// Go through all streams in the userStream list and send them the same message
	msg := fmt.Sprintf("Participant %s joined Chit Chat at logical time L - %d", req.Name, req.Lamport)
	broadcastMsg := &pb.ChitChatMessage{Message: msg}

	for i := 0; i < len(s.userStreams); i++ {
		// fmt.Println("	For loop iteration: ", i) 	//homemade debug statement
		err := s.userStreams[i].Send(broadcastMsg)
		if err != nil {
			log.Printf("error trying to send to clien: %d (stream closed) %v", i, err)
			continue
		}
	} 

	// fmt.Println("For loop over!")		//homemade debug statement
	return nil 
}

/* Uncomment method below later (and update it to fit with what is in main)
func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	pb.RegisterChitChatServer(grpcServer, s) //registers the unimplemented server (implements the server)

	err = grpcServer.Serve(listener) //activates the server

	if err != nil {
		log.Fatalf("Did not work in server")
	}

}
*/

func main() {
	//server := &ChitChatServer{}
	//server.start_server()

	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, &ChitChatServer{})
	log.Println("Server running on port 5050...")  // homemade, not part of the ITU template
	if err := grpcServer.Serve(lis); err !=nil {
		log.Fatalf("Did not work in server: %v", err)
	}

	// after some time close the server again and log it
}
