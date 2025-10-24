package main

import (
	pb "ChitChat/grpc" //pb used to be proto
	"log"
	"net"
	"time"
	"fmt"
	"google.golang.org/grpc"
)

type ChitChatServer struct {
	pb.UnimplementedChitChatServer
}

func (s *ChitChatServer) TestServerStream(req *pb.User, stream pb.ChitChat_TestServerStreamServer) error {
	for i := 0; i < 10; i++ {
		msg := fmt.Sprintf("Hello %s - %d", req.Name, i)// get message later

		
		if err := stream.Send(&pb.Broadcast{Message: msg}); err != nil {
			return err
		}

		time.Sleep(500 * time.Millisecond) 
	}

    return nil
}

/*
// One way stream from server to chat
func (s *ChitChatServer) JoinChat(joinReq *pb.ChitChatMessage, stream pb.ChitChat_JoinChatServer[Broadcast]) error {
	for i := 0; i < 10; i++ {
		message := "this is a test message?"// get message later

		if err := stream.Send(&pb.Broadcast{Message: message}); err != nil {
			return err
		}

		time.Sleep(10 * time.Millisecond) 
	}

    return nil
}


func (s *ChitChatServer) JoinChat(proto.BidiStreamingServer[proto.ChitChatMessage, proto.Broadcast]) error {
	//log.Printf("Client '%v' with id: '%s' joined the chat \n ", proto.ChitChatMessage, proto.Broadcast)
	return status.Errorf(codes.Unimplemented, "method JoinChat not implemented")
}

// removed clientUser *proto.User, from the arguments
func (s *ChitChatServer) JoinChat(grpc.BidiStreamingServer[proto.ChitChatMessage, proto.ChitChatMessage]) error {
	// save the stream to the server struct??

	// Below is before changes
	//log.Printf("Client '%s' with id: '%d' joined the chat \n", clientUser.Name, clientUser.Id)
	return status.Errorf(codes.Unimplemented, "method JoinChat not implemented")
}
*/
/* Older functions that should still work in the big picture
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

	pb.RegisterChitChatServer(grpcServer, s) //registers the unimplemented server (implements the server)

	err = grpcServer.Serve(listener) //activates the server

	if err != nil {
		log.Fatalf("Did not work in server")
	}

}*/

func main() {
	//server := &ChitChatServer{}
	//server.start_server()

	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Did not work in server")
	}

	grpcServer := grpc.NewServer()
	pb.RegisterChitChatServer(grpcServer, &ChitChatServer{})
	log.Println("Server running on port 5050...")
	if err := grpcServer.Serve(lis); err !=nil {
		log.Fatalf("Did not work in server: %v", err)
	}
}
