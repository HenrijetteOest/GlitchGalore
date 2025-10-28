package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
	"os"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"math/rand" //for a random id

	rn "github.com/random-names/go" // for a random name

	pb "ChitChat/grpc" //pb used to be proto
)

/* Local Lamport Clock */
var localLamport int32
var mu sync.Mutex

var fileLog *log.Logger
var termLog *log.Logger


/* ChitChat User */
type ChitChatter struct {
	ID   int32
	Name string
}

/* Creates a ChitChat User with random id and name */
func makeRandomClient() ChitChatter {
	// generate a random id
	randomID := int32(rand.Intn(999)+1)
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
func incrementLamport() {
	mu.Lock()
	defer mu.Unlock()
	localLamport++
}

/* Function to synchronize the lamport timestamp with the ChitChat Server Lamport Timestamp */
func syncLamport(serverLamport int32) {
	mu.Lock()
	defer mu.Unlock()
	localLamport = max(localLamport, serverLamport) + 1
}

/* JoinChat function call with lamport increments */
func localJoinChat(client pb.ChitChatServiceClient, localChitChatter ChitChatter, isActivePointer *bool) grpc.ServerStreamingClient[pb.ChitChatMessage] {

	incrementLamport()	
	
	fileLog.Printf("/ Client %d / Join Chat Request / Lamport %d", localChitChatter.ID, localLamport)
	
	stream1, err := client.JoinChat(context.Background(), &pb.User{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: localLamport})
	if err != nil {
		log.Fatalf("Not working in client 1")
	}
	*isActivePointer = true
	return stream1
}

// LeaveChat function call with lamport increments
func localLeaveChat(client pb.ChitChatServiceClient, localChitChatter ChitChatter, isActivePointer *bool) {

	incrementLamport()

	fileLog.Printf("/ Client %d / Leave Chat Request / Lamport %d", localChitChatter.ID, localLamport)

	client.LeaveChat(context.Background(), &pb.User{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: localLamport})
	*isActivePointer = false
}

// Receives messages as long as the stream is active,
// which is tracked by the isActivePointer
func receiveMessages(msgStream grpc.ServerStreamingClient[pb.ChitChatMessage], isActivePointer *bool, localChitChatter ChitChatter) {
	for *isActivePointer{
		msg, err := msgStream.Recv()
		if err == io.EOF {
			break
		}
	
		syncLamport(msg.Lamport)


		fileLog.Printf("/ Client %d / Received [Client %d]'s Message from Server / Lamport %d", localChitChatter.ID, msg.User.Id, localLamport)
		termLog.Println(msg.Message)
		
	}
}

// Sends a ChitChatMessage to the Server
// Increments the lamport making the request
func localSendMessage(client pb.ChitChatServiceClient, localChitChatter ChitChatter, message string) {
	incrementLamport()

	fileLog.Printf("/ Client %d / Send Message / Lamport %d", localChitChatter.ID, localLamport)

	client.Publish(context.Background(), &pb.ChitChatMessage{User: &pb.User{Id: localChitChatter.ID, Name: localChitChatter.Name, Lamport: localLamport}, Message: message, Lamport: localLamport})
}

// Sends a total of 50 proto ChitChatMessages 
func SendMessageLoop(client pb.ChitChatServiceClient, localChitChatter ChitChatter) {
	for i := 0; i < 20; i++ {

		message, err := rn.GetRandomName("./all.last", &rn.Options{})

		message = fmt.Sprintf("%s", message)
		if err != nil {
			log.Fatalf("Not working in messageLoop")
		}
		localSendMessage(client, localChitChatter, message)

	}
}

func main() {
	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	/* Create two independent loggers, one for logging to system.log and one for logging in terminal */
	fileLog = log.New(file, "",log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) 

	/* Lamport Timestamp */
	localLamport = 0

	/* ChitChatter State */
	isActive := false
	var isActivePointer *bool = &isActive

	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Not working, %v", err)
	}

	defer conn.Close()
	client := pb.NewChitChatServiceClient(conn)

	/* make a user with random name and random id */
	localChitChatter := makeRandomClient()

	/* join chat method call */
	stream1 := localJoinChat(client, localChitChatter, isActivePointer)

	go receiveMessages(stream1, isActivePointer, localChitChatter)

	go SendMessageLoop(client, localChitChatter)

	time.Sleep(30 * time.Second)

	localLeaveChat(client, localChitChatter, isActivePointer)

	for {
	}
}
