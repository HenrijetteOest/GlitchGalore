package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
	"strconv"
	"os"

	"google.golang.org/grpc"

	pb "Auction/grpc"
)

type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	ID       			int32
	IsLeader			bool
	//SynchronousID 	int32 // For Server backups
	//OtherServers 		map[int32]	//key: id //value: status i bidder?	// keep track of other servers //for server backup
	BestBid           	HighestBidder  // The best bid so far (potentially change to the proto message type instead)
	RegisteredClients 	map[int32]bool // Keep track of registered Clients
	AuctionOngoing    	bool
}

type HighestBidder struct {
	BidderID  int32
	BidAmount int32
}

func main() {
	fmt.Println("Starting Auction Server...")

	// Get the id and leader from the terminal
	i, _ := strconv.ParseInt(os.Args[1], 10, 32) 	 // parse id to int
	id := int32(i)									 // cast to int32
	isleader, _ := strconv.ParseBool(os.Args[2])	 // boolean, am a the leader?

	server := &AuctionServer{
		ID:                id,
		IsLeader:          isleader,
		BestBid:           HighestBidder{BidderID: -1, BidAmount: 0},
		RegisteredClients: make(map[int32]bool),
		AuctionOngoing:    false,
	}

	if (server.IsLeader == true) {  // If I am the leader node
		server.start_server()
	} else  { 						// or I am the synchronized node
		server.start_backup_server()
		
		// how do you keep up to date with leader
	} 
	
	/* Help to understand how to run the server with leader or not
					and  what true and false mean
	//Leader_bool = true
	//notLeader_bool = false
	
	go run ./server 1 1 (true)
	go run ./server 2 0 (false)
	*/

	/* Pseudokode
	if (crash happens...) {
		find new leader
	}	
	*/
	
	go server.StartAndEndBiddingRound() // starts auctions at intervals

	select {}
}

// Starts the backup server (serve() is a goroutine )
func (s *AuctionServer) start_backup_server() {
	lis, err := net.Listen("tcp", ":6060")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, s)
	fmt.Printf("Backup Auction Server %d listening on port 6060\n", s.ID)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve on the backup server: %v", err)
		}
	}()
}

// Starts the server (serve() is a goroutine )
func (s *AuctionServer) start_server() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, s)
	fmt.Printf("Auction Server %d listening on port 5050\n", s.ID)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve on the primary server: %v", err)
		}
	}()

	/*
	
	Establish client connection to Backup Server
		Backup running on port 6060
		We want to establish a connection to that backup
	
	conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewAuctionServiceClient(conn)


	*/
}

/* Starts each bidding round by changing Auction Ongoing to true or false and updating the best bid thereafter  */
func (s *AuctionServer) StartAndEndBiddingRound() {
	for i := 0; i < 1; i++ {
		// Total items to be bid on before the auction ends
		s.BestBid.BidderID = -1
		s.BestBid.BidAmount = 0 // reset the highest bidder

		s.AuctionOngoing = true // Auction round begins
		// Third grpc Call: Auction change
		// WAIT for response

		fmt.Printf("	Round %d of auction has begun \n", i)
		time.Sleep(10 * time.Second) // Auction round duration

		s.AuctionOngoing = false // Auction round ends
		// Fourth grpc call: Auction change
		// WAIT for response

		fmt.Printf("	Round %d of auction is over, winning bid: %d by client: %d \n", i, s.BestBid.BidAmount, s.BestBid.BidderID)
		time.Sleep(6 * time.Second) // Next item to be sold is prepared	// time.Duration(int32(rand.Intn(5)))
	}
	fmt.Println("The auction is now over!")
}

/* Does not properly return the pb.BidResponse to Client (don't know why yet) */
func (s *AuctionServer) Bid(ctx context.Context, bidder *pb.Bidder) (*pb.BidResponse, error) {
	err := errors.New("EXCEPTION") // makes the exception message

	if s.RegisteredClients[bidder.Id] != true {
		s.RegisteredClients[bidder.Id] = true
		// First Update: add grpc call to another server to update the database
		// WAIT for response before proceeding
	}

	if bidder.Bid > s.BestBid.BidAmount && s.AuctionOngoing == true {
		s.BestBid.BidAmount = bidder.Bid
		s.BestBid.BidderID = bidder.Id
		// Second Update: grpc call to db
		// WAIT for response before proceeding

		// return success or err (exception)
		return &pb.BidResponse{Status: "SUCCESS"}, err
	}

	return &pb.BidResponse{Status: "FAIL"}, err //used to be nil
}

/* Returns the highest bid and whether the item has been sold yet or not   */
func (s *AuctionServer) Result(ctx context.Context, empty *pb.Empty) (*pb.ResultResponse, error) {
	return &pb.ResultResponse{
		HighestBid: s.BestBid.BidAmount,
		ItemSold:   !s.AuctionOngoing, // if auction is ongoing then the Item hasn't been sold
	}, nil
}
