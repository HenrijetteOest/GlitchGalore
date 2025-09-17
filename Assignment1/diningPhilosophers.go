package main

import (
	"fmt"
)

/* ***************************************

Explanation:

There are total of three functions, main, Philosopher, and Fork.
A philosopher has four channels, two for each neighbouring fork.
Same for a fork, it also has four channelse, two for each neighbouring philosopher.

Deadlocks are avoided thanks to
if a philosopher is allowed to pick up a fork, 
then he/she asks for the other fork, 
and if he/she is not allowed to pick it, 
then the philosopher puts down the first fork.


Starvation is theoretically posible but statistically improbable
as when the main function continues towards infinity, the probability is evenly distributed.

**********************************************/

func Philosopher(toRightFork chan bool, fromRightFork chan bool, toLeftFork chan bool, FromLeftFork chan bool, name int) {
	fmt.Printf("Philosopher %d sits down at table \n", name)
	counter := 0

	for {
		//True: "Can I have the fork"
		//False: "I'm done"
		toRightFork <- true
		RightForkReply := <-fromRightFork //what does the fork say

		if RightForkReply {
			toLeftFork <- true
			LeftForkReply := <-FromLeftFork //if true picks left fork up, if false puts fork

			if LeftForkReply == true {
				counter++
				fmt.Printf("Philosopher %d is eating (%d) \n", name, counter)
				fmt.Printf("Philosopher %d is done eating and goes back to thinking \n", name)
				toRightFork <- false //puts right fork down
				toLeftFork <- false  // puts left fork down

			} else {
				toRightFork <- false //puts right fork down
			}
		} else {

			fmt.Printf("Philosopher %d is thinking \n", name)

		}
	}
}

func Fork(fromRightPhil chan bool, toRightPhil chan bool, fromLeftPhil chan bool, toLeftPhil chan bool, RPhil int, LPhil int) {
	onTable := true //fork starts on table
	PRrequest := false
	PLrequest := false

	for {

		select {

		case PRrequest = <-fromRightPhil:

			if onTable == false && PRrequest == false {
				onTable = true

			} else if onTable == false && PRrequest == true {
				toRightPhil <- false
			} else if onTable == true && PRrequest == false {
				fmt.Println("something went wrong")
			} else if onTable == true && PRrequest == true {
				onTable = false
				toRightPhil <- true
			}

		case PLrequest = <-fromLeftPhil:

			if onTable == false && PLrequest == false {
				onTable = true

			} else if onTable == false && PLrequest == true {
				toLeftPhil <- false
			} else if onTable == true && PLrequest == false {
				fmt.Println("something went wrong")

			} else if onTable == true && PLrequest == true {
				onTable = false
				toLeftPhil <- true

			}
		}

	}
}

func main() {

	// 20 channels for 5 philosophers and forks.
	P1Rin := make(chan bool)
	P1Rout := make(chan bool)
	P1Lin := make(chan bool)
	P1Lout := make(chan bool)

	P2Rin := make(chan bool)
	P2Rout := make(chan bool)
	P2Lin := make(chan bool)
	P2Lout := make(chan bool)

	P3Rin := make(chan bool)
	P3Rout := make(chan bool)
	P3Lin := make(chan bool)
	P3Lout := make(chan bool)

	P4Rin := make(chan bool)
	P4Rout := make(chan bool)
	P4Lin := make(chan bool)
	P4Lout := make(chan bool)

	P5Rin := make(chan bool)
	P5Rout := make(chan bool)
	P5Lin := make(chan bool)
	P5Lout := make(chan bool)

	//INITIALIZATION OF PHILOSOPHERS AND FORKS

	go Fork(P1Rout, P1Rin, P2Lout, P2Lin, 1, 2) // Fork A
	go Fork(P2Rout, P2Rin, P3Lout, P3Lin, 2, 3) // Fork B
	go Fork(P3Rout, P3Rin, P4Lout, P4Lin, 3, 4) // Fork C
	go Fork(P4Rout, P4Rin, P5Lout, P5Lin, 4, 5) // Fork D
	go Fork(P5Rout, P5Rin, P1Lout, P1Lin, 5, 1) // Fork F

	go Philosopher(P1Rout, P1Rin, P1Lout, P1Lin, 1) // Philosopher 1
	go Philosopher(P2Rout, P2Rin, P2Lout, P2Lin, 2) // Philosopher 2
	go Philosopher(P3Rout, P3Rin, P3Lout, P3Lin, 3) // Philosopher 3
	go Philosopher(P4Rout, P4Rin, P4Lout, P4Lin, 4) // Philosopher 4
	go Philosopher(P5Rout, P5Rin, P5Lout, P5Lin, 5) // Philosopher 5

	for {

	}

}
