# How to run the code

Run the program from 3 separate terminals from the Node directory

first terminal: `go run node.go 1 5000 5001 5002`
second terminal: `go run node.go 2 5000 5001 5002`
third terminal: `go run node.go 3 5000 5001 5002`


## Explanation of command line setup

The first number "1" is the index of the node's port nr, which refers to the index position of its portnumber. The four digit numbers after refer to the portnumber for a node.
`go run node.go 1 5000 5001 5002`
`go run node.go 2 5000 5001 5002`
`go run node.go 3 5000 5001 5002`

The node nr is not used for more than indexing which portnumber a node should use.



