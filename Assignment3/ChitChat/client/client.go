package main

import (
	pb "ChitChat/grpc" //pb used to be proto
	"log"
	"context"
	"io"
	"google.golang.org/grpc"
)

func main() {
	//conn, err := grpc.NewClient("localhost:5050", grpc.WithTranspotCredentials(insecure.NewCredentials()))
	conn, err := grpc.Dial("localhost:5050", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()	// new line from the internet
	client := pb.NewChitChatClient(conn)

	stream1, _ := client.TestServerStream(context.Background(), &pb.User{Name: "Alicia", Id: 0})
	for {
		msg, err := stream1.Recv()
		if err == io.EOF {
			break
		}
		log.Println("TestServerStream: ", msg.Message)
	}

	//client.JoinChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})
	//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
	//Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})
	
	//client.LeaveChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})
}
