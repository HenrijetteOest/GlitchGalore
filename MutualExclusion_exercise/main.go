package main


import(
	"fmt"
	"sync"
	"time"
)

var mu sync.Mutex; 

// Critical section which should only be accessed by one goroutine at a time 
func setLeader(id int32) {
	fmt.Printf("Muahahaaa, I, %d, am now the leader XD \n", id)
}

// a function to call setLeader at some point
func citizen(id int32) {
	for i := 0; i < 10; i++ {
		mu.Lock()
		setLeader(id)
		mu.Unlock()
		time.Sleep(100 * time.Millisecond)
	}
}

// main method
func main(){
	

	go citizen(1)
	go citizen(2)
	

	time.Sleep(20 * time.Second)
}