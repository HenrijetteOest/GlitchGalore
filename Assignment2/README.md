# README answers to mandatory 2

## What are packages in your implementation? What data structure do you use to transmit data and meta-data?

Our program has ints as our package type. We use channels as the data structure to send these packages between threads. 


## Does your implementation use threads or processes? Why is it not realistic to use threads?

Our implementation answers task (1)\[easy\] with the non-realistic version of the 3-way-handshake, consisting of threads instead of processes. Threads share the same memory space as their parent process, meaning it is impossible for two threads in different processes to be the exact same thread. The TCP stereotypically involve communication between different hosts over the Internet, making the 3-way-handshake incapable of using the same threads. Should one of the three threads (client, server, and main) in our implementation crash, they risk bringing down the whole process (with deadlocks, starvation, or other). Had we used two separate processes, the crashing of *client* or *server*  would not terminate the other process, instead ending only the handshake.


##  In case the network changes the order in which messages are delivered, how would you handle message re-ordering?

TCP ensures message order reliability by given each byte of sent data a unique sequence number. Sequence numbers are both used in the initial establishing of the 3-way-handshake connection, and in the following transferring of data between *client* and *server*. 

We would add a way for the server and client to track the sequence of the data segments, so that the server can reassemble them in the correct order. 


## In case messages can be delayed or lost, how does your implementation handle message loss?

We would add functionality to our client code so that it can keep track of the segements it sends to the server and the acknowledgements it receives from the server. If the client doesn't recieve an acknowledgement within a certain time frame, it would re-send the segment to the server. 


## Why is the 3-way handshake important?

The 3-way handshake is needed to ensure that the connection between the client and server have been established properly. A proper connection means that both sides are ready to exchange data, minimizing the risk for errors or data loss. 

