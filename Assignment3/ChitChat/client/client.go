package main

import (
	"context"
	"fmt"
	"io"
	"log"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"math/rand" //for a random id

	rn "github.com/random-names/go" // for a random name

	pb "ChitChat/grpc" //pb used to be proto
)

/* ChitChat User */

type ChitChatter struct {
	ID   int32
	Name string
}

/* Creates a ChitChat User with random id and name*/
func makeRandomClient() ChitChatter {
	// generate a random id
	randomID := rand.Int31()
	// generate a random user name with the random name methods from https://github.com/random-names/go
	randomName, err2 := rn.GetRandomName("./all.last", &rn.Options{})
	if err2 != nil {
		fmt.Println(err2)
	}

	chitChatter := ChitChatter{
		ID:   randomID,
		Name: randomName,
	}

	return chitChatter
}

/* Function to increment the lamport timestamp */

func incrementLamport(lamport *int32) {
	*lamport++
}

/* Function to synchronize the lamport timestamp with the ChitChat Server Lamport Timestamp */
func syncLamport(localLamport *int32, serverLamport int32) {
	*localLamport = max(*localLamport, serverLamport) + 1
}

/* JoinChat function call with lamport increments*/
func localJoinChat(client pb.ChitChatServiceClient, localChitChatter ChitChatter, lamportPointer *int32, isActivePointer *bool) grpc.ServerStreamingClient[pb.ChitChatMessage] {

	incrementLamport(lamportPointer)
	stream1, err := client.JoinChat(context.Background(), &pb.UserLamport{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: *lamportPointer})
	if err != nil {
		log.Fatalf("Not working in client 1")
	}
	*isActivePointer = true
	return stream1
}

/* LeaveChat function call with lamport increments*/
func localLeaveChat(client pb.ChitChatServiceClient, localChitChatter ChitChatter, lamportPointer *int32, isActivePointer *bool) {

	incrementLamport(lamportPointer)
	client.LeaveChat(context.Background(), &pb.UserLamport{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: *lamportPointer})
	*isActivePointer = false
	log.Println("client: I left the chat")
}

func receiveMessages(msgStream grpc.ServerStreamingClient[pb.ChitChatMessage], lamportPointer *int32, isActivePointer *bool) {
	for *isActivePointer {
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}
		syncLamport(lamportPointer, msg.User.GetLamport())
		log.Println(msg.Message)
	}
}

func localSendMessage(client pb.ChitChatServiceClient, localChitChatter ChitChatter, lamportPointer *int32, message string) {
	incrementLamport(lamportPointer)
	client.Publish(context.Background(), &pb.ChitChatMessage{User: &pb.UserLamport{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: *lamportPointer}, Message: message})
}

func SendMessageLoop(client pb.ChitChatServiceClient, localChitChatter ChitChatter, lamportPointer *int32) {

	for i := 0; i < 5; i++ {
		message, err := rn.GetRandomName("./all.last", &rn.Options{})
		message = fmt.Sprintf("%s from loop %d", message, i)
		if err != nil {
			log.Fatalf("Not working in messageLoop")
		}
		localSendMessage(client, localChitChatter, lamportPointer, message)
		time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	}
}

func main() {

	/* Lamport Timestamp */
	localLamport := int32(0)
	var lamportPointer *int32 = &localLamport

	/* ChitChatter State */
	isActive := false
	var isActivePointer *bool = &isActive

	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()
	client := pb.NewChitChatServiceClient(conn)

	// make a user with random name and random id
	localChitChatter := makeRandomClient()

	/* join chat method call */

	stream1 := localJoinChat(client, localChitChatter, lamportPointer, isActivePointer)

	go receiveMessages(stream1, lamportPointer, isActivePointer)

	go SendMessageLoop(client, localChitChatter, lamportPointer)

	time.Sleep(40 * time.Second)

	localLeaveChat(client, localChitChatter, lamportPointer, isActivePointer)

	// Below for loop lets the client live forever!
	// migth be irrelevant
	for {
	}

}

/* FREEZER */

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

// older methods we might be able to reuse or use as guides
//client.SendMessage(context.Background(), &proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "Hello World!", Lamport: 10})
//Stream.Send(&proto.ChitChatMessage{User: &proto.User{Id: 1, Name: "Alice"}, Message: "This is a message", Lamport: 10})

/*OLD recieve message loop*/

/*

//hardcoded forloop for testing that a client can actually leave the chat
//this happens after a client has received 2 broadcasts

for i := 0; i < 2; i++ {
		msg, err := stream1.Recv()

		if err == io.EOF {
			break
		}
		log.Println(msg.Message)

	}
*/
