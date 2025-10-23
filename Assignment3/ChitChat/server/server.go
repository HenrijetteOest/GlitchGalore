package main

import (
	proto "ChitChat/grpc"
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChitChatServer struct {
	proto.UnimplementedChitChatServer
	//stream proto.ChitChat_JoinChatServer
}

// removed clientUser *proto.User, from the arguments
func (s *ChitChatServer) JoinChat(grpc.BidiStreamingServer[proto.ChitChatMessage, proto.ChitChatMessage]) error {
	// save the stream to the server struct??

	// Below is before changes
	//log.Printf("Client '%s' with id: '%d' joined the chat \n", clientUser.Name, clientUser.Id)
	return status.Errorf(codes.Unimplemented, "method JoinChat not implemented")
}
func (s *ChitChatServer) LeaveChat(ctx context.Context, clientUser *proto.User) (*proto.Empty, error) {
	log.Printf("Client '%s' with id: '%d' left the chat \n", clientUser.Name, clientUser.Id)
	return nil, status.Errorf(codes.Unimplemented, "method LeaveChat not implemented")
}

func (s *ChitChatServer) SendMessage(ctx context.Context, clientMessage *proto.ChitChatMessage) (*proto.Empty, error) {
	out := new(proto.Empty)
	log.Printf("Client %s says %s at time: %d \n", clientMessage.User.Name, clientMessage.Message, clientMessage.Lamport)
	return out, nil
}

func (s *ChitChatServer) start_server() {
	grpcServer := grpc.NewServer()
	listener, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	proto.RegisterChitChatServer(grpcServer, s) //registers the unimplemented server (implements the server)

	err = grpcServer.Serve(listener) //activates the server

	if err != nil {
		log.Fatalf("Did not work in server")
	}

}

func main() {
	server := &ChitChatServer{}
	server.start_server()
}
