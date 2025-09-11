
package main

import (
	"fmt"
	"time" 
)

/* ***************************************

Notes

philosopher asks left fork first
left fork checks if it is free, changes itself if necessary, and then responds yes or no
   - the fork only changes if it's free and the philosopher wants it
   - or the philosopher is done using it and returns it (where it turns from false to true (back on table))
The fork doesn't change upon pick-up request if it's not-free 
Process repeats on the right side (which is exactly the same as the left side)
    if this fork is not free, put the left fork down again (to avoid deadlocks)
	if it is free then the philosopher eats until some time (sleep method) then she puts the forks down again 

Total of 20 channels in the question.

A philosopher has 4 channels, two for each neighbouring fork.
Same for a fork, it also has 4 channelse, two for each neighbouring philosopher

**********************************************/



// channels for 3 philosophers and forks.
var P1R := make(chan bool)
var P1L := make(chan bool)
var P2R := make(chan bool)
var P2L := make(chan bool)
var P3R := make(chan bool)
var P3L := make(chan bool)
var P4R := make(chan bool)
var P4L := make(chan bool)
var P5R := make(chan bool)
var P5L := make(chan bool)

func Philosopher(){
	//table.Lock()  // table is out
	//philosopher tries to pick up forks via channels 
	var iHaveR = false
	var iHaveL = false

	rFork := someChannel1
	if (rFork) {
		iHaveR = true
		someChannel1 = <- false
	}

	lFork := someChannel2
	if (rFork) {
		iHaveL = true
		someChannel2 = <- false
	}

	if (rFork && !lFork) {
	
	}

	if (rFork && lFork) {	// you pick them both up
		fmt.Println(Philosopher + " Eating")
		someChannel1 <- false
		someChannel2 <- false
		
	} 

	//table.Unlock()
	

}

func Fork(chan RightPhil, chan LeftPhil) {
	var onTable = true
	var takeMe =
	var placeMe = 
}


func main() {
	go Philosopher()
	go Fork()

	for{		
	}
}
 
