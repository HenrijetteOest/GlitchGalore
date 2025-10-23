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

	client.JoinChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})

	client.LeaveChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})

	client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})

}
