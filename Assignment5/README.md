# Mandatory Assignment 5

## To run the program

You will need at minimum 3 terminals (but you can use more if you want). 2 terminals for the servers and 1 for the client (add more clients to see a bidding war XD)

**For the two Servers:**
Navigate to the Auction folder and call:
`go run ./server 1 true`
`go run ./server 2 false`
The command setup is `go run ./server id isLeader_bool`, we accept *1, true, True, t* and more for true, and *0, false, FALSE, f* and more for false.

**For the Client:**
In another terminal navigate to the Auction folder, and create as many clients as you want:
`go run ./client 1`
`go run ./client 2`
Remember to give each client a different id number.

## TODO

The current to do list:

1) add logic for the last two grpc methods to Server (the calls between the leader and backup server)

2) update the bid() grpc call to update the backup server as well.

3) Client needs to wait for a bid response before placing the next bid (if it doesn't already do this)

4) Create and handle crashes    (we are partially resistant to the backup server crashing (but not the leader))
    - logic to crash leader
    - logic for Client to switch to Backup (or for backup to take)

5) Check linearisability is upheld (we might need lamports...)

Nice to have:
6) a log file
7) Client logic to handle poor initial connection (such that a client doesn't crash if it is made before the server)
8) Let the HighestBidder struct in Server be a proto message type instead
