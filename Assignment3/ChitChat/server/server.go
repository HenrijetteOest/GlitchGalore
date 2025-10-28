package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"os"
	"io"

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
func (s *ChitChatServer) JoinChat(req *pb.User, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams[req.Id] = stream //map version

	//log.Printf("Local Lamport in JoinChat before sync: %d and message lamport: %d", s.lamport, req.Lamport)
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	//log.Printf("Lamport in Join after sync: %d", s.lamport)
	// message for joining
	msg := fmt.Sprintf("JoinChat: Participant %s with id %d joined Chit Chat at logical time L - %d", req.Name, req.Id, s.lamport)

	//fmt.Printf("arrayet er opdateret: %v \n", s.userStreams) //homemade debug statement (delete later)
	s.Broadcast(req, msg)
	//log.Printf("Broadcasting burde være færdigt...") //homemade debug statement (delete later)

	select {} // homemade from chat
}

func (s *ChitChatServer) LeaveChat(ctx context.Context, req *pb.User) (*pb.Empty, error) {
	// message for leaving
	//log.Printf("Local Lamport in LeaveChat before sync: %d and message lamport: %d", s.lamport, req.Lamport)
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	//log.Printf("Lamport in LeaveChat after sync: %d", s.lamport)
	msg := fmt.Sprintf("LeaveChat: Participant %s with id %d left Chit Chat at logical time L - %d", req.Name, req.Id, s.lamport)

	delete(s.userStreams, req.Id)

	s.Broadcast(req, msg)

	return &pb.Empty{}, nil
}

func (s *ChitChatServer) Publish(ctx context.Context, req *pb.ChitChatMessage) (*pb.Empty, error) {
	//log.Printf("Local Lamport in Publish before sync: %d and message lamport: %d", s.lamport, req.User.Lamport)
	//sync mechanism works
	if s.lamport < req.User.Lamport {
		s.lamport = req.User.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	//log.Printf("Lamport in Publish after sync: %d", s.lamport)
	msg := fmt.Sprintf("%s: %s - Timestamp: %d", req.GetUser().GetName(), req.GetMessage(), s.lamport)
	s.Broadcast(req.GetUser(), msg)
	return &pb.Empty{}, nil
}

// Broadcasts a message to all clients in the list of user streams
// The userStreams array is updated in JoinChat to contain multiple streams
func (s *ChitChatServer) Broadcast(req *pb.User, msg string) error {
	//fmt.Println("Number of clients with 'open' streams: ", len(s.userStreams)) // homemade debugging

	waitGroup := sync.WaitGroup{}
	done := make(chan int32)
	s.lamport++
	broadcastMsg := &pb.ChitChatMessage{User: req, Message: msg, Lamport: s.lamport}

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

	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}

	//code below (line 127-132) comes from chatGPT
	defer file.Close()
	// Create a multi-writer to write to both the file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)
	// Set the log output to the multi-writer
	log.SetOutput(multiWriter)
	log.Println("Logging to both file and terminal!")

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
	log.Printf("Server, Event: Startup, Lamport Timestamp: %d", s.lamport)
	log.Println("Server running on port 5050...")

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	

}
