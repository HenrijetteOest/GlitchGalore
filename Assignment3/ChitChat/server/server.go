package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"google.golang.org/grpc"

	pb "ChitChat/grpc" //pb used to be proto
)

// userStreams is a map of active streams
// meant to receive new streams in JoinChat when a client joins
// and decrement when in LeaveChat, when a client leaves.
type ChitChatServer struct {
	pb.UnimplementedChitChatServiceServer
	lamport     int32
	userStreams map[int32]pb.ChitChatService_JoinChatServer
}

// The join chat method, whereby a client joins the chat
// The server then saves the client stream in an array of streams
// Also calls the local method Broadcast() which then sends a message
// to all active clients that "x has joined the chat"
func (s *ChitChatServer) JoinChat(req *pb.UserLamport, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams[req.Id] = stream //map version

	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	log.Printf("lamport in JoinChat %d", s.lamport)
	// message for joining
	msg := fmt.Sprintf("JoinChat: Participant %s with id %d joined Chit Chat at logical time L - %d", req.Name, req.Id, s.lamport)

	//fmt.Printf("arrayet er opdateret: %v \n", s.userStreams) //homemade debug statement (delete later)
	s.Broadcast(req, msg)
	//log.Printf("Broadcasting burde være færdigt...") //homemade debug statement (delete later)

	select {} // homemade from chat
}

func (s *ChitChatServer) LeaveChat(ctx context.Context, req *pb.UserLamport) (*pb.Empty, error) {
	// message for leaving
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	log.Printf("lamport in LeaveChat %d", s.lamport)
	msg := fmt.Sprintf("LeaveChat: Participant %s with id %d left Chit Chat at logical time L - %d", req.Name, req.Id, s.lamport)

	delete(s.userStreams, req.Id)

	s.Broadcast(req, msg)

	return &pb.Empty{}, nil
}

func (s *ChitChatServer) Publish(ctx context.Context, req *pb.ChitChatMessage) (*pb.Empty, error) {
	if s.lamport < req.User.Lamport {
		s.lamport = req.User.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	log.Printf("lamport i publish %d", s.lamport)
	msg := fmt.Sprintf("%s: %s - Timestamp: %d", req.GetUser().GetName(), req.GetMessage(), s.lamport)
	s.lamport++
	s.Broadcast(req.GetUser(), msg)
	return &pb.Empty{}, nil
}

// Broadcasts a message to all clients in the list of user streams
// The userStreams array is updated in JoinChat to contain multiple streams
func (s *ChitChatServer) Broadcast(req *pb.UserLamport, msg string) error {
	//fmt.Println("Number of clients with 'open' streams: ", len(s.userStreams)) // homemade debugging

	waitGroup := sync.WaitGroup{}
	done := make(chan int32)
	s.lamport++
	log.Print(s.lamport)
	broadcastMsg := &pb.ChitChatMessage{User: req, Message: msg}

	for key, _ := range s.userStreams {

		waitGroup.Add(1)

		//go broadcastToClient(broadcast msg, clientId)
		go func(clientId int32) {
			defer waitGroup.Done()
			err := s.userStreams[clientId].Send(broadcastMsg)
			if err != nil {
				log.Printf("error trying to send to client: %d (stream closed) %v", clientId, err)
			}
		}(key)
	}
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	<-done
	return nil
}

/*
// Broadcasts a message to all clients in the list of user streams
// The userStreams array is updated in JoinChat to contain multiple streams
func (s *ChitChatServer) Broadcast(req *pb.UserLamport, msg string) error {
	fmt.Println("Number of clients with 'open' streams: ", len(s.userStreams)) // homemade debugging

	// Go through all streams in the userStream list and send them the same message

	broadcastMsg := &pb.ChitChatMessage{Message: msg}

	for key, _ := range s.userStreams {
		err := s.userStreams[key].Send(broadcastMsg)
		if err != nil {
			log.Printf("error trying to send to clien: %d (stream closed) %v", key, err)
			continue
		}
	}

	return nil
}*/

func main() {

	server := &ChitChatServer{lamport: 0, userStreams: make(map[int32]pb.ChitChatService_JoinChatServer)}
	server.start_server()
}

// Uncomment method below later (and update it to fit with what is in main)
func (s *ChitChatServer) start_server() {

	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, s) //registers the unimplemented server (implements the server)
	log.Println("Server running on port 5050...")

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Did not work in server")
	}

}
