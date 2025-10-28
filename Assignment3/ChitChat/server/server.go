package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"os"
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
// Adds the received stream with the User ID to the global map of active user streams

// The join chat method, whereby a client joins the chat
// The server then saves the client stream in an array of streams
// Also calls the local method Broadcast() which then sends a message
// to all active clients that "x has joined the chat"
func (s *ChitChatServer) JoinChat(req *pb.User, stream pb.ChitChatService_JoinChatServer) error {
	s.userStreams[req.Id] = stream //map version
	s.streamStart <- struct{}{}

	//log.Printf("Local Lamport in JoinChat before sync: %d and message lamport: %d", s.lamport, req.Lamport)
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	fileLog.Printf("/ Server / Client %d Joins Chat / Lamport %d", req.Id, s.lamport)
	//log.Printf("Lamport in Join after sync: %d", s.lamport)
	// message for joining
	msg := fmt.Sprintf("%s with id %d joined Chit Chat at logical time %d", req.Name, req.Id, s.lamport)

	//fmt.Printf("arrayet er opdateret: %v \n", s.userStreams) //homemade debug statement (delete later)
	s.Broadcast(req, msg)
	//log.Printf("Broadcasting burde være færdigt...") //homemade debug statement (delete later)

	select {} // homemade from chat
}

// 
func (s *ChitChatServer) LeaveChat(ctx context.Context, req *pb.User) (*pb.Empty, error) {
	//log.Printf("Local Lamport in LeaveChat before sync: %d and message lamport: %d", s.lamport, req.Lamport)
	if s.lamport < req.Lamport {
		s.lamport = req.Lamport + 1
	} else {
		s.lamport = s.lamport + 1
	}
	fileLog.Printf("/ Server / Client %d Leaves Chat / Lamport %d", req.Id, s.lamport)
	//log.Printf("Lamport in LeaveChat after sync: %d", s.lamport)
	msg := fmt.Sprintf("%s with id %d left Chit Chat at logical time %d", req.Name, req.Id, s.lamport)

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
	fileLog.Printf("/ Server / Receive message from user %d / Lamport %d", req.User.Id, s.lamport)
	//log.Printf("Lamport in Publish after sync: %d", s.lamport)
	msg := fmt.Sprintf("%s at logical time %d: %s", req.User.Name, s.lamport, req.GetMessage())
	s.Broadcast(req.GetUser(), msg)
	return &pb.Empty{}, nil
}

// Broadcase returns an error if it did not succesfully send
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


// Calls for starting the server and create (?) the log file
func main() {

	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	// Create two independent loggers
	fileLog = log.New(file, "",log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) // plain chat-style output

	server := &ChitChatServer{lamport: 0, userStreams: make(map[int32]pb.ChitChatService_JoinChatServer), streamStart: make(chan struct{})}
	server.start_server()

	//fileLog.Printf("/ Server / before time sleep / %d", server.lamport)
	time.Sleep(45 * time.Second)
	//server.shutdown_server()

	/*
	//for shutting down the server after some time 
	//maybe this function should be renamed to start_and_shutdown_server
	time.Sleep(45 * time.Second)
	fileLog.Printf("/ Server / Shutdown / %d", server.lamport)
	termLog.Println("Server shutting down on port 5050...")
	//server.Shutdown()
	//Stop()
	*/
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
	
	
	go func() {
		<-s.streamStart
		time.Sleep(1 * time.Second)
		fileLog.Println("/ Server / initiating graceful shutdown / %d", s.lamport)
		termLog.Println("initiating graceful shutdown...")
		timer := time.AfterFunc(10 * time.Second, func() {
			fileLog.Printf("/ Server / force stop / %d ", s.lamport)
			termLog.Println("Server couldn't stop gracefully in time. Doing a force stop")
			grpcServer.Stop()
		})
		defer timer.Stop()
		grpcServer.GracefulStop();
		fileLog.Println("/ Server / Stopped gracefully / %d", s.lamport)
		//termLog.Println("Server stopped gracefully")
	}()
	
	
	err = grpcServer.Serve(listener) //activates the server
	if err != nil {
		log.Fatalf("Did not work in server")
	}
}

func (s *ChitChatServer) shutdown_server() {
    if s != nil {
        fileLog.Printf("/ Server / Shutdown / %d", s.lamport)
        termLog.Println("Server shutting down on port 5050...")
        //s.grpcServer.GracefulStop()
    } else {
        termLog.Println("No server instance to shut down.")
    }
}




/* FREEZER */
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

