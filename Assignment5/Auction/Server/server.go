package main

import (
	"context"
	"fmt"
	"time"
	"math/rand"
	"errors"

	"google.golang.org/grpc"

	pb "Auction/grpc"

)


type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	ID 					int32
	LeaderID 			int32
	//SynchronousID 	int32
	//OtherServers 		map[int32]	//key: id //value: status i bidder?	// Track of other servers
	BestBid 			HighestBidder	// Keep track of registered Client
	RegisteredClients   [int32]bool 
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

	go StartAndEndBiddingRound() // starts auctions at intervals
}

func (s *AuctionServer) start_server() {
	lis, err := net.Listen("tcp", ":5050")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, s)
	log.Printf("Auction Server %d listening on port 5050", s.ID)

	go func(){
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
}

func (s *AuctionServer) StartAndEndBiddingRound() {
	for int i := 0; i < 5; i++ { 		// Total items to be bid on before the auction ends
		s.HighestBidder(BidderID: -1, BidAmount: 0)		// reset the highest bidder
		s.AuctionOngoing = true			// Auction round begins
		log.Printf("	Round %d of auction has begun \n", i)
		time.Sleep(10 * time.Second)	// Auction round duration
		s.AuctionOngoing = false		// Auction round ends
		log.Printf("	Round %d of auction is over \n", i)
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)	// Next item to be sold is prepared
	}
	log.Println("The auction is now over!")
}

func (s *AuctionServer) Bid(ctx context.Context, bidder *pb.Bidder) (*pb.BidResponse, error) {
	err := errors.New("EXCEPTION")	// makes the exception message
	
	if s.registeredClients[bidder.my_id] != true {
		s.registeredClients[bidder.my_id] = true
		// grpc call to another server to update the database
	}
	
	if bidder.my_bid > s.HighestBidder.BidAmount && s.AuctionOngoing == true{
		s.HighestBidder.BidAmount = bidder.my_bid
		s.HighestBidder.BidderID = bidder.my_id
		// return success or err (exception)
		return &pb.BidResponse{status: "SUCCESS"}, err 
	} 
	 
	return &pb.BidResponse{status: "FAIL"}, err //used to be nil
}

func (s *AuctionServer) Result(ctx context.Context, *pb.Empty) (*pb.ResultResponse, error) {
	return &pb.ResultResponse{
		higest_bid: s.HighestBidder.BidAmount,	
		item_sold: !s.AuctionOngoing,				// if auction is ongoing then the Item hasn't been sold
	}, nil	
}


/*****  FREEZER   ******/