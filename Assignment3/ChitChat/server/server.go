package main

import (
	"log"
	"net"
	"time"
	"fmt"

	"google.golang.org/grpc"

	pb "ChitChat/grpc" //pb used to be proto

)

type ChitChatServer struct {
	pb.UnimplementedChitChatServiceServer
}

func (s *ChitChatServer) JoinChat(req *pb.UserLamport, stream pb.ChitChatService_JoinChatServer) error {
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Hello %s - %d", req.Name, i)// get message later

		
		if err := stream.Send(&pb.ChitChatMessage{Message: msg}); err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond) 
	}

    return nil
}

/*
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
