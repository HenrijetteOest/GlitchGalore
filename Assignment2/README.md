# README answers to mandatory 2

## What are packages in your implementation? What data structure do you use to transmit data and meta-data?
meta data : could be time, location, IP-address, 
    e.g mads: tidspunktet x = noget helt konkret fx 100 eller 101
    
data: int, vi sender frem og tilbage

## Does your implementation use threads or processes? Why is it not realistic to use threads?
threads i client og server 
løst opgave 1 "easy"

##  In case the network changes the order in which messages are delivered, how would you handle message re-ordering?
we use dont use that, we would use time outs and window sliding

## In case messages can be delayed or lost, how does your implementation handle message loss?

pas for nu, it does not, not in easy, but deadlocks....

## Why is the 3-way handshake important?

security.......... UDP er hurtig men er ikke stabil/sikker, TCP som vi har lavet er stabil og giver garanti for at alt bliver handlet(lavet) korrekt

