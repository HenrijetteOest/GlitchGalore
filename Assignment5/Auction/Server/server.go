package main

import (
	"context"
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
	AuctionRound      int32
	BestBid           HighestBidder  
	RegisteredClients map[int32]bool 
	AuctionOngoing    bool
	BackupConnection  pb.AuctionServiceClient
	Connection        *grpc.ClientConn
	Mu                sync.Mutex
}

type HighestBidder struct {
	BidderID  int32
	BidAmount int32
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

	fileLog.Println("Starting Auction Server...")

	// Get the id and leader from the terminal
	i, _ := strconv.ParseInt(os.Args[1], 10, 32) // parse id to int
	id := int32(i)                               // cast to int32
	isleader, _ := strconv.ParseBool(os.Args[2]) // boolean, am I the leader?

	server := &AuctionServer{
		ID:                id,
		IsLeader:          isleader,
		AuctionRound:      0,
		BestBid:           HighestBidder{BidderID: -1, BidAmount: 0},
		RegisteredClients: make(map[int32]bool),
		AuctionOngoing:    false,
		BackupConnection:  nil,
		Connection:        nil,
		Mu:                sync.Mutex{},
	}

	// Either make the primary server
	// or the backup server
	if server.IsLeader == true { 
		server.start_server()               // Host the server
		server.start_backup_connection()    // connect to backup server
		go server.StartAndEndBiddingRound() // starts auctions at intervals
	} else {  
		server.start_backup_server()
		go server.StartAndEndBiddingRound()
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
	fileLog.Printf("Auction Server %d listening on port 5050\n", s.ID)
	termLog.Printf("Auction Server %d listening on port 5050\n", s.ID)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve on the primary server: %v", err)
		}
	}()
}

/********** Connection to Backup Server **********************/

// Makes a client conn for the calling function (should be the primary server)
func (s *AuctionServer) start_backup_connection() {
	conn, err := grpc.Dial("localhost:6060", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Primary Server could not dial to Backup server: %v", err)
		return
	}

	// lock access to server fields to safely update server fields
	s.Mu.Lock()
	// initialize servers Backup Connection
	s.BackupConnection = pb.NewAuctionServiceClient(conn)
	// save the connection for later check ups (the connection_helper_method needs this connection to work)
	s.Connection = conn
	// unlock mutex when done
	s.Mu.Unlock()

	// Check the connection before proceeding
	if s.Connection_Helper_Method(15, 1) == false {
		fileLog.Printf("Could not get the connection to the backup server working \n")
		termLog.Printf("Could not get the connection to the backup server working \n")
		return
	}
	fileLog.Printf("Server: %d | Primary server connection to Backup server created, state: %s \n", s.ID, conn.GetState())
	termLog.Printf("Primary server connection to Backup server created, state: %s \n", conn.GetState())
}

// Starts the backup server (serve() is a goroutine )
func (s *AuctionServer) start_backup_server() {
	lis, err := net.Listen("tcp", ":6060")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, s)
	termLog.Printf("Backup Auction Server %d listening on port 6060\n", s.ID)
	fileLog.Printf("Backup Auction Server %d listening on port 6060\n", s.ID)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve on the backup server: %v", err)
		}
	}()
}

/********** Auction Logic **********************/

/* Starts each bidding round by changing Auction Ongoing to true or false and updating the best bid thereafter  */
func (s *AuctionServer) StartAndEndBiddingRound() {
	// The backup server will be stuck in this loop as long as it is not the leader
	// Once the primary server crashes, there will at most pass 4 seconds before backup proceeds
	for !s.IsLeader {
		time.Sleep(4 * time.Second)
	}

	for s.AuctionRound < 5 { // Total items to be bid on before the auction ends
		s.Mu.Lock()

		/********* reset the highest bidder *************/

		s.BestBid.BidderID = -1
		s.BestBid.BidAmount = 0

		/********* Update the backup server *************/
		// Below if statement is only for the Primary server (with a backup)
		if s.Connection_Helper_Method(5, 2) == false && s.Connection != nil { //looking for connection within 5 seconds
			fileLog.Printf("Server: %d | connection to backup server not active: StartAndEndBiddingRound \n", s.ID)
			termLog.Printf("Server: %d | connection to backup server not active: StartAndEndBiddingRound \n", s.ID)
		} else if s.Connection != nil {
			fileLog.Printf("Primary: reset auction values, bid = %d and bidderId = %d  \n", s.BestBid.BidderID, s.BestBid.BidAmount)
			termLog.Printf("Primary: reset auction values, bid = %d and bidderId = %d  \n", s.BestBid.BidderID, s.BestBid.BidAmount)
			res, err := s.BackupConnection.UpdateHighestBid(context.Background(), &pb.Bidder{Bid: s.BestBid.BidderID, Id: s.BestBid.BidAmount})
			if err != nil || res.Success != true { // We technically can never return success false as the code is now...
				fileLog.Printf("Failed to update Highest Bid in Backup server \n")
				termLog.Printf("Failed to update Highest Bid in Backup server \n")
			}
		}

		/********* Start the auction round *************/
		s.AuctionOngoing = true               // Auction round begins
		s.local_update_auction_state("start") // The local rpc handler to update auction state
		s.Mu.Unlock()

		fileLog.Printf("------ Round %d of auction has begun ------------------------------------------ \n", s.AuctionRound)
		termLog.Printf("------ Round %d of auction has begun ------------------------------------------ \n", s.AuctionRound)
		time.Sleep(10 * time.Second) // Auction round duration

		/********* End the auction round *************/
		s.Mu.Lock()
		fileLog.Printf("------ Round %d of auction is over: winning bid: %d by client: %d  ------------ \n", s.AuctionRound, s.BestBid.BidAmount, s.BestBid.BidderID)
		termLog.Printf("------ Round %d of auction is over: winning bid: %d by client: %d  ------------ \n", s.AuctionRound, s.BestBid.BidAmount, s.BestBid.BidderID)
		s.AuctionRound++
		s.AuctionOngoing = false // Auction round ends
		s.local_update_auction_state("end")
		s.Mu.Unlock()
		time.Sleep(3 * time.Second) // Next item to be sold is being prepared (takes 3 seconds)

	}
	fileLog.Printf("------ The auction is now over! ----------------------------------------------- \n")
	termLog.Printf("------ The auction is now over! ----------------------------------------------- \n")
}

