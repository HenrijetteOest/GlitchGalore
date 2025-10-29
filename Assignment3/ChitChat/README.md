
# To run the program

Start by running the `server.go` file from the ChitChat folder with the following command:
`go run ./server`

And then create a client in another terminal (also from the ChitChat folder):
`go run ./client`

## Special mentions

### Simulating a client

Our client code includes a function `SendMessageLoop`, that we use to simulate a user sending messsages. It has a loop that repeats 20 times with a message (the message is up to 5 random words from Bram Stoker's *Dracula*, see more in section "Client Message *0 0 Dracula 0 0*").

### Random Client Names

To give *ChitChat* a more realistic feel of users interacting, we use the *random-names* package, from <https://github.com/random-names/go>.
A dataset of female names `all.last` has also been included to use with the package, from <https://github.com/random-names/names/tree/master>

### Client Message *0 0 Dracula 0 0*

To generate messages for client to send we take a random word from Bram Stoker's *Dracula*. The method to get a random word, which is the same as for getting a random name. Note, if the package finds a space symbol " " it replaces it with a *0* in the text. The txt version of *Dracula*  was found here <https://github.com/newtfire/introDH-Hub/blob/master/textFiles/19c-fiction/stoker-dracula.txt>
