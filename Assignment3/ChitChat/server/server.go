package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"

	pb "ChitChat/grpc" //pb used to be proto
)

var fileLog *log.Logger
var termLog *log.Logger

// The ChitChatServer containing:
// UnimplementedChitChatServiceServer (which makes go happy),
// a 32 bit integer lamport,
// a map of active user streams to clients
type ChitChatServer struct {
	pb.UnimplementedChitChatServiceServer
	lamport     int32
	userStreams map[int32]pb.ChitChatService_JoinChatServer
	streamStart chan struct{}
}

// JoinChat receives a proto ChitChatMessage stream and a proto user
// Adds the received stream with the User ID to the server's map of active user streams
// And then broadcasts to all active users in the map
func (s *ChitChatServer) JoinChat(req *pb.User, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams[req.Id] = stream //map version
	s.streamStart <- struct{}{}

	
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	fileLog.Printf("/ Server / Client %d Joins Chat / Lamport %d", req.Id, s.lamport)
	
	// message for joining
	msg := fmt.Sprintf("%s with id %d joined Chit Chat at logical time %d", req.Name, req.Id, s.lamport)


	s.Broadcast(req, msg)

	select {} 
}

// Does lamport synchronization before logging and removing the user from the map
func (s *ChitChatServer) LeaveChat(ctx context.Context, req *pb.User) (*pb.Empty, error) {
	
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	fileLog.Printf("/ Server / Client %d Leaves Chat / Lamport %d", req.Id, s.lamport)
	msg := fmt.Sprintf("%s with id %d left Chit Chat at logical time %d", req.Name, req.Id, s.lamport)

	delete(s.userStreams, req.Id)

	s.Broadcast(req, msg)

	return &pb.Empty{}, nil
}

// Takes a context and a ChitChatMessage from client to further broadcast
// Also does lamport arithmetics
func (s *ChitChatServer) Publish(ctx context.Context, req *pb.ChitChatMessage) (*pb.Empty, error) {
	
	if s.lamport < req.User.Lamport {
		s.lamport = req.User.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	fileLog.Printf("/ Server / Receive message from user %d / Lamport %d", req.User.Id, s.lamport)
	
	msg := fmt.Sprintf("%s at logical time %d: %s", req.User.Name, s.lamport, req.GetMessage())
	s.Broadcast(req.GetUser(), msg)
	return &pb.Empty{}, nil
}

// Broadcasts returns an error if it did not succesfully send
// Broadcasts a message to all clients in the list of user streams
// The userStreams array is updated in JoinChat to contain multiple streams
func (s *ChitChatServer) Broadcast(req *pb.User, msg string) error {

	waitGroup := sync.WaitGroup{}
	done := make(chan int32)
	s.lamport++
	broadcastMsg := &pb.ChitChatMessage{User: req, Message: msg, Lamport: s.lamport}

	for key, _ := range s.userStreams {

		waitGroup.Add(1)

		go func(clientId int32) {
			defer waitGroup.Done()
			err := s.userStreams[clientId].Send(broadcastMsg)
			if err != nil {
				log.Printf("error trying to send to client: %d (stream closed) %v", clientId, err)
			}
			fileLog.Printf("/ Server / Broadcast message to user %d from user %d / Lamport %d", clientId, req.Id, s.lamport)
		}(key)
	}
	go func() {
		waitGroup.Wait()
		close(done)
	}()

	<-done
	return nil
}

// Calls for starting the server and instantiating logger objects
func main() {

	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	// Create two independent loggers
	fileLog = log.New(file, "", log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) // plain chat-style output

	server := &ChitChatServer{lamport: 0, userStreams: make(map[int32]pb.ChitChatService_JoinChatServer), streamStart: make(chan struct{})}
	server.start_server()

}

// Method for starting the server
func (s *ChitChatServer) start_server() {

	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServiceServer(grpcServer, s) //registers the unimplemented server (implements the server)

	fileLog.Printf("/ Server / Startup / %d", s.lamport)
	termLog.Println("Server running on port 5050...")

	// Below method inspired by https://github.com/grpc/grpc-go/blob/master/examples/features/gracefulstop/server/main.go
	go func() {
		<-s.streamStart  // tracks when the stream first starts
		time.Sleep(45 * time.Second) // waits 10 seconds before starting shut down
		grpcServer.Stop()
		fileLog.Printf("/ Server / shutdowned / %d", s.lamport)
		termLog.Println("Server force shutdown")
	}()

	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Did not work in server")
	}
}