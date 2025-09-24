package main

import (
	"fmt"
	"time"
)

func client(SYN chan int, ACK chan int) {
	var x, y int

	//handshake 1 starts here
	fmt.Println("Handshake 1 start in client")
	x = 100
	SYN <- x

	// 	Handshake 2
	if <-SYN == (x + 1) {
		y = <-ACK
		fmt.Printf("	handshake ended, x = '%d', y = '%d' \n", x, y)
		// Handshake 3
		fmt.Println("Handshake 3 start in client")
		ACK <- (y + 1)
		SYN <- (x + 1)
		// Potentially send data
	} else {
		fmt.Println("Wrong return value from server")
	}
}

func server(SYN chan int, ACK chan int) {
	var x, y, tmp int
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
	tmp = <-ACK
	if tmp == (y + 1) {
		fmt.Printf("	seq should be 101 and is '%d', ack should be 301 and is '%d' \n", <-SYN, tmp)
		fmt.Println("	connection established")
	} else {
		fmt.Println("Wrong return value from client")
	}
}

func main() {
	SYN := make(chan int)
	ACK := make(chan int)

	go client(SYN, ACK)
	go server(SYN, ACK)

	time.Sleep(2 * time.Second)

}
