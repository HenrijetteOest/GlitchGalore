# Notes

## Message
string: text 0-128 chars 

## User:
string: username (for fun, not necessary)
int: id

## Client:
- Needs to be able to take input from a terminal (goroutine der læser consecutively fra terminalen eller hardcode det)
  - To save Name
  - To continuosly get "messages" to send to the server.
- Needs a one-way stream to the Server, which is used whenever SendMessage() is called.
- (A method for getting a unique id)


## Server:
- Needs a one-way stream from Server to *all* connected clients (that is clients in the chat). This stream is used whenever broadcasting() is called.

## Proto:
- proto should not be empty when joinChat() is called (?)

## Streams:
1) client-to-server: sendMessage() ("Server here is what I want to share with the world")
2) Server-to-client: broadcast() ("Listen all chat participants, this is a message for *all* of you")


## Useful Links

https://dev.to/yash_mahakal/implementing-bidirectional-grpc-streaming-a-practical-guide-3afi

Different streaming versions
https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc

https://github.com/grpc/grpc-go/blob/master/examples/route_guide/routeguide/route_guide.proto 

















# Report Notes:

## discuss, whether you are going to use server-side streaming, client-side streaming, or bidirectional streaming?


## describe your system architecture - do you have a server-client architecture, peer-to-peer, or something else?



## describe what RPC methods are implemented, of what type, and what messages types are used for communicationdescribe how you have implemented the calculation of the timestamps


## provide a diagram, that traces a sequence of RPC calls together with the Lamport timestamps, that corresponds to a chosen sequence of interactions: Client X joins, Client X Publishes, ..., Client X leaves. 