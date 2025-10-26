package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"math/rand" //for a random id

	rn "github.com/random-names/go" // for a random name

	pb "ChitChat/grpc" //pb used to be proto
)


func makeRandomClient() *pb.UserLamport {
	var lamport int32 = 0
	// make a random id
	RandomId := rand.Int31()
	// make a random user name with the random name methods from https://github.com/random-names/go
	randomName, err2 := rn.GetRandomName("./all.last", &rn.Options{})
	if err2 != nil {
		fmt.Println(err2)
	}
	user := &pb.UserLamport{Id: RandomId, Name: randomName, Lamport: lamport}
	return user
}
/*
func int32 updateLamport(l lamport){
	
}
func int32 syncLamport(client lamport, server lamport){}
*/

func main() {
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()
	client := pb.NewChitChatServiceClient(conn)

	userTest := makeRandomClient()

	// join chat method call
	stream1, _ := client.JoinChat(context.Background(), userTest)

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

	client.LeaveChat(context.Background(), userTest)

	// log.Println("client: trying to leave chat")

	// Below for loop lets the client live forever!
	// migth be irrelevant
	for {
	}

	// older methods we might be able to reuse or use as guides
	//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
	//Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})
}
