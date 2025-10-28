
# To run the program

Start by running the `server.go` file from the ChitChat folder with the following command:
`go run ./server`

And then create a client in another terminal (also from the ChitChat folder):
`go run ./client`

## Special mentions: Random Client Names

To give *ChitChat* a more realistic feel of users interacting, we use the *random-names* package, from <https://github.com/random-names/go>.
A dataset of female names `all.last` has also been included to use with the package, from <https://github.com/random-names/names/tree/master>
