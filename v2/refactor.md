# Refactor

The original plan was to make every part of the pieces independently game,
server, bot, machine and I chose the best data type for the each piece,
but I like the machine data the best. The simplest way possible to describe
the data. I heard that before mostly from functinal programming. So refactoring
the whole thing and baseing it on a commen datatype that is as simple as
possible make sense.

## Game

All pieces in battleline is destinct so the whole game can be described with the
postion of those pieces and who moves next. The rules describe who can move zero
or more pieces and where the can be moved, a game move. The number of possible
moves is not big so it makes sense to calculate all the possible moves there by
avoiding to implement them more places and easy way to make sure only legal
moves is made.

We know the start position so every game can be fully described by a list of legal
moves and any game position can be calculed from the init position by moving the
pieces accordingly to the moves. When a game is in progress we only care about
the current position not the history of a game. When we save a game we care about the
history. We need to split the two definition to make saving easier.

GameHistory: Moves []GameMoves, PlayerIds []int,TimeStamp Time
Game: Pos GamePos,gameHist GameHistory

Time type is coverted to RFC 3339Nano when marshal:
Js can convert with [Date.Parse(txt)](https://stackoverflow.com/questions/11318634/how-to-convert-date-in-rfc-3339-to-the-javascript-date-objectmilliseconds-since)

The key function adds ts.Nano/1000.000 to the key as 64bit BigEndian.Nano is
1.000.000.000 so that is 2 bytes to big. One byte is only 256 which gives 100
per second. I don't like that a factor of 10 faster, maybe the new struct
is slower :).
We add diggits manualy but could use the format function with .000 after
seconds. RFC3339Nano use 9999999 which is the reason that it can't sort 9 is
with out trailing zeros.


## GameMove

A game move is a mover and a list of board piece moves.

GameMove: Mover int,MoveType int, Moves []BoardPieceMove

To calculate the next possible moves and mover we need a move type
as we can not see the different between scout return 2 or 3.
The type could be in postion only but we know it when we calculate
the moves so adding them to the move make sense.
Type: Init,Cone,Deck,Hand,Scout1-3,Give-Up,Pause

## BoardPieceMove

The board piece is either a Card or a Cone there is 70 cards and 9 cones. The card
have 23 possible position and the cone 3. So describe the move you need three fields
and four if we want to go both ways, adding and substracting a move from a game
position. 
 
A move may not be fully revealed depending on the viewer and time. For example pick
a card from the deck. The card is not known by the player before the move have been
made but what is known is the deck. Scout return posses similar problems. To handle
that we need to more cardixs index 98 for troop deck and 99 for tactic deck.

* **BoardPiece** Card or Cone.
* **Index** 1-72, for card and for cone 1-9.
* **NewPos** 0-22 for cards and 0-2 it make sense to have zero the start position.
* **OldPos** by adding old position we can go both ways adding and substracting a move.

Flag and dish is 20 positions hand is 2 and deck 1 that 23 position. Maybe deck should be two position
Tac and Troop? It will help with display and deal and it would not matter much for machine data as
deck troop and deck tac is only mixed on scout return positions, normal card postions just
substituted.

Two other moves is possible **give-up** and **pause** the game. 
* **GiveUp**  moves all non claimed flags to opponent.
* **Pause** a new type and change only the last mover. When resume next mover calculation should keep should delete the move from the list of moves and set the nextmover.

## Game position

The game position is all the pieces and their positions and who move next.

* **CardPos**: [70]int adding one and so cardix is easy recognised make sense [71]
* **ConePos**: [9]int
* **LastMover**: int 0,1 for player and -1 for none.
* **LastMoveType**: Init,Cone,Deck,Hand,Scout1-3,Give-Up,Pause
* **LastMoveIx**: The move index in history.

It may be more usefull to reverse the card to position map as game logic and display 
depend on all cards in one position.

**CardPos**: [23][]int to replace **Cards**: [70]int

The move gets a little more comblicated as the the card must be removed from the old
position and added to the new position. Two opperation and removing a value from a slice is slow,
first find the index and move all the rest one postion. The two is not equal in the information they
can hold as a list can hold order.

The order of the deck can be used to suffle and then draw a card move can be done with
out search and select. This will make two deck positions optimal. Load game from file
will still cost extra time. It also make scout return very logic. However it must 
be resuffle when send on to the players.

The cards is more or lesh a deck swap the "map" with position instead boolean and it
is the same and we need that when loading a game.

All card position is realy only needed with traitor move and only for
calculating moves. Lets try without **CardPos**introducing it as a cache if needed.

The order of cards in the gui(cards group) should be part of the gui (react state) not
the logic state(prop) this is a problem as order of arrival can influence the order.
Making it necessary to maintain it own lists of positions of cards. We know those lists
will only change when cards change or the hand we could speed up things by implementing
shouldComponentUpdate(nextProps, nextState)

Scount returned cards is knowledge about card postions by one player. We need extra 
fields for that knowledge.

* **PlayerReturned**: int 0,1 and 2 for none.
* **FirstCardReturned**: Cardix int
* **SecondCardReturned**: Cardix int

## Game position view

There can be different views on the game position player and spectator.
To handle that we need a few extra fields and to modify the Game position.

* **GamePos**: GamePos modified to hide hand card and scout return card.
* **View**: player index 0,1 and 2 for spectator and 3 for God.
* **NoTac[player]**: Number of unknown tractic cards on player hand.
* **NoTroop[player]**: Number of unknown troop cards on player hand. 

The game position must be changed depending on the view. 
The opponent cards hand must be changed to Deck unless it is in
scout returning. The opponent scout return cards must be change to troop or
tactic deck when the card is in deck if it is on player hand it should not be 
changed any other place it should be change to 0.
When the player returned scout no change is need.

## Rand

Right now the same rand source is used by all goroutines login, starter and create deck and every time they are used
I recreate the source that is not a good idea it should only be done once the program starts. There is also
crypto/rand that should be used for security in session id client.go. 256 bit created with from 32 byte.
