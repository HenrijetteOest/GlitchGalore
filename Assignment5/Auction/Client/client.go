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
	/********* Create the log file  *************/
	file, e := os.OpenFile("system.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if e != nil {
		log.Fatalf("Failed to open log file: %v", e)
	}
	defer file.Close()

	// Create two independent loggers
	fileLog = log.New(file, "", log.Ldate|log.Ltime)
	termLog = log.New(os.Stdout, "", 0) // plain chat-style output


	/********* Establish connection to primary server *************/
	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewAuctionServiceClient(conn)


	/********* Create Auction Client  *************/
	i, err := strconv.ParseInt(os.Args[1], 10, 32) // ParseInt always returns a int64, but we specify it should only take up 32 bits
	bidderId := int32(i)                           // cast to int32

	LocalBidder := &AuctionClient{
		ID:     bidderId,                 // Remember to give clients different ids in the terminal
		My_Bid: int32(rand.Intn(99) + 1), // Random start bid, between 1 and 100
		Client: client,
	}

	fileLog.Printf("Starting Auction Client %d...\n", LocalBidder.ID)
	termLog.Printf("Starting Auction Client %d...\n", LocalBidder.ID)

	/********* Starting Bidding logic  *************/
	go LocalBidder.MakeBid()

	// Keep the client running
	select {}
}

// Gets the status on the item to bid on, if it has been sold, then reset bidding price
// If the item has not been sold, then compare the highest bid to our bid, and increment
// our bid such that it is greater and place a bid
// Sleeps for 1-4 seconds after placing a bid to make the log easier to read (and create randomness) 
func (s *AuctionClient) MakeBid() {
	for {
		// Get current result of the auction 
		res, err := s.Client.Result(context.Background(), &pb.Empty{})
		if err != nil {

			fileLog.Printf("Client: %d | asked for result from primary server: Lost Connection: %v. Will attempt to connect to backup server \n", s.ID, err)
			termLog.Printf("I asked for result from the primary server: Lost Connection to primary server: %v will attempt to connect to backup server  \n", err)
			
			// Create connection to backup server
			s.ReestablishConnection()

			// Resend failed result query
			res, err = s.Client.Result(context.Background(), &pb.Empty{})
			if err != nil {
				fileLog.Printf("Client: %d | failed to switch to backup after primary crashed (request for result)\n", s.ID)
				termLog.Printf("I failed to switch to backup after primary crashed (request for result)\n")
			}

			termLog.Printf("Getting result (highest bid) from from backup server: %v \n", res)
		}

		// If the auction is over, then reset price
		// else compare bid amounts and place a new bid
		if res.ItemSold == true { 
			// Resets bid field with a new random start bid, between 1 and 100 for next auction
			s.My_Bid = int32(rand.Intn(99) + 1) 

		} else { 
			// if our bid is too low increment it by 1-50
			if res.HighestBid >= s.My_Bid {
				s.My_Bid = res.HighestBid + int32(rand.Intn(49)+1) 
				termLog.Printf("my bid is increasing bid to %d\n", s.My_Bid)
			}
			s.PlaceBid()
		}
		// Waits for 1-4 seconds before placing next bid
		time.Sleep(time.Duration(int32(rand.Intn(3)+1)) * time.Second)
	}
}
// Handles successful or unsuccessful bid to Leader server
func (s *AuctionClient) PlaceBid() {
	res, err := s.Client.Bid(context.Background(), &pb.Bidder{Bid: s.My_Bid, Id: s.ID})
	if err != nil { // Handle primary server crash
		s.ReestablishConnection() // establish connection to the backup server
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

// Tries to establish connection to the backup server 
func (s *AuctionClient) ReestablishConnection() {

	conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect to backup server: %v", err)
	}

	s.Client = pb.NewAuctionServiceClient(conn)
	fileLog.Printf("Client: %d | succesfully connected to backup server", s.ID)
	termLog.Printf("I succesfully connected to backup server")
}