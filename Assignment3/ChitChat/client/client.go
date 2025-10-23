package main

import (
	proto "ChitChat/grpc"
	"context"
	"log"
	

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working")
	}

	client := proto.NewChitChatClient(conn)

	// When the stream works, we can use that to send messages to Server, instead of using the SendMessage rpc method
	// Below method was suggested through auto-generated grpc code, so no clue if it is actually correct
	Stream, err := client.JoinChat(context.Background())
	if err != nil {
		log.Fatalf("Error while joining chat: %v", err)
	}	

	//client.JoinChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})
	//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
	Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})
	
	client.LeaveChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})
}