// Updates the backup up server with the change in auction state.
// Should only be called by the leader and a server with a backup connection
func (s *AuctionServer) local_update_auction_state(when string) {
	// ensures only the leader server should be making this call
	if !s.IsLeader { 
		return
	}

	// Below if-else statement helps the primary server check the client connection to the back up server 
	// AND has the backup server exit this method. As it should not try and update its own non-existing backup server
	if s.Connection_Helper_Method(5, 3) == false && s.Connection != nil { //looking for connection within 5 seconds
		fileLog.Printf("Server: %d | connection to backup server not active: local_update_auction_state  \n", s.ID)
		termLog.Printf("Server: %d | connection to backup server not active: local_update_auction_state \n", s.ID)
		return
	} else if s.Connection == nil {
		return
	}

	res, err := s.BackupConnection.UpdateAuctionState(context.Background(), &pb.AuctionState{Ongoing: s.AuctionOngoing, AuctionRound: s.AuctionRound})
	if err != nil || res.Success != true { // We technically can never return success == false as the code is now
		fileLog.Printf("Server: %d | Failed to update Backup server auction state at %s of round: %v \n", s.ID, when, err)
		termLog.Printf("Server: %d | Failed to update Backup server auction state at %s of round: %v \n", s.ID, when, err)
	}

	// Homemade ERROR HANDLING:
	if res.Success {
		fileLog.Printf("Server: %d | Backup Server updated at the %s of auction round \n", s.ID, when)
		termLog.Printf("Backup Server updated at the %s of auction round \n", when)
	}
}

