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
	//"google.golang.org/grpc/connectivity"

	pb "Auction/grpc"

)

type AuctionClient struct {
	// client pb.AuctionServiceClient
	ID     int32
	My_Bid int32
	Client 	 pb.AuctionServiceClient
	Conn	*grpc.ClientConn
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
		Client:	client,
		Conn:	conn,
	}

	fileLog.Printf("Starting Auction Client %d...\n",LocalBidder.ID) 
	termLog.Printf("Starting Auction Client %d...\n", LocalBidder.ID) 

	go LocalBidder.MakeBid()

	// Keep the client running
	select {}
}

// bidding logic
/* Only resets it's bidding price if it asks and the bidding is over  */
func (s *AuctionClient) MakeBid() {
	for { // i := 0; i < 100; i++
		// Get current status of the action
		res, err := s.Client.Result(context.Background(), &pb.Empty{})
		if err != nil {
			fileLog.Printf("Client %d: Failed to get result from grpc call: %v will attempt to connect to backup server \n", s.ID, err)
			termLog.Printf("Client %d: Failed to get result from grpc call: %v will attempt to connect to backup server  \n", s.ID, err)

			conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Client could not connect to backup server: %v", err)
			}
			defer conn.Close()
			s.Conn = conn
			s.Client = pb.NewAuctionServiceClient(conn)
			fileLog.Printf("Client %d should have connected to backup server", s.ID)
			termLog.Printf("Client %d should have connected to backup server", s.ID)
	
			res, err = s.Client.Result(context.Background(), &pb.Empty{})
			if err != nil {
				fileLog.Printf("Failed to connect to backup server \n")
				termLog.Printf("Failed to connect to backup server \n")
			}	
			fileLog.Printf("res: %v\n", res)
			termLog.Printf("res: %v \n", res)
		}

		if res.ItemSold == true { // If the auction is over, then reset price
			s.My_Bid = int32(rand.Intn(99) + 1) // Resest bid field with a new random start bid, between 1 and 100

		} else { // Auction round is still ongoing!
			/*if res.HighestBid == 0 {
				LocalBidder.My_Bid = int32(rand.Intn(99) + 1)
			}*/
			if res.HighestBid >= s.My_Bid {
				// Increase our bid price and then bid
				s.My_Bid = res.HighestBid + int32(rand.Intn(49)+1) // increment bid by 1-50
				//fileLog.Printf("Client %d is increasing bid to %d\n", LocalBidder.ID, LocalBidder.My_Bid)
				termLog.Printf("Client %d is increasing bid to %d\n", s.ID, s.My_Bid)
			}
			s.PlaceBid() // Do the grpc call
		}
		//time.Sleep(time.Duration(int32(rand.Intn(3)+1)) * time.Second)
		time.Sleep(1 * time.Second)
	}
}

func (s *AuctionClient) PlaceBid() {  //client pb.AuctionServiceClient, LocalBidder *AuctionClient
	/*
	client.Bid(context.Background(), &pb.Bidder{
		Bid: LocalBidder.My_Bid,
		Id:  LocalBidder.ID,
	})*/	

	res, err := s.Client.Bid(context.Background(), &pb.Bidder{ Bid: s.My_Bid, Id:  s.ID,})
	if err != nil { 	// Handle primary server crash
		conn, err := grpc.NewClient("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("Client could not connect to backup server: %v", err)
		}
		defer conn.Close()
		s.Conn = conn
		s.Client = pb.NewAuctionServiceClient(conn)
		fileLog.Printf("Client %d should have connected to backup server", s.ID)
		s.PlaceBid() // If things went wrong, repeat last grpc call
	} 

	fileLog.Printf("Client %d | res is: %s \n", s.ID, res.Status)

	if res.Status == "SUCCESS" {
		fileLog.Printf("SUCCESS: Client %d placed a bid of %d \n", s.ID, s.My_Bid)
		termLog.Printf("SUCCESS: Client %d placed a bid of %d \n", s.ID, s.My_Bid)
	}
}

/* FREEZER */

/*
// Place a bid cases
	Case 1: 	current bid < our_bid  &&   item NOT sold	-->  just bid
	Case 2: 	current bid > our_bid  && 	item NOT sold	-->  increase our price and then bid
	Case 3:		current bid < our_bid  && 	item IS sold	-->  lower our start price
	Case 4:  	current bid > our_bid  && 	item IS sold	-->  lower our start price

*/

/*
// Below code hinders the client from sending grpc calls to the Server.
// Don't know why yet
// FROM SERVER
// in case Backup Server is not made within 30 seconds, give up on it and continue
ctx, cancel := context.WithTimeout(context.Background(), 30 * time.Second)
defer cancel()

// Save our current state such that we can check if it is READY later
currentState := conn.GetState()

for {
	// Below line will block until there is a state change or context times out
	if !conn.WaitForStateChange(ctx, currentState)  {
		fmt.Printf("Client: Timeout error in waiting for state change, state: %s \n", conn.GetState())
		return
	}
	currentState = conn.GetState() 				// Update our current state

	if currentState == connectivity.Ready {		// Connection is ready
		fmt.Printf("Connection is now ready for use %v \n", conn.GetState())
		break
	}

	if currentState == connectivity.Shutdown {   // Connection has been shutdown
		fmt.Printf("Connection was shut down... but we want to continue anyway (but can't just yet) \n")
		break
	}

	// Delete error handling later
	fmt.Printf("State changed to %s, wait for a state change\n", currentState.String())
}*/
