package main;

import (
	"context"
	"fmt"
	"log"
	"net"	
	"time"
	"google.golang.org/grpc"

	pb "Auction/grpc"
)
type BackupAuctionServer struct {
	pb.UnimplementedAuctionServiceServer
	ID       int32		
	LeaderID int32
	BestBid           HighestBidder  
	RegisteredClients map[int32]bool 
	AuctionOngoing    bool
}
type HighestBidder struct {
	BidderID  int32
	BidAmount int32
}
func main() {
	fmt.Println("Starting Backup Auction Server...")
	server := &BackupAuctionServer{	
		ID:                2,
		LeaderID:          1,
		BestBid:           HighestBidder{BidderID: -1, BidAmount: 0},	
		RegisteredClients: make(map[int32]bool),
		AuctionOngoing:    false,
	}
	server.start_backup_server()		
	select {}
}
// Starts the backup server (serve() is a goroutine )
func (s *BackupAuctionServer) start_backup_server() {
	lis, err := net.Listen("tcp", ":6060")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterAuctionServiceServer(grpcServer, s)
	fmt.Printf("Backup Auction Server %d listening on port 6060\n", s.ID)
	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()
}

