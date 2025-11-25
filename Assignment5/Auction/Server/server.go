package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	pb "Auction/grpc"
)

type AuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	ID                int32
	IsLeader          bool
	BestBid           HighestBidder  // The best bid so far (potentially change to the proto message type instead)
	RegisteredClients map[int32]bool // Keep track of registered Clients
	AuctionOngoing    bool
	BackupConnection  pb.AuctionServiceClient
	Connection        *grpc.ClientConn
	Mu                sync.Mutex
}

type HighestBidder struct {
	BidderID  int32
	BidAmount int32
}

func main() {
	fmt.Println("Starting Auction Server...")

	// Get the id and leader from the terminal
	i, _ := strconv.ParseInt(os.Args[1], 10, 32) // parse id to int
	id := int32(i)                               // cast to int32
	isleader, _ := strconv.ParseBool(os.Args[2]) // boolean, am a the leader?

	server := &AuctionServer{
		ID:                id,
		IsLeader:          isleader,
		BestBid:           HighestBidder{BidderID: -1, BidAmount: 0},
		RegisteredClients: make(map[int32]bool),
		AuctionOngoing:    false,
		BackupConnection:  nil,
		Connection:        nil,
		Mu:                sync.Mutex{},
	}

	if server.IsLeader == true { // If I am the leader node
		server.start_server()               // Host the server
		server.start_backup_connection()    // connect to backup server
		go server.StartAndEndBiddingRound() // starts auctions at intervals
	} else { // Else I am the synchronized node (and only does what I am told)
		server.start_backup_server()
		// the backup does NOT start a bidding rounds
	}

	select {} // Keep the server running
}

/***********	Create the Server and Backup Server   ******************/

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
}

/********** Connection to Backup Server **********************/

func (s *AuctionServer) start_backup_connection() {

	// changed grpc.NewClient() to grpc.Dial
	conn, err := grpc.Dial("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Primary Server could not dial to Backup server: %v", err)
		return
	}

	// lock access to server fields to safely update server fields
	s.Mu.Lock()
	// initialize servers Backup Connection
	s.BackupConnection = pb.NewAuctionServiceClient(conn)
	// save the connection for later check ups
	s.Connection = conn
	// unlock mutex when done
	s.Mu.Unlock()

	// Check the connection before proceeding
	if s.Connection_Helper_Method(30, 1) == false {
		fmt.Printf("Could not get the connection to the backup server working \n")
		return
	}

	fmt.Printf("Primary server connection to Backup server created, state: %s \n", conn.GetState())
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

/********** Auction Logic **********************/

/* Starts each bidding round by changing Auction Ongoing to true or false and updating the best bid thereafter  */
func (s *AuctionServer) StartAndEndBiddingRound() {
	for i := 0; i < 5; i++ { // Total items to be bid on before the auction ends

		s.Mu.Lock()
		s.BestBid.BidderID = -1
		s.BestBid.BidAmount = 0 // reset the highest bidder
		// (FIFTH) This should technically also have a grpc call UpdateHighestBid() to the backup server which updates the highest bid so far

		s.AuctionOngoing = true               // Auction round begins
		s.local_update_auction_state("start") // The local rpc handler to update auction state
		s.Mu.Unlock()

		fmt.Printf("Round %d of auction has begun \n", i)
		time.Sleep(10 * time.Second) // Auction round duration

		s.Mu.Lock()
		s.AuctionOngoing = false // Auction round ends
		s.local_update_auction_state("end")
		fmt.Printf("Round %d of auction is over, winning bid: %d by client: %d \n", i, s.BestBid.BidAmount, s.BestBid.BidderID)
		s.Mu.Unlock()
		time.Sleep(6 * time.Second) // Next item to be sold is being prepared (takes 6 seconds)
	}
	fmt.Println("The auction is now over!")
}

// Ensures only the leader does this call
// and that we contact the backup server with a working connection
func (s *AuctionServer) local_update_auction_state(when string) {
	if !s.IsLeader { // only the leader server should be making this call
		return
	}

	// Should only give false if the Backup server failed
	if s.Connection_Helper_Method(5, 2) == false { //looking for connection within 5 seconds
		fmt.Printf("	connection to backup server not active \n")
		return
	}

	res, err := s.BackupConnection.UpdateAuctionState(context.Background(), &pb.AuctionState{Ongoing: s.AuctionOngoing})
	if err != nil || res.Success != true { // We technically can never return success false as the code is now...
		fmt.Printf("	Failed to update Backup server auction state at %s of round: %v \n", when, err)
	}

	// Delete later Homemade ERROR HANDLING:
	if res.Success {
		fmt.Printf("		Backup Server updated at the %s of auction round \n", when)
	}
}

/* Does not properly return the pb.BidResponse to Client (don't know why yet) */
func (s *AuctionServer) Bid(ctx context.Context, bidder *pb.Bidder) (*pb.BidResponse, error) {
	err := errors.New("EXCEPTION") // makes the exception message

	s.Mu.Lock()
	defer s.Mu.Unlock()
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
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return &pb.ResultResponse{
		HighestBid: s.BestBid.BidAmount,
		ItemSold:   !s.AuctionOngoing, // if auction is ongoing then the Item hasn't been sold
	}, nil
}

/***********	Primary Server and Backup Server grpc Calls      ******************/

func (s *AuctionServer) UpdateAuctionState(ctx context.Context, state *pb.AuctionState) (*pb.BackupResponse, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	fmt.Printf("Backup Server, old Auction ongoing: %t new Auction ongoing: %t \n", s.AuctionOngoing, state.Ongoing)
	s.AuctionOngoing = state.Ongoing
	return &pb.BackupResponse{Success: true}, nil
}

// TODO: make logic of the two missing grpc calls between the leader and backup server
/*
func (s *AuctionServer) UpdateRegisteredClient(ctx context.Context, *pb.Bidder) (*pb.BackupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateRegisteredClient not implemented")
}
func (s *AuctionServer) UpdateHighestBid(ctx context.Context, *pb.Bidder) (*pb.BackupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHighestBid not implemented")
}
*/

/*******       Connection Helper Methods        ***********************************/

/* Below method for checking the connection status, was made in cooperation with Gemini.
Meaning we did make changes and other ressources as well	*/

func (s *AuctionServer) Connection_Helper_Method(timeout int, num int) bool {
	conn := s.Connection

	// in case Backup Server is not made within 30 seconds, give up on it and continue
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(int(timeout))*time.Second)
	defer cancel()

	// Save our current state such that we can check if it is READY later
	currentState := conn.GetState()

	for {
		if currentState == connectivity.Ready { // Connection is ready
			fmt.Printf("	Connection is now ready for use state: %v \n", conn.GetState())
			break
		}

		// Below line will block until there is a state change or context times out
		if !conn.WaitForStateChange(ctx, currentState) {
			fmt.Printf("	Primary server timeout error, state: %v, call number: %d \n", conn.GetState(), num)
			return false
		}

		currentState = conn.GetState() // Update our current state

		//could be deleted later
		if currentState == connectivity.Shutdown { // Connection has been shutdown
			fmt.Printf("	Connection was shut down... but we want to continue anyway (but can't just yet) \n")
			break
		}

		// Delete error handling later
		fmt.Printf("State changed to %s, wait for a state change\n", currentState.String())
	}
	return true
}
