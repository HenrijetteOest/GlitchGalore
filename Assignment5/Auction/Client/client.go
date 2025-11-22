package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "Auction/grpc"
)

type AuctionClient struct {
	// client pb.AuctionServiceClient
	ID     int32
	My_Bid int32
}

func main() {
	fmt.Println("Starting Auction Client...")

	conn, err := grpc.NewClient("localhost:5050", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client could not connect: %v", err)
	}

	defer conn.Close()
	client := pb.NewAuctionServiceClient(conn)

	i, err := strconv.ParseInt(os.Args[1], 10, 32) 	// ParseInt always returns a int64, but we specify it should only take up 32 bits
	bidderId := int32(i)	// cast to int32


	LocalBidder := &AuctionClient{
		ID:     bidderId,                 // Remember to give clients different ids in the terminal
		My_Bid: int32(rand.Intn(99) + 1), // Random start bid, between 1 and 100
	}

	go PlaceBid(client, LocalBidder)

	// Keep the client running
	select {}
}

// bidding logic
/* Only resets it's bidding price if it asks and the bidding is over  */
func PlaceBid(client pb.AuctionServiceClient, LocalBidder *AuctionClient) {
	for i := 0; i < 30; i++ {
		// Get current status of the action
		res, err := client.Result(context.Background(), &pb.Empty{})
		if err != nil {
			fmt.Printf("Failed to get result from grpc call: %v", err)
		}

		if res.ItemSold == true { // If the auction is over, then reset price
			LocalBidder.My_Bid = int32(rand.Intn(99) + 1) // Resest bid field with a new random start bid, between 1 and 100

		} else { // Auction round is still ongoing!
			/*if res.HighestBid == 0 {
				LocalBidder.My_Bid = int32(rand.Intn(99) + 1)
			}*/
			if res.HighestBid >= LocalBidder.My_Bid {
				// Increase our bid price and then bid
				LocalBidder.My_Bid = res.HighestBid + int32(rand.Intn(49)+1) // increment bid by 1-50
				fmt.Printf("	Client %d is increasing bid to %d\n", LocalBidder.ID, LocalBidder.My_Bid)
			}
			BidCall(client, LocalBidder) // Do the grpc call
		}
		time.Sleep(time.Duration(int32(rand.Intn(5))) * time.Second)
	}
}

func BidCall(client pb.AuctionServiceClient, LocalBidder *AuctionClient) {
	// res, err :=
	// Bid() should return some things, but that doesn't work at the moment
	client.Bid(context.Background(), &pb.Bidder{
		Bid: LocalBidder.My_Bid,
		Id:  LocalBidder.ID,
	})

	fmt.Printf("Client %d placed a bid of %d\n", LocalBidder.ID, LocalBidder.My_Bid)

	/*
		if err == nil {
			fmt.Println("something went wrong in BidCall")
		}

		if res.Status == "SUCCESS" {
			fmt.Printf("SUCCESS: Client %d placed a bid of %d\n", LocalBidder.ID, LocalBidder.My_Bid)
		} else if res.Status == "FAIL" {
			fmt.Printf("FAIL: Client %d failed to place a bid \n", LocalBidder.ID)
		}*/
}

/* FREEZER */
// Place a bid
/*
	Case 1: 	current bid < our_bid  &&   item NOT sold	-->  just bid
	Case 2: 	current bid > our_bid  && 	item NOT sold	-->  increase our price and then bid
	Case 3:		current bid < our_bid  && 	item IS sold	-->  lower our start price
	Case 4:  	current bid > our_bid  && 	item IS sold	-->  lower our start price

*/
