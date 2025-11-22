# Mandatory Assignment 5

## To run the program

You will need at minimum 3 terminals (but you can use more if you want). 2 terminals for the servers and 1 for the client (add more clients to see a bidding war XD)

**For the two Servers:**
Navigate to the Auction folder and call:
`go run ./server 1 true`
`go run ./server 2 false`
The command setup is `go run ./server id IsLeader_bool`, we accept *1, true, True, t* and more for true, and *0, false, FALSE, f* and more for false.

**For the Client:**
In another terminal navigate to the Auction folder, and create as many clients as you want:
`go run ./client 1`
`go run ./client 2`
Remember to give each client a different id number
