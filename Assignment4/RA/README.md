# How to run the code

Run the program from the RA folder:
`go run ./node 1 5000`

Run the program from the node folder:
`go run node.go 1 5000`

## Explanation of command line setup

The first number "1" is the node nr, which refers to the index position of its portnumber. The four digit numbers after refer to the portnumber for a node.
`go run node.go 1 5000 5001 5002`
`go run node.go 2 5000 5001 5002`
`go run node.go 3 5000 5001 5002`

The node nr is not used for more than indexing which portnumber a node should use.


## Potential errors:
Not having made all instances of a node before client connections are established (?) (like start two nodes, and wait too long with the third note.)


