package main

import (
	"fmt"
	"time"
)

/*type handshake struct{
	seq, ack int;

}*/

func client(SYN chan int, ACK chan int) {
	var x, y int
	y = 0

	//handshake 1 starts here
	fmt.Println("Handshake 1 start in client")
	x = 100
	SYN <- x

	// 	Handshake 2
	if <-SYN == (x + 1) {
		fmt.Printf("	handshake ended, x = '%d', y = '%d' \n", x, y)
		// Handshake 3
		fmt.Println("Handshake 3 start in client")
		y = <-ACK
		SYN <- (x + 1)
		ACK <- (y + 1)
		// Potentially send data
	}
}

func server(SYN chan int, ACK chan int) {
	var x, y int
	y = 0

	//handshake 1
	x = <-SYN
	fmt.Printf("	handshake ended, x = '%d', y = '%d' \n", x, y)
	// handshake 2
	fmt.Println("Handshake 2 start in server")
	y = 300
	SYN <- (x + 1)
	ACK <- y

	// handshake 3
	fmt.Printf("	seq should be 101 and is '%d', ack should be 301 and is '%d' \n", <-SYN, <-ACK)
}

func main() {
	SYN := make(chan int)
	ACK := make(chan int)

	go client(SYN, ACK)
	go server(SYN, ACK)

	time.Sleep(2 * time.Second)
	//handshake 1 = SYN seq=x (100)
	//handshake 2 = SYN-ACK  ack = x+1(100 +1 ) seq= y(300)
	//handshake 3 = ACK  ack=y+1(300+1) seq=x+1 (100+1)
}
