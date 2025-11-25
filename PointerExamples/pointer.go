package main

import (
	"fmt"
)


func main() {
	var i int = 0
	var j = &i

	fmt.Println("i is:", i)
	fmt.Println("i's adress is:", &i)
	fmt.Println("j is esseentially i's adress:", j)
	fmt.Println("j's adress is:", &j) //which is keeping the adress to i (&i)
	fmt.Println("everyting inside j is:", *j)
	fmt.Println("everyting inside j is still:", **&j)

	var p = &j
	fmt.Println("for fun, what is the answer you think?:", p)
	fmt.Println("for fun, what is the answer you think?:", &p)
	fmt.Println("for fun, what is the answer you think?:", *&p)
	fmt.Println("for fun, what is the answer you think?:", **&p)
	fmt.Println("for fun, what is the answer you think?:", ***&p)
}