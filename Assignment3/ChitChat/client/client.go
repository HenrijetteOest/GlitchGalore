package main

import (
	"log"
	"context"
	"io"
	"math/rand" //for a random id

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
	RandomId := rand.Int31()
	stream1, _ := client.JoinChat(context.Background(), &pb.UserLamport{Id: RandomId, Name: "Alice", Lamport: 1})

	// Below loop is where a client receives a response to a joinChat request
	/*
	for {
		msg, err := stream1.Recv()
		
		if err == io.EOF {
			break
			//log.Println("Something went wrong with receiving from stream: ")
		}
		log.Println("JoinChat: ", msg.Message)
	}*/

	// hardcoded forloop for testing that a client can actually leave the chat
	// this happens after a client has received 2 broadcasts
	for i := 0; i < 2; i++ {
		msg, err := stream1.Recv()
		
		if err == io.EOF {
			break
			//log.Println("Something went wrong with receiving from stream: ")
		}
		log.Println(msg.Message)
		
	}

	client.LeaveChat(context.Background(), &pb.UserLamport{Id: RandomId, Name: "Alice", Lamport: 2})
	
	// log.Println("client: trying to leave chat")
	// example leavechat
	

	// Below for loop lets the client live forever!
	// migth be irrelevant
	for {	}

	// older methods we might be able to reuse or use as guides
	//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
	//Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})
}