// grpc call between Auction client and server (the leader server)
// registers a cliet to the map of clients (if it is the first call)
// places a bid if the bid is valid and returns success or fail thereafter
func (s *AuctionServer) Bid(ctx context.Context, bidder *pb.Bidder) (*pb.BidResponse, error) {

	s.Mu.Lock()
	defer s.Mu.Unlock()

	if s.IsLeader == false {
		s.IsLeader = true
		fileLog.Println("!!!-- PRIMARY SERVER CRASH. BACKUP HAS BEEN PROMOTED TO PRIMARY ----!!!")
	}

	// Register the client if they have not already been registered
	if s.RegisteredClients[bidder.Id] != true {
		s.RegisteredClients[bidder.Id] = true

		// Below if-else statement is only relavant for the primary server (as it regards updating the backup server)
		if s.Connection_Helper_Method(5, 4) == false && s.Connection != nil { //looking for connection within 5 seconds
			fileLog.Printf("Server: %d | connection to backup server not active: registered client \n", s.ID)
			termLog.Printf("Server: %d | connection to backup server not active: registered client \n", s.ID)
		} else if s.Connection != nil{
			fileLog.Printf("Server: %d | Updated map in backup: %v", s.ID, s.RegisteredClients)
			termLog.Printf("Server: %d | Updated map in backup: %v", s.ID, s.RegisteredClients)
			res, err := s.BackupConnection.UpdateRegisteredClient(context.Background(), &pb.Bidder{Bid: bidder.Bid, Id: bidder.Id})
			if err != nil || res.Success != true { // We technically can never return success false as the code is now...
				fileLog.Printf("Server: %d | Failed to update Backup server registered client\n", s.ID)
				termLog.Printf("Server: %d | Failed to update Backup server registered client\n", s.ID)
			}
		}
	}

	if bidder.Bid > s.BestBid.BidAmount && s.AuctionOngoing == true {
		s.BestBid.BidAmount = bidder.Bid
		s.BestBid.BidderID = bidder.Id

		// Below if-else statement is BOTH for primary and the backup server 
		// This returns a result to the client who tried to bid
		if s.Connection_Helper_Method(5, 5) == false && s.Connection != nil { //looking for connection within 5 seconds
			fileLog.Printf("Server: %d | connection to backup server not active: highest bid \n", s.ID)
			termLog.Printf("Server: %d | connection to backup server not active: highest bid \n", s.ID)
		} else if s.Connection == nil {
			fileLog.Printf("Server: %d | Backup: updating auction values bid = %d, bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)
			termLog.Printf("Server: %d | Backup: updating auction values bid = %d, bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)
			return &pb.BidResponse{Status: "SUCCESS"}, nil
		} else {
			fileLog.Printf("Server: %d | Primary: updating auction values, bid = %d, bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)
			termLog.Printf("Server: %d | Primary: updating auction values, bid = %d, bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)
			res, err := s.BackupConnection.UpdateHighestBid(context.Background(), &pb.Bidder{Bid: bidder.Bid, Id: bidder.Id})
			if err != nil || res.Success != true { // We technically can never return success false as the code is now...
				fileLog.Printf("Server: %d | Failed to update Highest Bid in Backup server \n", s.ID)
				termLog.Printf("Server: %d | Failed to update Highest Bid in Backup server \n", s.ID)
			}
		}
		return &pb.BidResponse{Status: "SUCCESS"}, nil
	}

	return &pb.BidResponse{Status: "FAIL"}, nil
}

/* Returns the highest bid and whether the item has been sold yet or not   */
func (s *AuctionServer) Result(ctx context.Context, empty *pb.Empty) (*pb.ResultResponse, error) {
	s.Mu.Lock()
	defer s.Mu.Unlock()
	return &pb.ResultResponse{
		HighestBid: s.BestBid.BidAmount,
		ItemSold:   !s.AuctionOngoing, // if auction is ongoing then the item hasn't been sold
	}, nil
}

/***************	Primary Server and Backup Server grpc Calls      ***********************/

// Updates the auction state (auction ongoing and the auction round) in the backup server
// returns true
func (s *AuctionServer) UpdateAuctionState(ctx context.Context, state *pb.AuctionState) (*pb.BackupResponse, error) {
	fileLog.Printf("Server: %d | Backup Server: old Auction ongoing: %t new Auction ongoing: %t \n", s.ID, s.AuctionOngoing, state.Ongoing)
	termLog.Printf("Server: %d | Backup Server: old Auction ongoing: %t new Auction ongoing: %t \n", s.ID, s.AuctionOngoing, state.Ongoing)

	s.AuctionOngoing = state.Ongoing
	s.AuctionRound = state.AuctionRound
	return &pb.BackupResponse{Success: true}, nil
}

// Updates regsitered client in the backup server
// returns true
func (s *AuctionServer) UpdateRegisteredClient(ctx context.Context, bidder *pb.Bidder) (*pb.BackupResponse, error) {
	fileLog.Printf("Server: %d | Backup Server: updating registered client \n", s.ID)
	termLog.Printf("Server: %d | Backup Server: updating registered client \n", s.ID)

	s.RegisteredClients[bidder.Id] = true

	fileLog.Printf("Server: %d | Updated map in back up: %v", s.ID, s.RegisteredClients)
	termLog.Printf("Server: %d | Updated map in back up: %v", s.ID, s.RegisteredClients)

	return &pb.BackupResponse{Success: true}, nil
}

// Updates the highest bid in the backup server
// returns true
func (s *AuctionServer) UpdateHighestBid(ctx context.Context, bidder *pb.Bidder) (*pb.BackupResponse, error) {
	s.BestBid.BidAmount = bidder.Bid
	s.BestBid.BidderID = bidder.Id

	fileLog.Printf("Server: %d | Backup: updating auction values: bid = %d and bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)
	termLog.Printf("Server: %d | Backup: updating auction values: bid = %d and bidderId = %d \n", s.ID, s.BestBid.BidderID, s.BestBid.BidAmount)

	return &pb.BackupResponse{Success: true}, nil
}

/********************       Connection Helper Methods        ***********************************/

// Checks the client connection in *AuctionServer is READY for use within a given time frame
// returns true if it is READY, 
// returns false if the connection is not READY or if it doesn't exist
func (s *AuctionServer) Connection_Helper_Method(timeout int, num int) bool {
	// num = 1	initial connection creation
	// num = 2	startAndEndBiddingRound
	// num = 3	local_update_auction_state
	// num = 4 	Bid -> Register client
	// num = 5 	Bid -> update highest bid

	conn := s.Connection
	if conn == nil {
		return false
	}

	// in case Backup Server is not made within "timeout" number of seconds, give up on it and continue
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(int(timeout))*time.Second)
	defer cancel()

	// Save our current state such that we can check if it is READY later
	currentState := conn.GetState()

	for {
		if currentState == connectivity.Ready { // Connection is ready
			break
		}

		// Below line will block until there is a state change or context times out
		if !conn.WaitForStateChange(ctx, currentState) {
			//fileLog.Println("!!!----- BACKUP SERVER CRASHED -----!!!")
			fileLog.Printf("Server: %d | Primary server lost connection to backup \n", s.ID)
			termLog.Printf("Server: %d | Primary server lost connection to backup \n", s.ID)
			return false
		}

		currentState = conn.GetState() // Update our current state

	}
	return true
}