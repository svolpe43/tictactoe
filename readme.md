  # Tic Tac Toe

There are two main Go programs included in this repo. A frontend and a backend. The backend code is in the `server` directory and is running on an EC2 instance on shawnvolpe.com. The frontend code you can run locally to play Tic Tac Toe via the CLI.

### Running the client

After installing the repo in your local Golang environment, run the following command to start the client. Once you see the prompt `->` you are ready to start playing TicTacToe.

Should look like this...
```
cd tictactoe
shawn@machine:~/go/src/github.com/svolpe43/ttt$ go run ./frontend
Welcome to Tic Tak Toe!
->
```

## Commands

### List 
`list`

Lists available hosted tic tac toe games to join.

### Create a game
`create <name> <choice of symbol>`

Creates a new tic tac toe game and connects the client to that game. There are two parameters.
`name` - the human readable name of the game.
`choice of symbol` - the symbol you would like to be. Valid options are `X` and `O`

Example: `create joe-shawn-game X`

### Join a game
`join <name>`

Connects the client to an existing tic tac toe game. Your symbol will automatically be selected based on what is not already taken. This will then render the board and play with your opponent.

Example: `join joe-shawn-game`

### Make a move
`move <index>`

Once a game is joined this command makes a move. You will only be able to make a move when it is your turn. There is one parameter.

`index` - Index is an integer from 0 to 8 and represents the square you would like to occupy. The cells are numbered left to right and top to bottom.

Example: `move 5`

### End a game
`end <name>`

This command ends a game and disconnects the client. The allows the client to either select a new game to join or to create a new game. There is one parameter.

`name` - The name of the game to end.

Example: `end joe-shawn-game`

