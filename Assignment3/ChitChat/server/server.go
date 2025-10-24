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


// the Joinchat method, whereby a client can join the chat 
// server saves the stream to client in an array of streams
// calls the local method Broadcast() which then sends a message
// to all active clients that "X" user has joined the chat
// note Broadcast has an error at the moment
func (s *ChitChatServer) JoinChat(req *pb.UserLamport, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams = append(s.userStreams, stream)
	
	fmt.Printf("arrayet er opdateret: %v \n", s.userStreams)
	s.Broadcast(req)
	log.Printf("Broadcasting burde være færdigt...")

    return nil
}

// function that should broadcast to all clients in the list users
// Broadcasting will only broadcast to a client once, as it never progresses
// in the for loop past the first iteration. 
// The array is updated in JoinChat to contain multiple streams
// we suspect the error to stem from either the way the Server sends to a stream
// or the way a client reads from a stream (which might not be continous (the stream might have closed)) 
func (s *ChitChatServer) Broadcast (req *pb.UserLamport) error {
	fmt.Println("Number of clients with 'open' streams: ", len(s.userStreams))
	
	for i := 0; i < len(s.userStreams); i++ {
		fmt.Println("	For loop iteration: ", i)
		msg := fmt.Sprintf("Participant %s joined Chit Chat at logical time L - %d", req.Name, req.Lamport)
		
		fmt.Println("	current 0th field in array: ", s.userStreams[0])

		// Below three lines hinder the loop from proceeding past the first iteration
		if err := s.userStreams[i].Send(&pb.ChitChatMessage{Message: msg}); err != nil {
			fmt.Println("	vi kom hertil!")
			return err
		}
	}
	fmt.Println("For loop over!")
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
	log.Println("Server running on port 5050...")
	if err := grpcServer.Serve(lis); err !=nil {
		log.Fatalf("Did not work in server: %v", err)
	}

	// after some time close the server again and log it
}
