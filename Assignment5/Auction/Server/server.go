package main

import (
	"context"
	"fmt"
	"time"
	"math/rand"
	"errors"
	"net"
	"log"

	"google.golang.org/grpc"

	pb "Auction/grpc"

)


type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	ID 					int32
	LeaderID 			int32
	//SynchronousID 	int32 // For Server backups
	//OtherServers 		map[int32]	//key: id //value: status i bidder?	// keep track of other servers //for server backup
	BestBid 			HighestBidder  // The best bid so far
	RegisteredClients   map[int32]bool // Keep track of registered Clients
	AuctionOngoing		bool
}

type HighestBidder struct {
	BidderID	int32
	BidAmount	int32
}

func main() {
	fmt.Println("Starting Auction Server...")
	server := &AuctionServer{
		ID:       1,
		LeaderID: 1,
		BestBid:  HighestBidder{BidderID: -1, BidAmount: 0},
		RegisteredClients: make(map[int32]bool),
		AuctionOngoing: false,
	}
	server.start_server()

	go server.StartAndEndBiddingRound() // starts auctions at intervals

	select {}
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

	go func(){
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
}

/* Starts each bidding round by changing Auction Ongoing to true or false and updating the best bid thereafter  */
func (s *AuctionServer) StartAndEndBiddingRound() {
	for i := 1; i < 6; i++ { 	
			// Total items to be bid on before the auction ends
		s.BestBid.BidderID = -1
		s.BestBid.BidAmount = 0		// reset the highest bidder
		
		s.AuctionOngoing = true			// Auction round begins
		fmt.Printf("	Round %d of auction has begun \n", i)
		time.Sleep(10 * time.Second)	// Auction round duration
		
		s.AuctionOngoing = false		// Auction round ends
		fmt.Printf("	Round %d of auction is over, winning bid: %d by client: %d \n", i, s.BestBid.BidAmount, s.BestBid.BidderID)
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)	// Next item to be sold is prepared
	}
	fmt.Println("The auction is now over!")
}

/* Does not properly return the pb.BidResponse to Client (don't know why yet) */
func (s *AuctionServer) Bid(ctx context.Context, bidder *pb.Bidder) (*pb.BidResponse, error) {
	err := errors.New("EXCEPTION")	// makes the exception message
	
	if s.RegisteredClients[bidder.Id] != true {
		s.RegisteredClients[bidder.Id] = true
		// add grpc call to another server to update the database
	}
	
	if bidder.Bid > s.BestBid.BidAmount && s.AuctionOngoing == true{
		s.BestBid.BidAmount = bidder.Bid
		s.BestBid.BidderID = bidder.Id
		// return success or err (exception)
		return &pb.BidResponse{Status: "SUCCESS"}, err 
	} 
	 
	return &pb.BidResponse{Status: "FAIL"}, err //used to be nil
}

/* Returns the highest bid and whether the item has been sold yet or not   */
func (s *AuctionServer) Result(ctx context.Context, empty *pb.Empty) (*pb.ResultResponse, error) {
	return &pb.ResultResponse{
		HighestBid: s.BestBid.BidAmount,	
		ItemSold: !s.AuctionOngoing,		// if auction is ongoing then the Item hasn't been sold
	}, nil	
}

