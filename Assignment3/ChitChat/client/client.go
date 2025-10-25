package main

import (
	"log"
	"context"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "ChitChat/grpc" //pb used to be proto

)

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()	
	client := pb.NewChitChatServiceClient(conn)

	// join chat method
	stream1, _ := client.JoinChat(context.Background(), &pb.UserLamport{Name: "Alice", Id: 0})
	// Below loop is where a client receives a response to a joinChat request
	// We suspect the stream might be closed too soon
	// The internet talked about perhaps using a goroutine to keep it alive
	// but we did not fully understand how that would work.
	for {
		msg, err := stream1.Recv()
		
		if err == io.EOF {
			break
			//log.Println("Something went wrong with receiving from stream: ")
		}
		log.Println("JoinChat: ", msg.Message)
	}

	// Below for loop lets the client live forever!
	// migth be irrelevant
	for {

	}


	// older methods we might be able to reuse or use as guides
	//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
	//Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})
	
	//client.LeaveChat(context.Background(), &proto.User{Id: 1, Name: "Alice"})
}
