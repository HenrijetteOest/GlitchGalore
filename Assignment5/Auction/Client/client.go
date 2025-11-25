package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "Auction/grpc"
)

type AuctionClient struct {
	ID     int32
	My_Bid int32
	Client pb.AuctionServiceClient
}

var fileLog *log.Logger
var termLog *log.Logger

func main() {

	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	// Create two independent loggers
	fileLog = log.New(file, "", log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) // plain chat-style output

	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewAuctionServiceClient(conn)

	i, err := strconv.ParseInt(os.Args[1], 10, 32) // ParseInt always returns a int64, but we specify it should only take up 32 bits
	bidderId := int32(i)                           // cast to int32

	LocalBidder := &AuctionClient{
		ID:     bidderId,                 // Remember to give clients different ids in the terminal
		My_Bid: int32(rand.Intn(99) + 1), // Random start bid, between 1 and 100
		Client: client,
	}

	fileLog.Printf("Starting Auction Client %d...\n", LocalBidder.ID)
	termLog.Printf("Starting Auction Client %d...\n", LocalBidder.ID)

	go LocalBidder.MakeBid()

	// Keep the client running
	select {}
}

// bidding logic
/* Only resets it's bidding price if it asks and the bidding is over  */
func (s *AuctionClient) MakeBid() {
	for {
		// Get current status of the action
		res, err := s.Client.Result(context.Background(), &pb.Empty{})
		if err != nil {

			fileLog.Printf("Client: %d | asked for result: Lost Connection to primary server: %v. Will attempt to connect to backup server \n", s.ID, err)
			termLog.Printf("I asked for result: Lost Connection to primary server: %v will attempt to connect to backup server  \n", err)
			//create connection to backup server
			s.ReestablishConnection()

			//resend failed result query
			res, err = s.Client.Result(context.Background(), &pb.Empty{})
			if err != nil {
				fileLog.Printf("Client: %d | failed to switch to backup after primary crashed (request for result)\n", s.ID)
				termLog.Printf("I failed to switch to backup after primary crashed (request for result)\n")
			}

			termLog.Printf("error handling for request result from server: %v \n", res)
		}

		if res.ItemSold == true { // If the auction is over, then reset price
			s.My_Bid = int32(rand.Intn(99) + 1) // Resest bid field with a new random start bid, between 1 and 100 for next auction

		} else { // Auction round is still ongoing!
			if res.HighestBid >= s.My_Bid {
				// Increase our bid price and then bid
				s.My_Bid = res.HighestBid + int32(rand.Intn(49)+1) // increment bid by 1-50
				termLog.Printf("my bid is increasing bid to %d\n", s.My_Bid)
			}
			s.PlaceBid() // Do the grpc call
		}
		time.Sleep(time.Duration(int32(rand.Intn(3)+1)) * time.Second)
		//time.Sleep(1 * time.Second)
	}
}

func (s *AuctionClient) PlaceBid() {
	res, err := s.Client.Bid(context.Background(), &pb.Bidder{Bid: s.My_Bid, Id: s.ID})
	if err != nil { // Handle primary server crash
		s.ReestablishConnection()
		s.PlaceBid() // If things went wrong, repeat last grpc call
	}

	switch res.Status {
	case "SUCCESS":
		fileLog.Printf("Client: %d | placed a succesful bid of %d \n", s.ID, s.My_Bid)
		termLog.Printf("I placed a succesful bid of %d \n", s.My_Bid)
	case "FAIL":
		fileLog.Printf("Client: %d's bid of %d was rejected \n", s.ID, s.My_Bid)
		termLog.Printf("My bid of %d was rejected\n", s.My_Bid)
	}
}

func (s *AuctionClient) ReestablishConnection() {

	conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect to backup server: %v", err)
	}

	s.Client = pb.NewAuctionServiceClient(conn)
	fileLog.Printf("Client: %d | succesfully connected to backup server", s.ID)
	termLog.Printf("I succesfully connected to backup server")
}

/* FREEZER */

/*
// Place a bid cases
	Case 1: 	current bid < our_bid  &&   item NOT sold	-->  just bid
	Case 2: 	current bid > our_bid  && 	item NOT sold	-->  increase our price and then bid
	Case 3:		current bid < our_bid  && 	item IS sold	-->  lower our start price
	Case 4:  	current bid > our_bid  && 	item IS sold	-->  lower our start price

*/
